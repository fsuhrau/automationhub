import ITestProtocolData from './test.protocol';
import ITesRunLogEntryData from './test.run.log.entry';
import { TestResultState } from './test.result.state.enum';
import ITestData from './test';
import { IAppBinaryData } from "./app";

export default interface ITestRunData {
    ID?: number,
    TestID: number,
    Test: ITestData,
    SessionID: string,
    AppID: number,
    App: IAppBinaryData | null,
    Parameter: string,
    TestResult: TestResultState,
    Protocols: Array<ITestProtocolData>,
    Log: Array<ITesRunLogEntryData>,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
