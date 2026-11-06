import IProtocolEntryData from './protocol.entry';
import { TestResultState } from './test.result.state.enum';
import IDeviceData from './device';
import IProtocolPerformanceEntryData from './protocol.performance.entry';
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
    if (endedAt === null || endedAt === undefined) {
        return 'running';
    }
    const totalSecs = Math.max(0, Math.floor((new Date(endedAt).valueOf() - new Date(startedAt).valueOf()) / 1000));
    const hours = Math.floor(totalSecs / 3600);
    const mins = Math.floor((totalSecs % 3600) / 60);
    const secs = totalSecs % 60;
    if (hours > 0) {
        return `${hours}Std ${mins}Min ${secs}Sec`;
    }
    if (mins > 0) {
        return `${mins}Min ${secs}Sec`;
    }
    return `${secs}Sec`;
};
