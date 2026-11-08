// The data side of the protocol page: log summaries and filters, the place of
// the protocol in its run, checkpoints and performance against earlier runs.
// Kept free of React so it can be checked on its own.
import IProtocolEntryData from '../../types/protocol.entry';
import IProtocolPerformanceEntryData from '../../types/protocol.performance.entry';
import ITestProtocolData from '../../types/test.protocol';
import {TestResultState} from '../../types/test.result.state.enum';
import {TestRow} from './testRunModel';

// formatRuntime shows seconds since the test started as m:ss.mmm (h:mm:ss
// past an hour), the clock the log is read by
export const formatRuntime = (seconds: number | null | undefined): string => {
    const s = Math.max(0, seconds ?? 0);
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    const sec = s % 60;
    if (h > 0) {
        return `${h}:${String(m).padStart(2, '0')}:${String(Math.floor(sec)).padStart(2, '0')}`;
    }
    return `${m}:${sec.toFixed(3).padStart(6, '0')}`;
};

export const isError = (e: IProtocolEntryData): boolean => e.level === 'error';
export const isScreen = (e: IProtocolEntryData): boolean => e.source === 'screen' && !!e.data;
export const isStep = (e: IProtocolEntryData): boolean => e.source === 'step';

// sourceName labels the log sources; entries without one come from the runner
export const sourceName = (source: string): string => source === '' ? 'runner' : source;

export interface LogSummary {
    errors: IProtocolEntryData[],
    steps: IProtocolEntryData[],
    screens: IProtocolEntryData[],
    // entries per source, in the order sources first appear
    sources: { source: string, count: number }[],
}

export const summarize = (entries: IProtocolEntryData[]): LogSummary => {
    const counts = new Map<string, number>();
    entries.forEach(e => counts.set(sourceName(e.source), (counts.get(sourceName(e.source)) ?? 0) + 1));
    return {
        errors: entries.filter(isError),
        steps: entries.filter(isStep),
        screens: entries.filter(isScreen),
        sources: Array.from(counts, ([source, count]) => ({source, count})),
    };
};

export interface LogFilter {
    // sources switched off
    hidden: Set<string>,
    errorsOnly: boolean,
    text: string,
}

export const filterLog = (entries: IProtocolEntryData[], f: LogFilter): IProtocolEntryData[] => {
    const needle = f.text.trim().toLowerCase();
    return entries.filter(e => (!f.errorsOnly || isError(e))
        && !f.hidden.has(sourceName(e.source))
        && (needle === '' || (e.message ?? '').toLowerCase().includes(needle)));
};

// headline is the first line of a message, the rest is its detail (stack
// traces, JSON)
export const splitMessage = (message: string): { headline: string, detail: string } => {
    const text = (message ?? '').replace(/\s+$/, '');
    const nl = text.indexOf('\n');
    return nl < 0 ? {headline: text, detail: ''} : {headline: text.substring(0, nl), detail: text.substring(nl + 1)};
};

// logAsText is the log as a plain text file for download
export const logAsText = (testName: string, entries: IProtocolEntryData[]): string =>
    [`# ${testName}`, ...entries.map(e =>
        `${formatRuntime(e.runtime)}\t${sourceName(e.source)}\t${e.level}\t${(e.message ?? '').replace(/\n/g, '\n\t\t\t')}${e.data ? `\t[${e.data}]` : ''}`)]
        .join('\n') + '\n';

export interface Neighbours {
    // 1-based position among the tests of the run, 0 when not found
    position: number,
    total: number,
    prev: TestRow | null,
    next: TestRow | null,
    // the next failed or unstable test after this one
    nextProblem: TestRow | null,
}

// neighbours places a protocol among the tests of its run; a row is found by
// its own protocol or its runner's
export const neighbours = (tests: TestRow[], protocolId: number): Neighbours => {
    const i = tests.findIndex(t => t.protocol.id === protocolId || t.runner?.id === protocolId);
    if (i < 0) {
        return {position: 0, total: tests.length, prev: null, next: null, nextProblem: null};
    }
    const problem = (t: TestRow) => t.result === TestResultState.TestResultFailed || t.result === TestResultState.TestResultUnstable;
    return {
        position: i + 1,
        total: tests.length,
        prev: tests[i - 1] ?? null,
        next: tests[i + 1] ?? null,
        nextProblem: tests.slice(i + 1).find(problem) ?? null,
    };
};

// earlierRuns are the history entries of the same test from other runs, the
// newest first. The history also lists the protocol itself and its runner,
// which aren't a comparison.
export const earlierRuns = (protocol: ITestProtocolData): ITestProtocolData[] =>
    (protocol.testProtocolHistory ?? [])
        .filter(h => h.id !== protocol.id && h.testRunId !== protocol.testRunId)
        .sort((a, b) => new Date(b.startedAt).valueOf() - new Date(a.startedAt).valueOf());

export interface CheckpointRow {
    checkpoint: string,
    // seconds since the test started
    at: number,
    // seconds since the previous checkpoint
    took: number,
    // the same checkpoint in the earlier run, if it had it
    earlierTook: number | null,
}

const timed = (perf: IProtocolPerformanceEntryData[] | null | undefined) =>
    (perf ?? []).filter(p => p.checkpoint !== 'schedule')
        .map((p, i, all) => ({checkpoint: p.checkpoint, at: p.runtime, took: i === 0 ? p.runtime : p.runtime - all[i - 1].runtime}));

// checkpointRows lists the checkpoints with the time each took, next to the
// same checkpoint of an earlier run (matched by name, first occurrence)
export const checkpointRows = (perf: IProtocolPerformanceEntryData[] | null | undefined,
                               earlier: IProtocolPerformanceEntryData[] | null | undefined): CheckpointRow[] => {
    const before = new Map<string, number>();
    timed(earlier).forEach(c => {
        if (!before.has(c.checkpoint)) before.set(c.checkpoint, c.took);
    });
    return timed(perf).map(c => ({...c, earlierTook: before.get(c.checkpoint) ?? null}));
};

export type Verdict = 'better' | 'worse' | 'same' | null;

// compare says whether a value is meaningfully better or worse than its usual
// value; within 5% it's the same, and without a usual value there's nothing
// to say
export const compare = (value: number | null | undefined, usual: number | null | undefined, higherIsBetter: boolean): Verdict => {
    if (value === null || value === undefined || !isFinite(value) || !usual || !isFinite(usual)) {
        return null;
    }
    const change = (value - usual) / Math.abs(usual);
    if (Math.abs(change) < 0.05) return 'same';
    return (change > 0) === higherIsBetter ? 'better' : 'worse';
};

export interface Metric {
    key: 'fps' | 'mem' | 'cpu' | 'vertexCount' | 'triangles',
    label: string,
    unit: string,
    higherIsBetter: boolean,
    // the protocol's average and the usual average of earlier runs
    avg: (p: ITestProtocolData) => number | null | undefined,
    usual: (p: ITestProtocolData) => number | null | undefined,
    format: (v: number) => string,
}

const compact = (v: number) => v >= 1000 ? `${(v / 1000).toFixed(1)}k` : v.toFixed(0);

export const METRICS: Metric[] = [
    {key: 'fps', label: 'Frame rate', unit: 'fps', higherIsBetter: true, avg: p => p.avgFps, usual: p => p.histAvgFps, format: v => `${v.toFixed(1)} fps`},
    {key: 'mem', label: 'Memory', unit: 'MB', higherIsBetter: false, avg: p => p.avgMem, usual: p => p.histAvgMem,
        format: v => v >= 1024 ? `${(v / 1024).toFixed(2)} GB` : `${v.toFixed(0)} MB`},
    {key: 'cpu', label: 'CPU', unit: '%', higherIsBetter: false, avg: p => p.avgCpu, usual: p => p.histAvgCpu, format: v => `${v.toFixed(1)} %`},
    {key: 'vertexCount', label: 'Vertices', unit: '', higherIsBetter: false, avg: p => p.avgVertexCount, usual: p => p.histAvgVertexCount, format: compact},
    {key: 'triangles', label: 'Triangles', unit: '', higherIsBetter: false, avg: p => p.avgTriangles, usual: p => p.histAvgTriangles, format: compact},
];

// measured tells whether a metric was recorded at all; a metric that is zero
// or missing in every sample (CPU on some platforms) isn't shown
export const measured = (perf: IProtocolPerformanceEntryData[] | null | undefined, key: Metric['key']): boolean =>
    (perf ?? []).some(p => typeof p[key] === 'number' && p[key] !== 0);

// Averages of a run picked to compare with, by metric
export interface Averages {
    avgFps: number,
    avgMem: number,
    avgCpu: number,
    avgVertexCount: number,
    avgTriangles: number,
}

const AVERAGE: Record<Metric['key'], keyof Averages> = {
    fps: 'avgFps', mem: 'avgMem', cpu: 'avgCpu', vertexCount: 'avgVertexCount', triangles: 'avgTriangles',
};

export const averageOf = (a: Averages, key: Metric['key']): number => a[AVERAGE[key]];

// bundleLabel names the bundle a run used, by its name and version
export const bundleLabel = (b: { name: string, version: string } | null | undefined): string =>
    b ? [b.name, b.version].filter(s => s && s.length > 0).join(' ') : '';
