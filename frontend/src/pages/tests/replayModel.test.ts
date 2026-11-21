import {expect, it} from 'vitest';
import {TestResultState} from '../../types/test.result.state.enum';
import {
    buildTimeline, dataUrl, eventIndexAt, eventIsError, formatOffset, frameIndexAt, parseMjpegIndex, replayRange,
    ReplayEvent, SeekClock, segmentAt, segmentsOf, videoSeekTarget,
} from './replayModel';
import {at, entry, logEntry, protocol, recording, run, T0} from './replay/testFixtures';

const ev = (time: number, over: Partial<ReplayEvent> = {}): ReplayEvent => ({key: `k${time}`, time, kind: 'log', level: 'info', text: '', ...over});

it('segmentsOf: lists finished recordings per device, oldest first, with data urls', () => {
    const map = segmentsOf([
        recording({id: 2, startedAt: at(5000).toISOString(), endedAt: at(9000).toISOString()}),
        recording({id: 1, indexFile: 'i x.idx', codec: 'mjpeg'}),
        recording({id: 3, deviceId: 6}),
    ]);
    expect([...map.keys()]).toEqual([5, 6]);
    expect(map.get(5)?.map(s => s.id)).toEqual([1, 2]);
    expect(map.get(5)?.[0]).toMatchObject({url: '/api/data/a%20b.mp4', indexUrl: '/api/data/i%20x.idx', start: T0, end: T0 + 10000});
    expect(map.get(5)?.[1].indexUrl).toBeNull();
});

it.each([
    ['without an end', {endedAt: null}],
    ['ending before the start', {endedAt: at(0).toISOString()}],
    ['with an invalid start', {startedAt: 'nope'}],
])('skips a recording %s', (_, over) => {
    expect(segmentsOf([recording(over)]).size).toBe(0);
});

it('segmentsOf: handles no recordings', () => {
    expect(segmentsOf(undefined).size).toBe(0);
});

const segments = segmentsOf([recording()]).get(5) ?? [];

it('segmentAt: finds the segment showing the time, the end excluded', () => {
    expect(segmentAt(segments, T0)?.id).toBe(1);
    expect(segmentAt(segments, T0 + 10000)).toBeNull();
    expect(segmentAt(segments, T0 - 1)).toBeNull();
});

it('buildTimeline: merges log, test boundaries and entries in time order', () => {
    const events = buildTimeline(run(), 5);
    expect(events.map(e => [e.key, e.kind, e.time - T0])).toEqual([
        ['l1', 'log', 500], ['s11', 'test-start', 1000], ['p21', 'step', 2000], ['p22', 'screen', 3000],
        ['e11', 'test-end', 6000], ['l2', 'log', 8000],
    ]);
});

it('buildTimeline: labels tests with name and suite, and marks a failed end as error', () => {
    const events = buildTimeline(run(), 5);
    expect(events[1]).toMatchObject({text: 'Login (Ns.Cls)', protocolId: 11, level: 'info'});
    expect(events[4]).toMatchObject({level: 'error', result: TestResultState.TestResultFailed});
});

it('buildTimeline: gives a screenshot its image and a default text', () => {
    const shot = buildTimeline(run(), 5)[3];
    expect(shot).toMatchObject({image: '/api/data/shot%20one.png', text: 'Screenshot', source: 'screen'});
});

it('buildTimeline: treats a screen entry without data as a log line', () => {
    const p = protocol({entries: [entry({id: 1, source: 'screen', data: '', message: ''})]});
    expect(buildTimeline(run({protocols: [p]}), 5).find(e => e.key === 'p1')).toMatchObject({kind: 'log', text: '', image: undefined});
});

it('buildTimeline: keeps only the tests of the device, those without device included', () => {
    const other = protocol({id: 12, deviceId: 6, testName: 'Other'});
    const free = protocol({id: 13, deviceId: null, testName: 'Free', endedAt: undefined, testResult: undefined});
    const protocols = [protocol(), other, free];
    expect(buildTimeline(run({protocols}), 5).filter(e => e.kind === 'test-start').map(e => e.text)).toEqual(['Login (Ns.Cls)', 'Free']);
    expect(buildTimeline(run({protocols}), null).filter(e => e.kind === 'test-start')).toHaveLength(3);
});

it('buildTimeline: skips items without a usable time and falls back to createdAt', () => {
    const log = [logEntry({id: 1, log: 'x', time: 'bad' as unknown as Date})];
    const p = protocol({
        startedAt: 'bad' as unknown as Date, endedAt: undefined,
        entries: [entry({id: 1, timestamp: undefined, createdAt: at(4000)}), entry({id: 2, timestamp: undefined, createdAt: 'bad' as unknown as Date})],
    });
    const events = buildTimeline(run({log, protocols: [p]}), 5);
    expect(events.map(e => e.key)).toEqual(['p1']);
    expect(events[0].time).toBe(T0 + 4000);
});

it('buildTimeline: handles a run without log and protocols and keeps equal times in order', () => {
    expect(buildTimeline(run({log: undefined, protocols: undefined}), null)).toEqual([]);
    const same = [1, 2].map(id => logEntry({id, log: `n${id}`, time: 100}));
    expect(buildTimeline(run({log: same, protocols: []}), null).map(e => e.key)).toEqual(['l1', 'l2']);
});

it.each([
    [ev(0, {level: 'error'}), true],
    [ev(0, {level: 'exception'}), true],
    [ev(0, {level: 'info'}), false],
    [ev(0, {kind: 'test-end', result: TestResultState.TestResultFailed}), true],
    [ev(0, {kind: 'test-end', result: TestResultState.TestResultSuccess}), false],
])('classifies %#', (event, expected) => {
    expect(eventIsError(event)).toBe(expected);
});

const events = [ev(10), ev(20), ev(30)];

it.each([[5, -1], [10, 0], [25, 1], [30, 2], [99, 2]])('time lookups: eventIndexAt(%d) is %d', (time, index) => {
    expect(eventIndexAt(events, time)).toBe(index);
    expect(frameIndexAt(events.map(e => ({time: e.time, offset: 0, length: 0})), time)).toBe(index);
});

it('time lookups: replayRange spans segments and events, null when empty', () => {
    const range = segmentsOf([recording()]).get(5) ?? [];
    expect(replayRange(range, [ev(T0 + 12000)])).toEqual({start: T0, end: T0 + 12000});
    expect(replayRange([], events)).toEqual({start: 10, end: 30});
    expect(replayRange([], [])).toBeNull();
});

it('mjpeg index: reads "ms offset length" lines relative to the segment start', () => {
    expect(parseMjpegIndex('0 0 100\n 40 100 50 \nbad line\n1 2\nx 1 2\n', 1000))
        .toEqual([{time: 1000, offset: 0, length: 100}, {time: 1040, offset: 100, length: 50}]);
});

it.each([[0, '0:00.0'], [-5, '0:00.0'], [61234, '1:01.2'], [599999, '9:59.9']])('formatOffset: %d ms is %s', (value, text) => {
    expect(formatOffset(value)).toBe(text);
});

const clock = (playing: boolean): SeekClock => ({position: T0 + 5000, segmentStart: T0, playing});

it('videoSeekTarget: seeks when the video is off by more than the tolerance', () => {
    expect(videoSeekTarget(5.3, clock(false))).toBe(5);
    expect(videoSeekTarget(5.3, clock(true))).toBeNull();
    expect(videoSeekTarget(6, clock(true))).toBe(5);
    expect(videoSeekTarget(5.01, clock(false))).toBeNull();
});

it('videoSeekTarget: never targets before the file start', () => {
    expect(videoSeekTarget(3, {position: T0 - 500, segmentStart: T0, playing: false})).toBe(0);
});

it('dataUrl: encodes the file name', () => {
    expect(dataUrl('a/b c')).toBe('/api/data/a%2Fb%20c');
});
