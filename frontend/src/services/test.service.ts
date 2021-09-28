import http from '../http-common';
import ITestData from '../types/test';
import { AxiosResponse } from 'axios';
import ICreateTestData from '../types/request.create.test';

export const getAllTests = (): Promise<AxiosResponse<ITestData[]>> => {
    return http.get('/tests');
};

export const getTest = (id: string): Promise<AxiosResponse<ITestData>> => {
    return http.get(`/test/${id}`);
};

export const createTest = (data: ICreateTestData): Promise<AxiosResponse<ITestData>> => {
    return http.post('/test', data);
};

export const updateTest = (data: ITestData, id: string): Promise<AxiosResponse<ITestData>> => {
    return http.put(`/test/${id}`, data);
};

export const deleteTest = (id: string): Promise<AxiosResponse<void>> => {
    return http.delete(`/test/${id}`);
};

export type TestFilter = {
    name?: string;
};

export const findTest = (filter?: TestFilter): Promise<AxiosResponse<ITestData[]>> => {
    return http.get('/tests', { params: filter });
};

export const executeTest = (id: number | null | undefined, appid: number, envParams: string): Promise<AxiosResponse<unknown>> => {
    return http.post(`/test/${id}/run`, {
        AppID: appid,
        Params: envParams,
    });
};
