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

export const dataUrl = (name: string) => `/api/data/${encodeURIComponent(name)}`;

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

const isError = (level: string) => level === 'error' || level === 'exception';

export const eventIsError = (e: ReplayEvent) =>
    isError(e.level) || (e.kind === 'test-end' && e.result === TestResultState.TestResultFailed);

const testLabel = (p: ITestProtocolData) => {
    const {name, suite} = splitTestName(p.testName);
    return suite ? `${name} (${suite})` : name;
};

// buildTimeline merges the run log (all devices) with the tests of the device
// and their entries, in the order they happened.
export const buildTimeline = (run: ITestRunData, deviceId: number | null): ReplayEvent[] => {
    const events: ReplayEvent[] = [];
    (run.log ?? []).forEach(l => {
        const time = ms(l.createdAt);
        if (time === null) return;
        events.push({key: `l${l.id}`, time, kind: 'log', level: l.level, text: l.log});
    });
    const protocols = (run.protocols ?? []).filter(p => deviceId === null || !p.deviceId || p.deviceId === deviceId);
    protocols.forEach(p => {
        const id = p.id as number;
        const start = ms(p.startedAt);
        if (start !== null) {
            events.push({key: `s${id}`, time: start, kind: 'test-start', level: 'info', text: testLabel(p), protocolId: id});
        }
        const end = ms(p.endedAt);
        if (end !== null) {
            const result = p.testResult ?? TestResultState.TestResultOpen;
            events.push({key: `e${id}`, time: end, kind: 'test-end', level: result === TestResultState.TestResultFailed ? 'error' : 'info',
                text: testLabel(p), protocolId: id, result});
        }
        (p.entries ?? []).forEach((e: IProtocolEntryData) => {
            const time = ms(e.timestamp) ?? ms(e.createdAt);
            if (time === null) return;
            const screen = e.source === 'screen' && !!e.data;
            events.push({
                key: `p${e.id}`, time, kind: screen ? 'screen' : e.source === 'step' ? 'step' : 'log', level: e.level,
                text: e.message || (screen ? 'Screenshot' : ''), source: e.source, protocolId: id,
                image: screen ? dataUrl(e.data) : undefined,
            });
        });
    });
    // stable: events of the same time keep the order they were added in
    return events.map((e, i) => ({e, i})).sort((a, b) => a.e.time - b.e.time || a.i - b.i).map(x => x.e);
};

// eventIndexAt is the index of the last event at or before the time (-1 when
// none); events are sorted by time
export const eventIndexAt = (events: ReplayEvent[], time: number): number => {
    let lo = 0, hi = events.length - 1, found = -1;
    while (lo <= hi) {
        const mid = (lo + hi) >> 1;
        if (events[mid].time <= time) {
            found = mid;
            lo = mid + 1;
        } else {
            hi = mid - 1;
        }
    }
    return found;
};

// replayRange is the time span the replay covers: the recordings and the
// events, whichever reach further
export const replayRange = (segments: ReplaySegment[], events: ReplayEvent[]): { start: number, end: number } | null => {
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
export const frameIndexAt = (frames: MjpegFrame[], time: number): number => {
    let lo = 0, hi = frames.length - 1, found = -1;
    while (lo <= hi) {
        const mid = (lo + hi) >> 1;
        if (frames[mid].time <= time) {
            found = mid;
            lo = mid + 1;
        } else {
            hi = mid - 1;
        }
    }
    return found;
};

// formatOffset shows a time relative to the start of the replay, m:ss.t
export const formatOffset = (msSinceStart: number): string => {
    const total = Math.max(0, msSinceStart);
    const m = Math.floor(total / 60000);
    const s = Math.floor((total % 60000) / 1000);
    const t = Math.floor((total % 1000) / 100);
    return `${m}:${String(s).padStart(2, '0')}.${t}`;
};
