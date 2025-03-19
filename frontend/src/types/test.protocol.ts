import IProtocolEntryData from './protocol.entry';
import { TestResultState } from './test.result.state.enum';
import IDeviceData from './device';
import IProtocolPerformanceEntryData from './protocol.performance.entry';
import moment from 'moment';
import ITestRunData from "./test.run";

export default interface ITestProtocolData {
    ID?: number | null,
    TestRun: ITestRunData,
    TestRunID: number,
    ParentTestProtocolID: number | null,
    DeviceID?: number | null,
    Device?: IDeviceData | null,
    TestName: string,
    StartedAt: Date,
    EndedAt?: Date,
    Entries: IProtocolEntryData[]
    TestResult: TestResultState
    Performance: IProtocolPerformanceEntryData[]
    AvgFPS: number,
    AvgMEM: number,
    AvgCPU: number,
    AvgVertexCount: number,
    AvgTriangles: number,
    HistAvgFPS: number,
    HistAvgMEM: number,
    HistAvgCPU: number,
    HistAvgVertexCount: number,
    HistAvgTriangles: number,
    TestProtocolHistory: ITestProtocolData[],
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,

    ChildProtocols: ITestProtocolData[],
}

export const duration = (startedAt: Date, endedAt: Date | undefined | null): string => {
    if (endedAt !== null && endedAt !== undefined) {
        const start = new Date(startedAt);
        const end = new Date(endedAt);
        const duration = end.valueOf() - start.valueOf();
        const m = moment.utc(duration);
        const secs = duration / 1000;
        if (secs > 60 * 60) {
            return m.format('h') + 'Std ' + m.format('m') + 'Min ' + m.format('s') + 'Sec';
        }
        if (secs > 60) {
            return m.format('m') + 'Min ' + m.format('s') + 'Sec';
        }
        return m.format('s') + 'Sec';
    }
    return 'running';
};
