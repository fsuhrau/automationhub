import http from '../http-common';
import ITestRunData from '../types/test.run';
import {IRunSummary} from '../types/test';
import ITestRunResponseData from '../types/test.run.response';
import ITestProtocolData from '../types/test.protocol';
import IProtocolPerformanceEntryData from '../types/protocol.performance.entry';

export interface ITestRunListEntry extends Omit<IRunSummary, 'status'> {
    // undefined when an older master doesn't summarize runs
    status?: IRunSummary['status'],
    parameter: string,
    bundle: { id: number, name: string, version: string } | null,
    endedAt: Date | null,
}

export interface ITestRunsPage {
    runs: ITestRunListEntry[],
    // older runs exist; load them with before = the last run's id
    more: boolean,
}

// getRunsPage pages through a test's runs, newest first
export const getRunsPage = (projectId: string, appId: number, testId: number, before?: number, limit: number = 25): Promise<ITestRunsPage> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/runs`, {params: {limit, ...(before ? {before} : {})}}).then(resp => {
        if (Array.isArray(resp.data)) {
            // older masters answer with all runs and no summaries
            const runs = (resp.data as ITestRunData[]).sort((a, b) => (b.id ?? 0) - (a.id ?? 0)).map(r => ({
                id: r.id as number, createdAt: r.createdAt, state: 'not_running' as const, passed: 0, failed: 0, unstable: 0, open: 0,
                parameter: r.parameter, bundle: null, endedAt: null,
            }));
            return {runs, more: false};
        }
        return resp.data;
    })
};

export const getLastRun = (projectId: string, appId: number, testId: string): Promise<ITestRunResponseData> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/runs/last`).then(resp => resp.data)
};

export const getTestProtocol = (projectId: string, appId: number, testId: string, runId: string, protocolId: string): Promise<ITestRunData> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/run/${runId}/${protocolId}`).then(resp => resp.data)
};

export const getRun = (projectId: string, appId: number, testId: string, id: string): Promise<ITestRunResponseData> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/run/${id}`).then(resp => resp.data)
};

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

// getBaselines lists the other runs of the test the protocol can be compared
// with, newest first, and the tags of the app's bundles to filter by
export const getBaselines = (projectId: string, appId: number, testId: string | number, runId: string | number, protocolId: string | number,
                             filter: IBaselineFilter): Promise<{ candidates: IBaselineCandidate[], tags: string[] }> => {
    const params = {
        ...(filter.q ? {q: filter.q} : {}),
        ...(filter.tag ? {tag: filter.tag} : {}),
        ...(filter.anyDevice ? {anyDevice: 'true'} : {}),
        ...(filter.limit ? {limit: filter.limit} : {}),
    };
    return http.get(`/${projectId}/app/${appId}/test/${testId}/run/${runId}/${protocolId}/baselines`, {params}).then(resp => resp.data)
};

// getBaseline loads the protocol of the same test in another run with its
// checkpoints, on the given device or preferably the protocol's own
export const getBaseline = (projectId: string, appId: number, testId: string | number, runId: string | number, protocolId: string | number,
                            baselineRunId: string | number, deviceId?: string | number | null): Promise<IBaseline> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/run/${runId}/${protocolId}/baseline/${baselineRunId}`,
        {params: deviceId ? {deviceId} : {}}).then(resp => resp.data)
};
