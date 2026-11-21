import ITestProtocolData from './test.protocol';
import IProtocolPerformanceEntryData from './protocol.performance.entry';

export interface IBaselineCandidate {
    protocolId: number,
    testRunId: number,
    testId: number,
    startedAt: Date,
    testResult: ITestProtocolData['testResult'],
    deviceId?: number,
    device: string,
    bundle: { id: number, name: string, version: string, platform: string, tags: string } | null,
    parameter: string,
    avgFps: number,
    avgMem: number,
    avgCpu: number,
    avgVertexCount: number,
    avgTriangles: number,
}

export interface IBaseline extends IBaselineCandidate {
    performance: IProtocolPerformanceEntryData[],
}

export interface IBaselineFilter {
    // matches the bundle's name, version or tags, the run's parameter, or a run id
    q?: string,
    tag?: string,
    anyDevice?: boolean,
    limit?: number,
}

export interface IBaselineList {
    candidates: IBaselineCandidate[],
    tags: string[],
}
