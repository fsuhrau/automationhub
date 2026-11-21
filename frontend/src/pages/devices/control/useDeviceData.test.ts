import {describe, expect, it, vi} from 'vitest';
import {renderHook, waitFor} from '@testing-library/react';
import {APPS, DEVICE, fakeService} from './fakes';
import {useDeviceData} from './useDeviceData';

describe('useDeviceData', () => {
    it('loads the device and the apps', async () => {
        const report = vi.fn();
        const {result} = renderHook(() => useDeviceData(fakeService(), report));
        expect(result.current).toEqual({device: null, apps: []});
        await waitFor(() => expect(result.current.device).toEqual(DEVICE));
        expect(result.current.apps).toEqual(APPS);
        expect(report).not.toHaveBeenCalled();
    });

    it('takes an unknown device as none', async () => {
        const service = fakeService({device: vi.fn(async () => undefined)});
        const {result} = renderHook(() => useDeviceData(service, vi.fn()));
        await waitFor(() => expect(service.device).toHaveBeenCalled());
        expect(result.current.device).toBeNull();
    });

    it('reports a failure', async () => {
        const failure = new Error('down');
        const report = vi.fn();
        const service = fakeService({apps: vi.fn(async () => Promise.reject(failure))});
        renderHook(() => useDeviceData(service, report));
        await waitFor(() => expect(report).toHaveBeenCalledWith(failure));
    });
});
