import http from '../http-common';
import IDeviceData from '../types/device';
import ITestRunData from '../types/test.run';
import {PlatformType} from "../types/platform.type.enum";

export const getAllDevices = (projectId: string, platform?: PlatformType): Promise<IDeviceData[]> => {
    if (platform !== undefined) {
        return http.get(`/${projectId}/devices?platform=${platform}`).then(resp => resp.data)
    }
    return http.get(`/${projectId}/devices`).then(resp => resp.data)
};

export const getDevice = (projectId: string, id: string): Promise<IDeviceData | undefined> => {
    return http.get(`/${projectId}/device/${id}`).then(resp => resp.data)
};

export const createDevice = (projectId: string, data: IDeviceData): Promise<IDeviceData> => {
    return http.post(`/${projectId}/device`, data).then(resp => resp.data)
};

export const postUnlockDevice = (projectId: string, id: number): Promise<IDeviceData | undefined> => {
    return http.post(`/${projectId}/device/${id}/unlock`).then(resp => resp.data)
};

export const updateDevice = (projectId: string, data: IDeviceData, id: number): Promise<IDeviceData> => {
    return http.put(`/${projectId}/device/${id}`, data).then(resp => resp.data)
};

export const deleteDevice = (projectId: string, id: number): Promise<void> => {
    return http.delete(`/${projectId}/device/${id}`);
};

export type DeviceFilter = {
    name?: string;
};

export const findByNameFoo = (projectId: string, filter?: DeviceFilter): Promise<IDeviceData[]> => {
    return http.get(`/${projectId}/devices`, {params: filter}).then(resp => resp.data);
};

export const runTest = (projectId: string, id: number | null | undefined, testName: string, env: string): Promise<ITestRunData> => {
    return http.post(`/${projectId}/device/${id}/tests`, {testName: testName, env: env}).then(resp => resp.data)
};
