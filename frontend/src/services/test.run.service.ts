import http from '../http-common';
import ITestRunData from '../types/test.run';
import {IRunSummary} from '../types/test';
import ITestRunResponseData from '../types/test.run.response';

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
