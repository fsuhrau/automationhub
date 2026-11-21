import ITesRunLogEntryData from './test.run.log.entry';
import IProtocolEntryData from './protocol.entry';

// payload of the test_run_<id>_log event
export interface TestRunLogPayload {
    testRunId: number,
    entry: ITesRunLogEntryData,
}

// payload of the test_protocol_<id>_log event
export interface TestProtocolLogPayload {
    testProtocolId: number,
    entry: IProtocolEntryData,
}
