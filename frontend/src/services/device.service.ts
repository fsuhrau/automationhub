import http from '../http-common';
import IDeviceData from '../types/device';
import { AxiosResponse } from 'axios';
import ITestRunData from '../types/test.run';

export const getAllDevices = (): Promise<AxiosResponse<IDeviceData[]>> => {
    return http.get('/devices');
};

export const getDevice = (id: string): Promise<AxiosResponse<IDeviceData | undefined>> => {
    return http.get(`/device/${id}`);
};

export const createDevice = (data: IDeviceData): Promise<AxiosResponse<IDeviceData>> => {
    return http.post('/device', data);
};

export const updateDevice = (data: IDeviceData, id: number): Promise<AxiosResponse<IDeviceData>> => {
    return http.put(`/device/${id}`, data);
};

export const deleteDevice = (id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/device/${id}`);
};

export type DeviceFilter = {
    name?: string;
};

export const findByNameFoo = (filter?: DeviceFilter): Promise<AxiosResponse<IDeviceData[]>> => {
    return http.get('/devices', { params: filter });
};

export const runTest = (id: number | null | undefined, testName: string, env: string): Promise<AxiosResponse<ITestRunData>> => {
    return http.post(`/device/${id}/tests`, {testName: testName, env: env});
};
