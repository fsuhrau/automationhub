import http from '../http-common';
import ITestRunData from '../types/test.run';
import ITestRunResponseData from '../types/test.run.response';

export const getAllRuns = (projectId: string, appId: number, testId: string): Promise<ITestRunData[]> => {
    return http.get(`/${projectId}/app/${appId}/test/${testId}/runs`).then(resp => resp.data)
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
