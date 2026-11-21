import {useEffect, useMemo} from 'react';
import ITestProtocolData from '../../../types/test.protocol';
import {TestResultState} from '../../../types/test.result.state.enum';
import {neighbours, Neighbours, earlierRuns, runnerOf} from '../protocolModel';
import {buildRows, leaves, splitTestName, testResult as resultWithRunner} from '../testRunModel';
import {LogEntry, LogSummary, mergeRunnerLog, summarize} from './logModel';
import {measured, Metric, METRICS} from './performanceModel';
import {mergedKeys, ownErrors, phasesOf, relatedKeys} from './protocolViewModel';
import {ProtocolStream} from './useProtocolStream';

export interface ProtocolView {
    protocol: ITestProtocolData,
    name: string,
    suite: string,
    testResult: TestResultState,
    running: boolean,
    failed: boolean,
    place: Neighbours,
    // the runner that wrapped the test: it started the app before the test
    // method and ended it after
    runner: ITestProtocolData | null,
    // the test's own log for its steps
    summary: LogSummary,
    // with the runner's setup and teardown, for errors, screenshots and the log tab
    log: LogEntry[],
    logSummary: LogSummary,
    phases: { phase: 'setup' | 'teardown', count: number }[],
    errors: LogEntry[],
    metrics: Metric[],
    earlier: ITestProtocolData | null,
    keys: ProtocolStream['keys'],
    entries: ProtocolStream['entries'],
}

// the runner's log is loaded with the run; its teardown comes after the
// test, so the run is loaded again when the runner ends
const useReloadWhenRunnerEnds = (endedId: number | null, runner: ITestProtocolData | null, reload: () => void): void => {
    useEffect(() => {
        if (endedId !== null && endedId === runner?.id && !runner.endedAt) reload();
    }, [endedId, runner, reload]);
};

// useProtocolView derives what the protocol page shows from the live
// protocol, its entries and the protocols of its run
export const useProtocolView = (stream: ProtocolStream, runProtocols: ITestProtocolData[] | null, onRunChanged: () => void): ProtocolView => {
    const {protocol, entries} = stream;
    const tests = useMemo(() => runProtocols ? leaves(buildRows(runProtocols)) : [], [runProtocols]);
    const place = neighbours(tests, protocol.id as number);
    const runner = runnerOf(tests, place, protocol.id as number);
    useReloadWhenRunnerEnds(stream.endedId, runner, onRunChanged);

    // the test fails with its runner, e.g. when the app crashed in teardown
    const testResult = resultWithRunner(protocol, runner);
    const summary = useMemo(() => summarize(entries), [entries]);
    const log = useMemo(() => mergeRunnerLog(protocol, entries, runner), [protocol, entries, runner]);
    const logSummary = useMemo(() => summarize(log), [log]);
    const related = useMemo(() => relatedKeys(runProtocols, protocol), [runProtocols, protocol]);
    return {
        protocol, entries, place, runner, summary, log, logSummary, testResult,
        ...splitTestName(protocol.testName),
        running: testResult === TestResultState.TestResultOpen && !protocol.endedAt,
        failed: testResult === TestResultState.TestResultFailed || testResult === TestResultState.TestResultUnstable,
        phases: useMemo(() => phasesOf(log), [log]),
        errors: useMemo(() => ownErrors(logSummary, summary), [logSummary, summary]),
        metrics: METRICS.filter(m => measured(protocol.performance, m.key)),
        earlier: earlierRuns(protocol).find(h => (h.performance ?? []).length > 0) ?? null,
        keys: useMemo(() => mergedKeys(stream.keys, related), [stream.keys, related]),
    };
};
