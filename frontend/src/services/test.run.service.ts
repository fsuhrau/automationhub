import http from '../http-common';
import { AxiosResponse } from 'axios';
import ITestRunData from '../types/test.run';
import ITestRunResponseData from '../types/test.run.response';

export const getAllRuns = (testId: string): Promise<AxiosResponse<ITestRunData[]>> => {
    return http.get(`/test/${testId}/runs`);
};

export const getLastRun = (testId: string): Promise<AxiosResponse<ITestRunResponseData>> => {
    return http.get(`/test/${testId}/runs/last`);
};

export const getTestProtocol = (testId: string, runId: string, protocolId: string): Promise<AxiosResponse<ITestRunData>> => {
    return http.get(`/test/${testId}/run/${runId}/${protocolId}`);
};

export const getRun = (testId: string, id: string): Promise<AxiosResponse<ITestRunResponseData>> => {
    return http.get(`/test/${testId}/run/${id}`);
};
