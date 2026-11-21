import IDeviceData from './device';
import {IAppBinaryData, IAppData} from './app';
import {DeviceLease, DeviceLockStatus} from './device.control.lease';

// the device a control page works on
export interface DeviceTarget {
    projectId: string,
    deviceId: number,
}

export interface Point {
    x: number,
    y: number,
}

export interface Swipe {
    from: Point,
    to: Point,
    durationMs: number,
}

export interface ElementRef {
    elementId?: string,
    xpath?: string,
}

export type AppAction = 'install' | 'start' | 'stop';

export interface AppRequest {
    binaryId: number,
    clearData?: boolean,
}

// What the control page needs from the hub for one device. Calls with a
// token act on the device under that lease.
export interface DeviceControlApi {
    device(): Promise<IDeviceData | undefined>,

    apps(): Promise<IAppData[]>,

    bundles(appId: number): Promise<IAppBinaryData[]>,

    lockStatus(): Promise<DeviceLockStatus>,

    lock(): Promise<DeviceLease>,

    release(token: string): Promise<void>,

    tap(token: string, at: Point): Promise<void>,

    swipe(token: string, swipe: Swipe): Promise<void>,

    text(token: string, text: string): Promise<void>,

    key(token: string, key: string): Promise<void>,

    touchElement(token: string, ref: ElementRef): Promise<void>,

    hierarchy(): Promise<string>,

    logs(lines: number): Promise<string[]>,

    appAction(token: string, action: AppAction, request: AppRequest): Promise<void>,
}
