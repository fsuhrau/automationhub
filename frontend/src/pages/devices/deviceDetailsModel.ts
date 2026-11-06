// What the device page says about whether a device can be used now.
import {DeviceStateType} from '../../types/deviceStateType';

export interface Availability {
    tone: 'success' | 'warning' | 'error' | 'default',
    label: string,
    detail: string,
    // live control makes sense (the device is up)
    canControl: boolean,
    // someone or something holds the device and an unlock would free it
    canUnlock: boolean,
}

export interface DeviceLiveState {
    status: DeviceStateType,
    isLocked: boolean,
    lockedBy?: string,
}

// availability turns the device's live state into one answer; describe
// names a lease owner (a person, a token or an agent), me is the signed-in
// user's name
export const availability = (d: DeviceLiveState, describe: (owner: string) => string, me?: string): Availability => {
    const up = d.status === DeviceStateType.Booted || d.status === DeviceStateType.Locked;
    if (d.lockedBy && me && describe(d.lockedBy) === me) {
        return {tone: 'success', label: 'Controlled by you', canControl: up, canUnlock: true,
            detail: "You control the device live. Release it on the live control page when you're done."};
    }
    if (d.lockedBy) {
        return {tone: 'warning', label: `Controlled by ${describe(d.lockedBy)}`, canControl: up, canUnlock: true,
            detail: 'Someone controls the device live. Tests wait until the control ends.'};
    }
    if (d.isLocked || d.status === DeviceStateType.Locked) {
        return {tone: 'warning', label: 'In use by a test run', canControl: false, canUnlock: true,
            detail: 'A test run holds the device. Unlock it only when the run is stuck.'};
    }
    switch (d.status) {
        case DeviceStateType.Booted:
            return {tone: 'success', label: 'Ready', canControl: true, canUnlock: false,
                detail: 'Free for test runs and live control.'};
        case DeviceStateType.Unauthorized:
            return {tone: 'warning', label: 'Not authorized', canControl: false, canUnlock: false,
                detail: "The device doesn't trust the computer it is connected to yet."};
        case DeviceStateType.RemoteDisconnected:
            return {tone: 'error', label: 'Disconnected', canControl: false, canUnlock: false,
                detail: "The node doesn't see the device."};
        case DeviceStateType.NodeDisconnected:
            return {tone: 'error', label: 'Node offline', canControl: false, canUnlock: false,
                detail: "The node this device is connected to isn't connected to the master."};
        case DeviceStateType.Shutdown:
            return {tone: 'default', label: 'Shut down', canControl: false, canUnlock: false,
                detail: 'The device (or simulator, editor, browser) is not running.'};
        default:
            return {tone: 'default', label: 'Unknown', canControl: false, canUnlock: false,
                detail: "The device's state isn't known yet."};
    }
};
