// The place of a protocol among the tests of its run, and its earlier runs.
// Kept free of React so it can be checked on its own.
import ITestProtocolData from '../../types/test.protocol';
import {TestResultState} from '../../types/test.result.state.enum';
import {TestRow} from './testRunModel';

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
    const problem = (t: TestRow): boolean => t.result === TestResultState.TestResultFailed || t.result === TestResultState.TestResultUnstable;
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

// the runner that wrapped the test at a place of the run, unless the
// protocol is the runner itself
export const runnerOf = (tests: TestRow[], place: Neighbours, protocolId: number): ITestProtocolData | null => {
    const row = place.position > 0 ? tests[place.position - 1] : null;
    return row?.runner && row.runner.id !== protocolId ? row.runner : null;
};
