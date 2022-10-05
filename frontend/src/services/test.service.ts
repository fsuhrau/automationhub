import http from '../http-common';
import ITestData from '../types/test';
import { AxiosResponse } from 'axios';
import ICreateTestData from '../types/request.create.test';
import ITestRunData from '../types/test.run';
import { TestExecutionType } from "../types/test.execution.type.enum";
import IUnityTestFunctionData from "../types/unity.test.function";

export const getAllTests = (projectId: string, appId: number): Promise<AxiosResponse<ITestData[]>> => {
    return http.get(`/${projectId}/app/${appId}/tests`);
};

export const getTest = (projectId: string, appId: number, id: string): Promise<AxiosResponse<ITestData>> => {
    return http.get(`/${projectId}/app/${appId}/test/${id}`);
};

export const createTest = (projectId: string, appId: number, data: ICreateTestData): Promise<AxiosResponse<ITestData>> => {
    return http.post(`/${projectId}/app/${appId}/test`, data);
};

export interface UpdateTestData {
    Name: string,
    ExecutionType: TestExecutionType,
    AllDevices: boolean,
    Devices: number[],
    RunAllTests: boolean,
    Categories: string,
    TestFunctions: IUnityTestFunctionData[],
}

export const updateTest = (projectId: string, appId: number, id: number, data: UpdateTestData): Promise<AxiosResponse<ITestData>> => {
    return http.put(`/${projectId}/app/${appId}/test/${id}`, data);
};

export const deleteTest = (projectId: string, appId: number, id: string): Promise<AxiosResponse<void>> => {
    return http.delete(`/${projectId}/app/${appId}/test/${id}`);
};

export type TestFilter = {
    name?: string;
};

export const findTest = (projectId: string, appId: number, filter?: TestFilter): Promise<AxiosResponse<ITestData[]>> => {
    return http.get(`/${projectId}/app/${appId}/tests`, { params: filter });
};

export const executeTest = (projectId: string, appId: number, id: number | null | undefined, bundleId: number, envParams: string): Promise<AxiosResponse<ITestRunData>> => {
    return http.post(`/${projectId}/app/${appId}/test/${id}/run`, {
        AppBinaryID: bundleId,
        Params: envParams,
    });
};
