import ITestProtocolData from './test.protocol';
import ITesRunLogEntryData from './test.run.log.entry';
import { TestResultState } from './test.result.state.enum';
import ITestData from './test';
import { IAppBinaryData } from "./app";
import ITesRunDeviceStatusData from "./test.run.device.status";

export default interface ITestRunData {
    id?: number,
    testId: number,
    test: ITestData,
    sessionId: string,
    appBinaryId: number,
    appBinary: IAppBinaryData | null,
    startUrl: string,
    parameter: string,
    testResult: TestResultState,
    protocols: ITestProtocolData[],
    log: ITesRunLogEntryData[],
    deviceStatus: ITesRunDeviceStatusData[],
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
