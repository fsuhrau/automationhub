import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {fakeService, FREE, heldBy, leaseOf} from './fakes';
import {loadLease, storeLease} from './leaseStorage';
import {DeviceControlApi} from '../../../types/device.control.api';
import {DeviceLeaseState, LEASE_REFRESH_MS, useDeviceLease} from './useDeviceLease';

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
    sessionStorage.clear();
});

interface Setup extends RenderHookResult<DeviceLeaseState, unknown> {
    service: DeviceControlApi,
}

const setup = (over: Partial<DeviceControlApi> = {}): Setup => {
    const service = fakeService(over);
    const hook = renderHook(() => useDeviceLease(service, 7));
    return {service, ...hook};
};

const tick = (ms = 0): Promise<void> => act(async () => {
    await vi.advanceTimersByTimeAsync(ms);
});

describe('useDeviceLease', () => {
    it('reads the lock status at once and then every few seconds', async () => {
        const {service, result} = setup();
        await tick();
        expect(result.current.status).toEqual(FREE);
        await tick(LEASE_REFRESH_MS);
        expect(service.lockStatus).toHaveBeenCalledTimes(2);
    });

    it('moves the time on as often as the status is read', async () => {
        const {result} = setup();
        const first = result.current.now;
        await tick(LEASE_REFRESH_MS);
        expect(result.current.now).toBe(first + LEASE_REFRESH_MS);
    });

    it('takes control, keeps the token for the tab and reads the status again', async () => {
        const {service, result} = setup({lockStatus: vi.fn(async () => heldBy())});
        await tick();
        await act(() => result.current.acquire());
        expect(result.current.token).toBe('tok-1');
        expect(loadLease(7)).toBe('tok-1');
        expect(result.current.status?.lease?.owner).toBe('anna');
        expect(vi.mocked(service.lockStatus).mock.calls.length).toBeGreaterThan(1);
    });

    it('holds no token when the hub sends none', async () => {
        const {result} = setup({lock: vi.fn(async () => leaseOf({token: undefined}))});
        await act(() => result.current.acquire());
        expect(result.current.token).toBeNull();
    });

    it('stays without token when taking control fails', async () => {
        const {result} = setup({lock: vi.fn(async () => Promise.reject(new Error('taken')))});
        await expect(act(() => result.current.acquire())).rejects.toThrow('taken');
        expect(result.current.token).toBeNull();
    });
});

describe('useDeviceLease giving back', () => {
    it('gives the device back under its token', async () => {
        storeLease(7, 'tok-9');
        const {service, result} = setup({lockStatus: vi.fn(async () => heldBy({token: 'tok-9'}))});
        await tick();
        expect(result.current.token).toBe('tok-9');
        await act(() => result.current.release());
        expect(service.release).toHaveBeenCalledWith('tok-9');
        expect(result.current.token).toBeNull();
        expect(loadLease(7)).toBeNull();
    });

    it('has nothing to give back without a token', async () => {
        const {service, result} = setup();
        await act(() => result.current.release());
        expect(service.release).not.toHaveBeenCalled();
    });

    it('keeps the token when giving back fails', async () => {
        storeLease(7, 'tok-9');
        const {result} = setup({release: vi.fn(async () => Promise.reject(new Error('no')))});
        await expect(act(() => result.current.release())).rejects.toThrow('no');
        expect(result.current.token).toBe('tok-9');
    });
});

describe('useDeviceLease the view of the hub', () => {
    it('drops a token the hub no longer knows', async () => {
        storeLease(7, 'tok-9');
        const {result} = setup();
        await tick();
        expect(result.current.token).toBeNull();
        expect(loadLease(7)).toBeNull();
    });

    it('survives a failing status read', async () => {
        const {result} = setup({lockStatus: vi.fn(async () => Promise.reject(new Error('down')))});
        await tick(LEASE_REFRESH_MS);
        expect(result.current.status).toBeNull();
    });

    it('stops reading after unmount', async () => {
        const {service, unmount} = setup();
        await tick();
        unmount();
        await tick(LEASE_REFRESH_MS * 3);
        expect(service.lockStatus).toHaveBeenCalledTimes(1);
    });
});
