import {useSSE} from 'react-hooks-sse';
import IProtocolEntryData from '../../../types/protocol.entry';
import IProtocolKeyData, {ProtocolKeyPayload} from '../../../types/protocol.key';
import ITestProtocolData from '../../../types/test.protocol';
import {TestProtocolLogPayload} from '../../../types/test.run.stream';
import {addEntry, addKey, lastEndedId, mergeOwnProtocol} from './streamReducers';

export interface ProtocolStream {
    protocol: ITestProtocolData,
    entries: IProtocolEntryData[],
    // the keys the app reported for this protocol
    keys: IProtocolKeyData[],
    // the protocol of the run that ended last
    endedId: number | null,
}

// useProtocolStream follows the live result, log, keys and endings of the
// protocols of a run, starting from the loaded protocol. The events come
// through the SSEProvider, whose source tests replace.
export const useProtocolStream = (runId: number | undefined, initial: ITestProtocolData): ProtocolStream => {
    const runChannel = `test_run_${runId}_protocol`;
    const protocol = useSSE<ITestProtocolData, ITestProtocolData | null>(
        runChannel, initial, {stateReducer: (state, action) => mergeOwnProtocol(state, action.data)});
    const entries = useSSE<IProtocolEntryData[], TestProtocolLogPayload | null>(
        `test_protocol_${initial.id}_log`, initial.entries ?? [], {stateReducer: (state, action) => addEntry(state, action.data)});
    const keys = useSSE<IProtocolKeyData[], ProtocolKeyPayload | null>(
        `test_protocol_${initial.id}_keys`, initial.keys ?? [], {stateReducer: (state, action) => addKey(state, action.data)});
    const endedId = useSSE<number | null, ITestProtocolData | null>(
        runChannel, null, {stateReducer: (state, action) => lastEndedId(state, action.data)});
    return {protocol, entries, keys, endedId};
};
