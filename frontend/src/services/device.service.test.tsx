import {afterEach, describe, expect, it, vi} from 'vitest';
import {renderHook} from '@testing-library/react';
import React from 'react';
import http from '../http-common';
import {ApiClientProvider} from '../api/ApiClientProvider';
import {PlatformType} from '../types/platform.type.enum';
import {deviceService, getAllDevices, getDevice, useDeviceService} from './device.service';
import {fakeApi} from '../pages/devices/testKit';

describe('deviceService', () => {
    it('lists the devices of a project, optionally of one platform', async () => {
        const api = fakeApi({'GET /p1/devices': [{id: 1}]});
        const service = deviceService(api);
        expect(await service.list('p1')).toEqual([{id: 1}]);
        await service.list('p1', PlatformType.Android);
        expect(api.calls[0].config).toBeUndefined();
        expect(api.calls[1].config).toEqual({params: {platform: PlatformType.Android}});
    });

    it('gets, unlocks, updates and removes a device', async () => {
        const api = fakeApi({
            'GET /p1/device/2': {id: 2}, 'POST /p1/device/2/unlock': {id: 2}, 'PUT /p1/device/2': {id: 2}, 'DELETE /p1/device/2': undefined,
        });
        const service = deviceService(api);
        await service.get('p1', '2');
        await service.unlock('p1', 2);
        await service.update('p1', 2, {alias: 'a', customParameter: []});
        await service.remove('p1', 2);
        expect(api.calls.map(c => `${c.method} ${c.url}`)).toEqual([
            'GET /p1/device/2', 'POST /p1/device/2/unlock', 'PUT /p1/device/2', 'DELETE /p1/device/2',
        ]);
        expect(api.calls[2].data).toEqual({alias: 'a', customParameter: []});
    });

    it('starts a test on a device', async () => {
        const api = fakeApi({'POST /p1/device/2/tests': {id: 5}});
        expect(await deviceService(api).runTest('p1', 2, {testName: 't', env: 'e'})).toEqual({id: 5});
        expect(api.calls[0].data).toEqual({testName: 't', env: 'e'});
    });

    it('loads the test history with a limit', async () => {
        const api = fakeApi({'GET /p1/device/2/protocols': [{id: 1}]});
        const service = deviceService(api);
        expect(await service.protocols('p1', 2)).toEqual([{id: 1}]);
        await service.protocols('p1', 2, 3);
        expect(api.calls.map(c => c.config)).toEqual([{params: {limit: 10}}, {params: {limit: 3}}]);
    });

    it('rejects a history the master does not have', async () => {
        const api = fakeApi({'GET /p1/device/2/protocols': '<html>'});
        await expect(deviceService(api).protocols('p1', 2)).rejects.toThrow('the master has no device test history');
    });
});

describe('useDeviceService', () => {
    it('uses the client of the provider and keeps the service while it stays', () => {
        const api = fakeApi({'GET /p1/devices': []});
        const wrapper = ({children}: { children: React.ReactNode }): React.ReactElement =>
            <ApiClientProvider client={api}>{children}</ApiClientProvider>;
        const {result, rerender} = renderHook(() => useDeviceService(), {wrapper});
        const first = result.current;
        rerender();
        expect(result.current).toBe(first);
        void first.list('p1');
        expect(api.calls).toHaveLength(1);
    });
});

describe('legacy wrappers', () => {
    afterEach(() => vi.restoreAllMocks());

    it('getAllDevices and getDevice go through the axios instance', async () => {
        const get = vi.spyOn(http, 'get').mockResolvedValue({data: [{id: 1}]});
        expect(await getAllDevices('p1', PlatformType.iOS)).toEqual([{id: 1}]);
        expect(get).toHaveBeenCalledWith('/p1/devices', {params: {platform: PlatformType.iOS}});
        get.mockResolvedValue({data: {id: 4}});
        expect(await getDevice('p1', '4')).toEqual({id: 4});
        expect(get).toHaveBeenLastCalledWith('/p1/device/4', undefined);
    });
});
