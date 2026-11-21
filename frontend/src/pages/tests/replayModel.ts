// The replay of a recorded run without React: the video segments of a device,
// one timeline of the run log, the tests and their steps, and lookups by time.
// All times are milliseconds since the epoch, on the master's clock, which
// both the recordings and the logs use.
import ITestRunData from '../../types/test.run';
import ITestRunRecordingData from '../../types/test.run.recording';
import ITestProtocolData from '../../types/test.protocol';
import IProtocolEntryData from '../../types/protocol.entry';
import {TestResultState} from '../../types/test.result.state.enum';
import {splitTestName} from './testRunModel';

export interface ReplaySegment {
    id: number,
    deviceId: number,
    codec: 'mp4' | 'mjpeg',
    url: string,
    indexUrl: string | null,
    start: number,
    end: number,
    width: number,
    height: number,
}

export const dataUrl = (name: string): string => `/api/data/${encodeURIComponent(name)}`;

const ms = (d: Date | string | null | undefined): number | null => {
    if (d === null || d === undefined) return null;
    const v = new Date(d).valueOf();
    return isNaN(v) ? null : v;
};

// segmentsOf lists the finished recordings by device, oldest first; one
// that never got its end (the master stopped) can't be played.
export const segmentsOf = (recordings: ITestRunRecordingData[] | undefined): Map<number, ReplaySegment[]> => {
    const byDevice = new Map<number, ReplaySegment[]>();
    (recordings ?? []).forEach(r => {
        const start = ms(r.startedAt);
        const end = ms(r.endedAt);
        if (start === null || end === null || end <= start) return;
        const list = byDevice.get(r.deviceId) ?? [];
        list.push({
            id: r.id, deviceId: r.deviceId, codec: r.codec, url: dataUrl(r.fileName),
            indexUrl: r.indexFile ? dataUrl(r.indexFile) : null, start, end, width: r.width, height: r.height,
        });
        byDevice.set(r.deviceId, list);
    });
    byDevice.forEach(list => list.sort((a, b) => a.start - b.start));
    return byDevice;
};

// segmentAt is the segment that shows the time, if any
export const segmentAt = (segments: ReplaySegment[], time: number): ReplaySegment | null =>
    segments.find(s => time >= s.start && time < s.end) ?? null;

export type EventKind = 'log' | 'step' | 'test-start' | 'test-end' | 'screen';

export interface ReplayEvent {
    key: string,
    time: number,
    kind: EventKind,
    level: string,
    text: string,
    // the source of a step or log line of a test ("" is the runner)
    source?: string,
    protocolId?: number,
    // the screenshot of a "screen" entry
    image?: string,
    result?: TestResultState,
}

const isError = (level: string): boolean => level === 'error' || level === 'exception';

export const eventIsError = (e: ReplayEvent): boolean =>
    isError(e.level) || (e.kind === 'test-end' && e.result === TestResultState.TestResultFailed);

const testLabel = (p: ITestProtocolData): string => {
    const {name, suite} = splitTestName(p.testName);
    return suite ? `${name} (${suite})` : name;
};

const logEvents = (run: ITestRunData): ReplayEvent[] => (run.log ?? []).flatMap(l => {
    const time = ms(l.createdAt);
    return time === null ? [] : [{key: `l${l.id}`, time, kind: 'log' as const, level: l.level, text: l.log}];
});

const testBoundaries = (p: ITestProtocolData, id: number): ReplayEvent[] => {
    const text = testLabel(p);
    const start = ms(p.startedAt);
    const end = ms(p.endedAt);
    const result = p.testResult ?? TestResultState.TestResultOpen;
    const failed = result === TestResultState.TestResultFailed;
    return [
        ...(start === null ? [] : [{key: `s${id}`, time: start, kind: 'test-start' as const, level: 'info', text, protocolId: id}]),
        ...(end === null ? [] : [{key: `e${id}`, time: end, kind: 'test-end' as const, level: failed ? 'error' : 'info', text, protocolId: id, result}]),
    ];
};

const entryKind = (e: IProtocolEntryData): EventKind => {
    if (e.source === 'screen' && !!e.data) return 'screen';
    return e.source === 'step' ? 'step' : 'log';
};

const entryEvents = (p: ITestProtocolData, id: number): ReplayEvent[] => (p.entries ?? []).flatMap((e: IProtocolEntryData) => {
    const time = ms(e.timestamp) ?? ms(e.createdAt);
    if (time === null) return [];
    const kind = entryKind(e);
    return [{
        key: `p${e.id}`, time, kind, level: e.level, text: e.message || (kind === 'screen' ? 'Screenshot' : ''),
        source: e.source, protocolId: id, image: kind === 'screen' ? dataUrl(e.data) : undefined,
    }];
});

// buildTimeline merges the run log (all devices) with the tests of the device
// and their entries, in the order they happened.
export const buildTimeline = (run: ITestRunData, deviceId: number | null): ReplayEvent[] => {
    const protocols = (run.protocols ?? []).filter(p => deviceId === null || !p.deviceId || p.deviceId === deviceId);
    const events = [
        ...logEvents(run),
        ...protocols.flatMap(p => [...testBoundaries(p, p.id as number), ...entryEvents(p, p.id as number)]),
    ];
    // stable: events of the same time keep the order they were added in
    return events.map((e, i) => ({e, i})).sort((a, b) => a.e.time - b.e.time || a.i - b.i).map(x => x.e);
};

// lastAtOrBefore is the index of the last item at or before the time (-1
// when none); items are sorted by time
const lastAtOrBefore = (items: { time: number }[], time: number): number => {
    let lo = 0, hi = items.length - 1, found = -1;
    while (lo <= hi) {
        const mid = (lo + hi) >> 1;
        if (items[mid].time <= time) {
            found = mid;
            lo = mid + 1;
        } else {
            hi = mid - 1;
        }
    }
    return found;
};

// eventIndexAt is the index of the last event at or before the time (-1 when
// none); events are sorted by time
export const eventIndexAt = (events: ReplayEvent[], time: number): number => lastAtOrBefore(events, time);

export interface ReplayRange {
    start: number,
    end: number,
}

// replayRange is the time span the replay covers: the recordings and the
// events, whichever reach further
export const replayRange = (segments: ReplaySegment[], events: ReplayEvent[]): ReplayRange | null => {
    const times = [...segments.flatMap(s => [s.start, s.end]), ...events.map(e => e.time)];
    if (times.length === 0) return null;
    return {start: Math.min(...times), end: Math.max(...times)};
};

export interface MjpegFrame {
    time: number,
    offset: number,
    length: number,
}

// parseMjpegIndex reads the index of an MJPEG recording: "ms offset length"
// per line, ms since the segment start
export const parseMjpegIndex = (text: string, start: number): MjpegFrame[] =>
    text.split('\n').map(l => l.trim().split(/\s+/).map(Number))
        .filter(f => f.length === 3 && f.every(n => Number.isFinite(n)))
        .map(([t, offset, length]) => ({time: start + t, offset, length}));

// frameIndexAt is the frame on screen at the time (the last one at or before
// it), -1 before the first
export const frameIndexAt = (frames: MjpegFrame[], time: number): number => lastAtOrBefore(frames, time);

// formatOffset shows a time relative to the start of the replay, m:ss.t
export const formatOffset = (msSinceStart: number): string => {
    const total = Math.max(0, msSinceStart);
    const m = Math.floor(total / 60000);
    const s = Math.floor((total % 60000) / 1000);
    const t = Math.floor((total % 1000) / 100);
    return `${m}:${String(s).padStart(2, '0')}.${t}`;
};

const DRIFT_PLAYING_S = 0.5;
const DRIFT_PAUSED_S = 0.04;

export interface SeekClock {
    // ms since the epoch, and the start of the segment the video shows
    position: number,
    segmentStart: number,
    playing: boolean,
}

// videoSeekTarget is the time (seconds into the file) a video element has to
// jump to, or null while it is close enough; playing tolerates more drift so
// it stays smooth
export const videoSeekTarget = (currentTime: number, c: SeekClock): number | null => {
    const expected = Math.max(0, (c.position - c.segmentStart) / 1000);
    return Math.abs(currentTime - expected) > (c.playing ? DRIFT_PLAYING_S : DRIFT_PAUSED_S) ? expected : null;
};
