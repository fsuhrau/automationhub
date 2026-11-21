// Performance of a protocol: checkpoints, the metrics and how they compare
// with an earlier run. Kept free of React so it can be checked on its own.
import IProtocolPerformanceEntryData from '../../../types/protocol.performance.entry';
import ITestProtocolData from '../../../types/test.protocol';

export interface CheckpointRow {
    checkpoint: string,
    // seconds since the test started
    at: number,
    // seconds since the previous checkpoint
    took: number,
    // the same checkpoint in the earlier run, if it had it
    earlierTook: number | null,
}

interface TimedCheckpoint {
    checkpoint: string,
    at: number,
    took: number,
}

const timed = (perf: IProtocolPerformanceEntryData[] | null | undefined): TimedCheckpoint[] =>
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

const compact = (v: number): string => v >= 1000 ? `${(v / 1000).toFixed(1)}k` : v.toFixed(0);

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
