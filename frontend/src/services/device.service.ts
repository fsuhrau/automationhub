import http from '../http-common';
import IDeviceData from '../types/device';
import { AxiosResponse } from 'axios';
import ITestRunData from '../types/test.run';
import { PlatformType } from "../types/platform.type.enum";

export const getAllDevices = (projectId: string, platform?: PlatformType): Promise<AxiosResponse<IDeviceData[]>> => {
    if (platform !== undefined) {
        return http.get(`/${projectId}/devices?platform=${platform}`);
    }
    return http.get(`/${projectId}/devices`);
};

export const getDevice = (projectId: string, id: string): Promise<AxiosResponse<IDeviceData | undefined>> => {
    return http.get(`/${projectId}/device/${id}`);
};

export const createDevice = (projectId: string, data: IDeviceData): Promise<AxiosResponse<IDeviceData>> => {
    return http.post(`/${projectId}/device`, data);
};

export const updateDevice = (projectId: string, data: IDeviceData, id: number): Promise<AxiosResponse<IDeviceData>> => {
    return http.put(`/${projectId}/device/${id}`, data);
};

export const deleteDevice = (projectId: string, id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/${projectId}/device/${id}`);
};

export type DeviceFilter = {
    name?: string;
};

export const findByNameFoo = (projectId: string, filter?: DeviceFilter): Promise<AxiosResponse<IDeviceData[]>> => {
    return http.get(`/${projectId}/devices`, { params: filter });
};

export const runTest = (projectId: string, id: number | null | undefined, testName: string, env: string): Promise<AxiosResponse<ITestRunData>> => {
    return http.post(`/${projectId}/device/${id}/tests`, { testName: testName, env: env });
};
