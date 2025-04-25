import http from '../http-common';
import ITestData from '../types/test';
import ICreateTestData from '../types/request.create.test';
import ITestRunData from '../types/test.run';
import { TestExecutionType } from "../types/test.execution.type.enum";
import IUnityTestFunctionData from "../types/unity.test.function";
import { UnityTestCategory } from "../types/unity.test.category.type.enum";

export const getAllTests = (projectId: string, appId: number | null): Promise<ITestData[]> => {
    return http.get(`/${projectId}/app/${appId}/tests`).then(resp => resp.data)
};

export const getTest = (projectId: string, appId: number | null, id: string): Promise<ITestData> => {
    return http.get(`/${projectId}/app/${appId}/test/${id}`).then(resp => resp.data)
};

export const createTest = (projectId: string, appId: number | null, data: ICreateTestData): Promise<ITestData> => {
    return http.post(`/${projectId}/app/${appId}/test`, data).then(resp => resp.data)
};

export interface UpdateTestData {
    name: string,
    executionType: TestExecutionType,
    allDevices: boolean,
    devices: number[],
    unityTestCategoryType: UnityTestCategory,
    categories: string,
    testFunctions: IUnityTestFunctionData[],
}

export const updateTest = (projectId: string, appId: number | null, id: number, data: UpdateTestData): Promise<ITestData> => {
    return http.put(`/${projectId}/app/${appId}/test/${id}`, data).then(resp => resp.data)
};

export const deleteTest = (projectId: string, appId: number | null, id: string): Promise<void> => {
    return http.delete(`/${projectId}/app/${appId}/test/${id}`);
};

export type TestFilter = {
    name?: string;
};

export const findTest = (projectId: string, appId: number | null, filter?: TestFilter): Promise<ITestData[]> => {
    return http.get(`/${projectId}/app/${appId}/tests`, { params: filter }).then(resp => resp.data)
};

export interface RunTestData {
    appBinaryId: number | null,
    startUrl: string | null,
    params: string,
}

export const executeTest = (projectId: string, appId: number | null, id: number | null | undefined, testData: RunTestData): Promise<ITestRunData> => {
    return http.post(`/${projectId}/app/${appId}/test/${id}/run`, testData).then(resp => resp.data)
};

export const cancelTestRun = (projectId: string, appId: number | null, testId: number, runId: number): Promise<void> => {
    return http.post(`/${projectId}/app/${appId}/test/${testId}/run/${runId}/cancel`);
};