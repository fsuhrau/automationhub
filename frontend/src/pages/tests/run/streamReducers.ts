// The reducers that fold live events into the state of the run and protocol
// pages, so events arriving in the same render are not lost.
import {appendNew, Identified} from '../../../lib/appendNew';
import ITestProtocolData from '../../../types/test.protocol';
import IProtocolKeyData, {ProtocolKeyPayload} from '../../../types/protocol.key';
import {mergeKey, mergeProtocol} from '../testRunModel';

// log entries carry the id the master gave them, though the types allow it to be missing
type Logged = { id?: number | null };

// addEntry appends the entry of a log event unless its id is known
export const addEntry = <T extends Logged>(state: T[], payload: { entry: T } | null): T[] =>
    payload?.entry ? appendNew(state as (T & Identified)[], [payload.entry as T & Identified]) : state;

// the page's own protocol follows the updates that carry its id
export const mergeOwnProtocol = (state: ITestProtocolData, update: ITestProtocolData | null): ITestProtocolData =>
    update?.id === state.id ? mergeProtocol([state], update)[0] : state;

// the id of the protocol that ended last
export const lastEndedId = (state: number | null, update: ITestProtocolData | null): number | null =>
    update?.endedAt ? update.id as number : state;

export const addKey = (state: IProtocolKeyData[], payload: ProtocolKeyPayload | null): IProtocolKeyData[] =>
    mergeKey(state, payload?.key);
