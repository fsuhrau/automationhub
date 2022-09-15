import http from '../http-common';
import { AxiosResponse } from 'axios';
import ITestRunData from '../types/test.run';
import ITestRunResponseData from '../types/test.run.response';

export const getAllRuns = (projectId: string, appId: number, testId: string): Promise<AxiosResponse<ITestRunData[]>> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/runs`);
};

export const getLastRun = (projectId: string, appId: number, testId: string): Promise<AxiosResponse<ITestRunResponseData>> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/runs/last`);
};

export const getTestProtocol = (projectId: string, appId: number, testId: string, runId: string, protocolId: string): Promise<AxiosResponse<ITestRunData>> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/run/${runId}/${protocolId}`);
};

export const getRun = (projectId: string, appId: number, testId: string, id: string): Promise<AxiosResponse<ITestRunResponseData>> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/run/${id}`);
};
