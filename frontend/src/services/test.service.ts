import {useMemo} from 'react';
import {ApiClient} from '../api/ApiClient';
import {useApiClient} from '../api/ApiClientProvider';
import ITestData from '../types/test';
import {TestType} from '../types/test.type.enum';
import {TestScope, UpdateTestData} from '../types/test.requests';

// the tests of an app; starting and cancelling runs is the test run service
export interface TestService {
    list: (scope: TestScope) => Promise<ITestData[]>,
    get: (scope: TestScope, id: string) => Promise<ITestData>,
    // creates a test from the edit page's settings and its type
    create: (scope: TestScope, data: { type: TestType } & UpdateTestData) => Promise<ITestData>,
    update: (scope: TestScope, id: number, data: UpdateTestData) => Promise<ITestData>,
}

const base = ({projectId, appId}: TestScope): string => `/${projectId}/app/${appId}`;

export const testService = (api: ApiClient): TestService => ({
    list: scope => api.get(`${base(scope)}/tests`),
    get: (scope, id) => api.get(`${base(scope)}/test/${id}`),
    create: (scope, data) => api.post(`${base(scope)}/test`, data),
    update: (scope, id, data) => api.put(`${base(scope)}/test/${id}`, data),
});

// the test service on the ApiClient of the nearest provider
export const useTestService = (): TestService => {
    const api = useApiClient();
    return useMemo(() => testService(api), [api]);
};
