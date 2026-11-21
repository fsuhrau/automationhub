import {useEffect} from 'react';
import IDeviceData from '../../../types/device';
import {IAppData} from '../../../types/app';
import {DeviceControlApi} from '../../../types/device.control.api';
import {useLoader} from '../../../hooks/useLoader';
import {ErrorHandler} from './useGuardedAction';

export interface DeviceData {
    device: IDeviceData | null,
    apps: IAppData[],
}

// useDeviceData loads the device and the apps of the project; a failure goes
// to report.
export const useDeviceData = (service: DeviceControlApi, report: ErrorHandler): DeviceData => {
    const device = useLoader(() => service.device().then(d => d ?? null), [service]);
    const apps = useLoader(() => service.apps(), [service]);
    const error = device.error ?? apps.error;
    useEffect(() => {
        if (error) report(error);
    }, [error, report]);
    return {device: device.data, apps: apps.data ?? []};
};
