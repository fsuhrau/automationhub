import ITestProtocolData from './test.protocol';
import ITesRunLogEntryData from './test.run.log.entry';
import { TestResultState } from './test.result.state.enum';
import ITestData from './test';
import { IAppBinaryData } from "./app";
import ITesRunDeviceStatusData from "./test.run.device.status";

export default interface ITestRunData {
    ID?: number,
    TestID: number,
    Test: ITestData,
    SessionID: string,
    AppBinaryID: number,
    AppBinary: IAppBinaryData | null,
    StartURL: string,
    Parameter: string,
    TestResult: TestResultState,
    Protocols: Array<ITestProtocolData>,
    Log: Array<ITesRunLogEntryData>,
    DeviceStatus: Array<ITesRunDeviceStatusData>,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
