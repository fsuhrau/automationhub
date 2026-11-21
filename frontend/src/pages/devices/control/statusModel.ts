// Who controls the device, as the page's status chip shows it.
import {DeviceLockStatus} from '../../../types/device.control.lease';
import {describeController} from '../../../components/DeviceController';
import {formatDateTime} from '../../../components/DateTime';
import {leaseNote} from '../deviceControlModel';

export interface ControlStatus {
    color: 'success' | 'warning' | 'default',
    label: string,
    note: string | null,
    // this browser holds the lease
    own: boolean,
    // somebody else holds it, or a test run uses the device
    taken: boolean,
}

// controlStatus describes who controls a device: this browser (it has a
// lease token and the hub still knows the lease), somebody else, a test run,
// or nobody
export const controlStatus = (hasToken: boolean, lock: DeviceLockStatus | null, now: number): ControlStatus => {
    const lease = lock?.lease;
    if (hasToken && lease) {
        return {color: 'success', label: 'You control this device', note: leaseNote(lease.expiresAt, now), own: true, taken: false};
    }
    if (lease?.owner) {
        const note = lease.acquiredAt ? `since ${formatDateTime(lease.acquiredAt, 'HH:mm')}` : null;
        return {color: 'warning', label: `Controlled by ${describeController(lease.owner).label}`, note, own: false, taken: true};
    }
    if (lock?.lockedByTest) {
        return {color: 'warning', label: 'A test run is using this device', note: 'Control is possible when it ends.', own: false, taken: true};
    }
    return {color: 'default', label: 'Free', note: 'Take control to use the device.', own: false, taken: false};
};
