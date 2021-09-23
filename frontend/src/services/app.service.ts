import http from '../http-common';
import IAppData from '../types/app';
import { AxiosResponse } from 'axios';

export const getAllApps = (): Promise<AxiosResponse<IAppData[]>> => {
    return http.get('/apps');
};

export const getApp = (id: string): Promise<AxiosResponse<IAppData | undefined>> => {
    return http.get(`/app/${id}`);
};

export const createApp = (data: IAppData): Promise<AxiosResponse<IAppData>> => {
    return http.post('/app', data);
};

export const updateApp = (data: IAppData, id: number): Promise<AxiosResponse<IAppData>> => {
    return http.put(`/app/${id}`, data);
};

export const deleteApp = (id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/app/${id}`);
};

export type AppFilter = {
    name?: string;
};

export const findApp = (filter?: AppFilter): Promise<AxiosResponse<IAppData[]>> => {
    return http.get('/apps', { params: filter });
};
