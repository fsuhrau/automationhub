import {AxiosRequestConfig} from 'axios';
import {ApiClient, createApiClient} from '../api/ApiClient';
import {useApiClient} from '../api/ApiClientProvider';
import http from '../http-common';
import {useMemo} from 'react';
import ITestRunData from '../types/test.run';
import ITestRunResponseData from '../types/test.run.response';
import {IBaseline, IBaselineFilter, IBaselineList} from '../types/protocol.baseline';
import {ITestRunListEntry, ITestRunsPage} from '../types/test.run.list';
import {ProtocolRef, RunRef, TestRef} from '../types/test.run.refs';
import type {RunTestData} from '../types/test.requests';

export type {IBaseline, IBaselineCandidate, IBaselineFilter} from '../types/protocol.baseline';
export type {ITestRunListEntry, ITestRunsPage} from '../types/test.run.list';

export interface RunsQuery {
    before?: number,
    limit?: number,
}

export interface TestRunService {
    // pages through a test's runs, newest first
    runsPage(test: TestRef, query?: RunsQuery): Promise<ITestRunsPage>,
    lastRun(test: TestRef): Promise<ITestRunResponseData>,
    run(run: RunRef): Promise<ITestRunResponseData>,
    protocol(ref: ProtocolRef): Promise<ITestRunData>,
    // the other runs of the test the protocol can be compared with, newest
    // first, and the tags of the app's bundles to filter by
    baselines(ref: ProtocolRef, filter: IBaselineFilter): Promise<IBaselineList>,
    // the protocol of the same test in another run with its checkpoints, on
    // the given device or preferably the protocol's own
    baseline(ref: ProtocolRef, baselineRunId: string | number, deviceId?: string | number | null): Promise<IBaseline>,
    execute(test: TestRef, data: RunTestData): Promise<ITestRunData>,
    cancel(run: RunRef): Promise<void>,
}

const testUrl = (t: TestRef): string => `/${t.projectId}/app/${t.appId}/test/${t.testId}`;
const runUrl = (r: RunRef): string => `${testUrl(r)}/run/${r.runId}`;
const protocolUrl = (p: ProtocolRef): string => `${runUrl(p)}/${p.protocolId}`;

// older masters answer with all runs and no summaries
const legacyRuns = (all: ITestRunData[]): ITestRunsPage => ({
    runs: [...all].sort((a, b) => (b.id ?? 0) - (a.id ?? 0)).map((r): ITestRunListEntry => ({
        id: r.id as number, createdAt: r.createdAt, state: 'not_running', passed: 0, failed: 0, unstable: 0, open: 0,
        parameter: r.parameter, bundle: null, endedAt: null,
    })),
    more: false,
});

const baselineParams = (f: IBaselineFilter): AxiosRequestConfig['params'] => ({
    ...(f.q ? {q: f.q} : {}),
    ...(f.tag ? {tag: f.tag} : {}),
    ...(f.anyDevice ? {anyDevice: 'true'} : {}),
    ...(f.limit ? {limit: f.limit} : {}),
});

export const testRunService = (api: ApiClient): TestRunService => ({
    runsPage: async (test, {before, limit = 25} = {}) => {
        const data = await api.get<ITestRunsPage | ITestRunData[]>(`${testUrl(test)}/runs`, {params: {limit, ...(before ? {before} : {})}});
        return Array.isArray(data) ? legacyRuns(data) : data;
    },
    lastRun: test => api.get(`${testUrl(test)}/runs/last`),
    run: run => api.get(runUrl(run)),
    protocol: ref => api.get(protocolUrl(ref)),
    baselines: (ref, filter) => api.get(`${protocolUrl(ref)}/baselines`, {params: baselineParams(filter)}),
    baseline: (ref, baselineRunId, deviceId) =>
        api.get(`${protocolUrl(ref)}/baseline/${baselineRunId}`, {params: deviceId ? {deviceId} : {}}),
    execute: (test, data) => api.post(`${testUrl(test)}/run`, data),
    cancel: run => api.post(`${runUrl(run)}/cancel`),
});

// useTestRunService is the service over the ApiClient of the nearest provider
export const useTestRunService = (): TestRunService => {
    const api = useApiClient();
    return useMemo(() => testRunService(api), [api]);
};

const legacy = testRunService(createApiClient(http));

/** @deprecated legacy wrapper, removed in F13 */
// eslint-disable-next-line max-params
export const getRunsPage = (projectId: string, appId: number, testId: number, before?: number, limit?: number): Promise<ITestRunsPage> =>
    legacy.runsPage({projectId, appId, testId}, {before, limit});

/** @deprecated legacy wrapper, removed in F13 */
// eslint-disable-next-line max-params
export const getRun = (projectId: string, appId: number, testId: string, runId: string): Promise<ITestRunResponseData> =>
    legacy.run({projectId, appId, testId, runId});
