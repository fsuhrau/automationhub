import {useCallback} from 'react';
import {useError} from '../../../ErrorProvider';
import {errorMessage} from '../../../lib/errorMessage';
import IDeviceData from '../../../types/device';
import {IAppData} from '../../../types/app';
import {DeviceControlApi, DeviceTarget} from '../../../types/device.control.api';
import {deviceTitle} from '../deviceControlModel';
import {ControlStatus, controlStatus} from './statusModel';
import {useDeviceControlService, useDeviceTarget} from './useDeviceControlService';
import {useDeviceData} from './useDeviceData';
import {DeviceLeaseState, useDeviceLease} from './useDeviceLease';
import {ErrorHandler, Runner, useGuardedAction} from './useGuardedAction';

export interface ControlSession {
    target: DeviceTarget,
    service: DeviceControlApi,
    device: IDeviceData | null,
    apps: IAppData[],
    name: string,
    lease: DeviceLeaseState,
    status: ControlStatus,
    // runs an action of the page; a failure is shown as an error and the
    // lock status is read again
    run: Runner,
    busy: boolean,
    report: ErrorHandler,
    onLock: () => void,
    onRelease: () => void,
}

// useControlSession is what the parts of the control page share: the device
// and its lease, and running actions on it.
export const useControlSession = (): ControlSession => {
    const target = useDeviceTarget();
    const service = useDeviceControlService(target);
    const {setError} = useError();
    const report = useCallback((e: unknown) => setError(errorMessage(e)), [setError]);
    const lease = useDeviceLease(service, target.deviceId);
    const {refresh} = lease;
    const {run, busy} = useGuardedAction(useCallback((e: unknown) => {
        report(e);
        refresh();
    }, [report, refresh]));
    const {device, apps} = useDeviceData(service, report);
    const status = controlStatus(lease.token !== null, lease.status, lease.now);
    return {
        target, service, device, apps, lease, status, run, busy, report,
        name: deviceTitle(device, target.deviceId),
        onLock: () => run(lease.acquire),
        onRelease: () => run(lease.release),
    };
};
