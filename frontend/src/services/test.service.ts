import http from '../http-common';
import {ApiClient, createApiClient} from '../api/ApiClient';
import ITestData from '../types/test';
import ITestRunData from '../types/test.run';
import {TestType} from '../types/test.type.enum';
import {RunTestData, TestScope, UpdateTestData} from '../types/test.requests';

export interface TestService {
    list: (scope: TestScope) => Promise<ITestData[]>,
    get: (scope: TestScope, id: string) => Promise<ITestData>,
    // creates a test from the edit page's settings and its type
    create: (scope: TestScope, data: { type: TestType } & UpdateTestData) => Promise<ITestData>,
    update: (scope: TestScope, id: number, data: UpdateTestData) => Promise<ITestData>,
    execute: (scope: TestScope, id: number, data: RunTestData) => Promise<ITestRunData>,
    cancelRun: (scope: TestScope, testId: number, runId: number) => Promise<void>,
}

const base = ({projectId, appId}: TestScope): string => `/${projectId}/app/${appId}`;

export const testService = (api: ApiClient): TestService => ({
    list: scope => api.get(`${base(scope)}/tests`),
    get: (scope, id) => api.get(`${base(scope)}/test/${id}`),
    create: (scope, data) => api.post(`${base(scope)}/test`, data),
    update: (scope, id, data) => api.put(`${base(scope)}/test/${id}`, data),
    execute: (scope, id, data) => api.post(`${base(scope)}/test/${id}/run`, data),
    cancelRun: (scope, testId, runId) => api.post(`${base(scope)}/test/${testId}/run/${runId}/cancel`),
});

const legacy = testService(createApiClient(http));
// the old callers pass the app id as it comes from the context, null included
const legacyScope = (projectId: string, appId: number | null): TestScope => ({projectId, appId: appId as number});

/** @deprecated legacy wrapper, removed in F13 */
export const getTest = (projectId: string, appId: number | null, id: string): Promise<ITestData> =>
    legacy.get(legacyScope(projectId, appId), id);

/** @deprecated legacy wrapper, removed in F13 */
// eslint-disable-next-line max-params
export const executeTest = (projectId: string, appId: number | null, id: number | null | undefined, data: RunTestData): Promise<ITestRunData> =>
    legacy.execute(legacyScope(projectId, appId), id as number, data);

/** @deprecated legacy wrapper, removed in F13 */
// eslint-disable-next-line max-params
export const cancelTestRun = (projectId: string, appId: number | null, testId: number, runId: number): Promise<void> =>
    legacy.cancelRun(legacyScope(projectId, appId), testId, runId);
