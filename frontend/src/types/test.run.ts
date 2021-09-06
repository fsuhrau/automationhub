import ITestProtocolData from './test.protocol';
import ITesRunLogEntryData from './test.run.log.entry';

export default interface ITestRunData {
    ID?: number,
    TestID: number,
    SessionID: string,
    Protocols: Array<ITestProtocolData>,
    Log: Array<ITesRunLogEntryData>,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
