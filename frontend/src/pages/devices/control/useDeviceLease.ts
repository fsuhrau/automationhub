import {useCallback, useEffect, useState} from 'react';
import {DeviceControlApi} from '../../../types/device.control.api';
import {DeviceLockStatus} from '../../../types/device.control.lease';
import {useNow} from '../../../hooks/useNow';
import {loadLease, storeLease} from './leaseStorage';

// how often the lock status is read; it keeps the lease alive and shows
// when it ends
export const LEASE_REFRESH_MS = 5000;

export interface DeviceLeaseState {
    // the token of the lease this browser holds, or null
    token: string | null,
    // what the hub says about the device's lock, null until the first answer
    status: DeviceLockStatus | null,
    // the time of the last tick, for "ends in N minutes"; it ticks as often
    // as the status is read, so useNow replaces a timestamp after each read
    now: number,
    refresh: () => void,
    // take control; rejects when the hub refuses
    acquire: () => Promise<void>,
    // give the device back; rejects when the hub refuses
    release: () => Promise<void>,
}

// useDeviceLease holds the lease of a device: takes and gives it back,
// reads the lock status every few seconds and drops the token when the hub
// no longer knows the lease (it expired or was released elsewhere).
export const useDeviceLease = (service: DeviceControlApi, deviceId: number): DeviceLeaseState => {
    const [token, setToken] = useState<string | null>(() => loadLease(deviceId));
    const [status, setStatus] = useState<DeviceLockStatus | null>(null);
    const now = useNow(LEASE_REFRESH_MS);
    const remember = useCallback((value: string | null) => {
        setToken(value);
        storeLease(deviceId, value);
    }, [deviceId]);

    const refresh = useCallback(() => {
        service.lockStatus().then(next => {
            setStatus(next);
            if (!next.lease && token) remember(null);
        }).catch(() => undefined);
    }, [service, token, remember]);

    useEffect(() => {
        refresh();
        const timer = setInterval(refresh, LEASE_REFRESH_MS);
        return () => clearInterval(timer);
    }, [refresh]);

    const acquire = async (): Promise<void> => {
        remember((await service.lock()).token ?? null);
        refresh();
    };
    const release = async (): Promise<void> => {
        if (token) await service.release(token);
        remember(null);
        refresh();
    };
    return {token, status, now, refresh, acquire, release};
};
