import {useMemo} from 'react';
import {useNow} from '../../../hooks/useNow';
import {RunLifecycle} from '../../../types/run.lifecycle';
import ITestRunData from '../../../types/test.run';
import ITesRunLogEntryData from '../../../types/test.run.log.entry';
import {between, buildRows, countResults, ResultCounts, runEnd, runStatus, RunStatus, TestRow} from '../testRunModel';
import {RESULT_ORDER} from './runStatusModel';
import {RunStream} from './useRunStream';

export interface RunView {
    running: boolean,
    rows: TestRow[],
    counts: ResultCounts,
    total: number,
    logErrors: ITesRunLogEntryData[],
    status: RunStatus,
    // ms the run took or has been running; null without an end
    elapsed: number | null,
    // the run's recordings can be replayed
    recording: boolean,
    hasReplay: boolean,
}

// the replay needs the recordings, which the page only has when it was
// opened after the run; a run that finished while open was recorded when its
// test records sessions
const isRecording = (run: ITestRunData): boolean => !!run.test?.testConfig?.recordSession;

const replayable = (run: ITestRunData, running: boolean, finished: boolean): boolean =>
    !running && ((run.recordings?.length ?? 0) > 0 || (finished && isRecording(run)));

// useRunView derives what the run page shows from the run and its live stream
export const useRunView = (run: ITestRunData, state: RunLifecycle, stream: RunStream): RunView => {
    const {protocols, log, finished} = stream;
    const lifecycle: RunLifecycle = finished ? 'finished' : state;
    const running = lifecycle === 'running';
    const rows = useMemo(() => buildRows(protocols), [protocols]);
    const counts = useMemo(() => countResults(rows), [rows]);
    const logErrors = useMemo(() => log.filter(l => l.level === 'error'), [log]);
    const now = useNow(1000, running);
        const end = runEnd(lifecycle, protocols, log.map(l => l.createdAt));
    return {
        running, rows, counts, logErrors,
        recording: isRecording(run),
        total: RESULT_ORDER.reduce((sum, r) => sum + counts[r], 0),
        status: runStatus(lifecycle, counts, logErrors.length),
        elapsed: running ? now - new Date(run.createdAt).valueOf() : between(run.createdAt, end),
        hasReplay: replayable(run, running, finished),
    };
};
