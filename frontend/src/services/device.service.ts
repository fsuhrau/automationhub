import {useMemo} from 'react';
import {ApiClient} from '../api/ApiClient';
import {useApiClient} from '../api/ApiClientProvider';
import IDeviceData from '../types/device';
import ITestRunData from '../types/test.run';
import ITestProtocolData from '../types/test.protocol';
import {PlatformType} from '../types/platform.type.enum';
import {DeviceTestRequest, DeviceUpdate} from '../types/device.requests';

export interface DeviceService {
    list(projectId: string, platform?: PlatformType): Promise<IDeviceData[]>;

    get(projectId: string, id: string): Promise<IDeviceData | undefined>;

    // ends the test run or control session that holds the device
    unlock(projectId: string, id: number): Promise<IDeviceData | undefined>;

    // returns the saved device
    update(projectId: string, id: number, data: DeviceUpdate): Promise<IDeviceData>;

    remove(projectId: string, id: number): Promise<void>;

    runTest(projectId: string, id: number, request: DeviceTestRequest): Promise<ITestRunData>;

    // the latest test results on the device, newest first (tests of this project)
    protocols(projectId: string, id: number, limit?: number): Promise<ITestProtocolData[]>;
}

export const deviceService = (api: ApiClient): DeviceService => ({
    list: (projectId, platform) => api.get(`/${projectId}/devices`, platform === undefined ? undefined : {params: {platform}}),
    get: (projectId, id) => api.get(`/${projectId}/device/${id}`),
    unlock: (projectId, id) => api.post(`/${projectId}/device/${id}/unlock`),
    update: (projectId, id, data) => api.put(`/${projectId}/device/${id}`, data),
    remove: (projectId, id) => api.delete(`/${projectId}/device/${id}`),
    runTest: (projectId, id, request) => api.post(`/${projectId}/device/${id}/tests`, request),
    protocols: async (projectId, id, limit = 10) => {
        const body = await api.get<unknown>(`/${projectId}/device/${id}/protocols`, {params: {limit}});
        // older masters don't have the route and answer with something else
        if (!Array.isArray(body)) throw new Error('the master has no device test history');
        return body as ITestProtocolData[];
    },
});

// the device service on the ApiClient of the nearest provider
export const useDeviceService = (): DeviceService => {
    const api = useApiClient();
    return useMemo(() => deviceService(api), [api]);
};
