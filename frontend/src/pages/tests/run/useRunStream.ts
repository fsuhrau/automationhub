import {useSSE} from 'react-hooks-sse';
import ITestProtocolData from '../../../types/test.protocol';
import ITestRunData from '../../../types/test.run';
import ITesRunLogEntryData from '../../../types/test.run.log.entry';
import {TestRunLogPayload} from '../../../types/test.run.stream';
import {mergeProtocol} from '../testRunModel';
import {addEntry} from './streamReducers';

export interface RunStream {
    protocols: ITestProtocolData[],
    log: ITesRunLogEntryData[],
    // the run ended while the page was open
    finished: boolean,
}

// useRunStream follows the live protocol, log and finish events of a run,
// starting from what the loaded run holds. The events come through the
// SSEProvider, whose source tests replace.
export const useRunStream = (run: ITestRunData): RunStream => {
    const channel = `test_run_${run.id}`;
    const protocols = useSSE<ITestProtocolData[], ITestProtocolData | null>(
        `${channel}_protocol`, run.protocols ?? [], {stateReducer: (state, action) => mergeProtocol(state, action.data)});
    const log = useSSE<ITesRunLogEntryData[], TestRunLogPayload | null>(
        `${channel}_log`, run.log ?? [], {stateReducer: (state, action) => addEntry(state, action.data)});
    const finished = useSSE<object | null>(`${channel}_finished`, null);
    return {protocols, log, finished: !!finished};
};
