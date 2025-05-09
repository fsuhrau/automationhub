import IProtocolEntryData from './protocol.entry';
import { TestResultState } from './test.result.state.enum';
import IDeviceData from './device';
import IProtocolPerformanceEntryData from './protocol.performance.entry';
import moment from 'moment';
import ITestRunData from "./test.run";

export default interface ITestProtocolData {
    id?: number | null,
    testRun: ITestRunData,
    testRunId: number,
    parentTestProtocolId: number | null,
    deviceId?: number | null,
    device?: IDeviceData | null,
    testName: string,
    startedAt: Date,
    endedAt?: Date,
    entries: IProtocolEntryData[]
    testResult: TestResultState
    performance: IProtocolPerformanceEntryData[]
    avgFps: number,
    avgMem: number,
    avgCpu: number,
    avgVertexCount: number,
    avgTriangles: number,
    histAvgFps: number,
    histAvgMem: number,
    histAvgCpu: number,
    histAvgVertexCount: number,
    histAvgTriangles: number,
    testProtocolHistory: ITestProtocolData[] | null,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,

    childProtocols: ITestProtocolData[],
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
