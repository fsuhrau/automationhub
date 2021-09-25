import IAppData from './app';
import ITestProtocolData from './test.protocol';
import ITesRunLogEntryData from './test.run.log.entry';
import { TestResultState } from './test.result.state.enum';

export default interface ITestRunData {
    ID?: number,
    TestID: number,
    SessionID: string,
    AppID: number,
    App: IAppData | null,
    TestResult: TestResultState,
    Protocols: Array<ITestProtocolData>,
    Log: Array<ITesRunLogEntryData>,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
