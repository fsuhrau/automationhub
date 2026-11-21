// What the protocol page derives from a protocol, its log and its run, and
// the query that remembers its tab and comparison. Free of React.
import IProtocolKeyData from '../../../types/protocol.key';
import {IBaseline} from '../../../types/protocol.baseline';
import IProtocolPerformanceEntryData from '../../../types/protocol.performance.entry';
import ITestProtocolData from '../../../types/test.protocol';
import {mergeKey} from '../testRunModel';
import {LogEntry, LogSummary, Phase} from './logModel';

export type TabKey = 'overview' | 'log' | 'screenshots' | 'performance';

// old links used other tab names
export const tabFromQuery = (t: string | null): TabKey => {
    switch (t) {
        case 'log':
        case 'screenshots':
        case 'performance':
            return t;
        case 'checkpoint':
            return 'performance';
        default:
            return 'overview';
    }
};

// the keys of the protocol's parent and children, which the app reported
// to whichever protocol was current
export const relatedKeys = (protocols: ITestProtocolData[] | null, protocol: ITestProtocolData): IProtocolKeyData[] => {
    const {id, parentTestProtocolId: parentId} = protocol;
    return (protocols ?? [])
        .filter(p => (parentId != null && p.id === parentId) || (id != null && p.parentTestProtocolId === id))
        .flatMap(p => p.keys ?? [])
        .reduce((acc, k) => mergeKey(acc, k), [] as IProtocolKeyData[]);
};

// the protocol's own value of a key wins over the related ones
export const mergedKeys = (own: IProtocolKeyData[], related: IProtocolKeyData[]): IProtocolKeyData[] =>
    own.reduce((acc, k) => mergeKey(acc, k), related);

// the runner's setup and teardown that the log holds, with their entry counts
export const phasesOf = (log: LogEntry[]): { phase: Phase, count: number }[] =>
    (['setup', 'teardown'] as const)
        .map(phase => ({phase, count: log.filter(e => e.phase === phase).length}))
        .filter(p => p.count > 0);

// the runner often repeats the test's error, e.g. "Test X Failed"
export const ownErrors = (all: LogSummary<LogEntry>, own: LogSummary): LogEntry[] =>
    all.errors.filter(r => !r.phase || !own.errors.some(e => e.message === r.message));

// Comparison is what the performance of the protocol is held against: the
// latest earlier run on the device, or a run the user picked
export interface Comparison {
    runId: number,
    performance: IProtocolPerformanceEntryData[],
    // the run's averages; without them the usual averages of earlier runs
    averages: IBaseline | null,
}

// the picked run (?vs=<run>), or else the latest earlier one; null while the
// picked run loads
export const comparisonOf = (vs: string | null, picked: IBaseline | null, earlier: ITestProtocolData | null): Comparison | null => {
    if (vs) {
        return picked ? {runId: picked.testRunId, performance: picked.performance ?? [], averages: picked} : null;
    }
    return earlier ? {runId: earlier.testRunId, performance: earlier.performance ?? [], averages: null} : null;
};

// a loaded baseline is the picked one only while its run is the one asked for
export const pickedBaseline = (vs: string | null, loaded: IBaseline | null): IBaseline | null =>
    vs && loaded && String(loaded.testRunId) === vs ? loaded : null;
