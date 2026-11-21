// What the device page says about whether a device can be used now.
import IDeviceData from '../../types/device';
import ITestProtocolData from '../../types/test.protocol';
import IConnectionParameter from '../../types/device.connection.parameter';
import {INodeData} from '../../types/node';
import {DeviceStateType} from '../../types/deviceStateType';
import {DeviceType} from '../../types/device.type.enum';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {PlatformType} from '../../types/platform.type.enum';
import {routes} from '../../lib/routes';
import {DeviceControllerInfo} from '../../components/DeviceController';

export const DEVICE_TYPES: Record<DeviceType, string> = {
    [DeviceType.Phone]: 'Phone',
    [DeviceType.Tablet]: 'Tablet',
    [DeviceType.Desktop]: 'Desktop',
    [DeviceType.UnityEditor]: 'Unity editor',
    [DeviceType.Browser]: 'Browser',
};

export const CONNECTIONS: Record<DeviceConnectionType, string> = {
    [DeviceConnectionType.USB]: 'USB',
    [DeviceConnectionType.Remote]: 'Network',
    [DeviceConnectionType.HubNode]: 'Through a node',
};

// the device list follows live updates, the loaded device has the details:
// the stored device with the live status and lease of its list entry
export const mergeLiveDevice = (stored: IDeviceData, live?: IDeviceData): IDeviceData =>
    live ? {...stored, status: live.status, isLocked: live.isLocked, lockedBy: live.lockedBy} : stored;

// the node of a device as the hub knows it now, else as the device was loaded
export const resolveNode = (device: IDeviceData, nodes: INodeData[] | null): INodeData | undefined =>
    nodes?.find(n => n.id === device.nodeId) ?? device.node;

// logcat works whenever the phone is up, also while a test run holds it
export const showsLog = (device: IDeviceData): boolean =>
    device.platformType === PlatformType.Android
    && (device.status === DeviceStateType.Booted || device.status === DeviceStateType.Locked);

// ip and port of a network device, null for other connections
export const connectionAddress = (connection?: IConnectionParameter): string | null => {
    if (connection?.connectionType !== DeviceConnectionType.Remote || !connection.ip) return null;
    return connection.port ? `${connection.ip}:${connection.port}` : connection.ip;
};

export const alertSeverity = (tone: Availability['tone']): 'error' | 'warning' | 'info' =>
    tone === 'error' || tone === 'warning' ? tone : 'info';

// the page of a test result, null while the result has no test
export const recentTestLink = (projectId: string, p: ITestProtocolData): string | null => {
    const test = p.testRun?.test;
    if (!test) return null;
    return routes.project(projectId).app(test.appId).test(p.testRun.testId).run(p.testRunId).protocol(p.id as number);
};

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

const IN_USE: Availability = {
    tone: 'warning', label: 'In use by a test run', canControl: false, canUnlock: true,
    detail: 'A test run holds the device. Unlock it only when the run is stuck.',
};

const UNKNOWN: Availability = {
    tone: 'default', label: 'Unknown', canControl: false, canUnlock: false, detail: "The device's state isn't known yet.",
};

// what the state of a device that nobody holds says
const BY_STATUS: Partial<Record<DeviceStateType, Availability>> = {
    [DeviceStateType.Booted]: {
        tone: 'success', label: 'Ready', canControl: true, canUnlock: false, detail: 'Free for test runs and live control.',
    },
    [DeviceStateType.Unauthorized]: {
        tone: 'warning', label: 'Not authorized', canControl: false, canUnlock: false,
        detail: "The device doesn't trust the computer it is connected to yet.",
    },
    [DeviceStateType.RemoteDisconnected]: {
        tone: 'error', label: 'Disconnected', canControl: false, canUnlock: false, detail: "The node doesn't see the device.",
    },
    [DeviceStateType.NodeDisconnected]: {
        tone: 'error', label: 'Node offline', canControl: false, canUnlock: false,
        detail: "The node this device is connected to isn't connected to the master.",
    },
    [DeviceStateType.Shutdown]: {
        tone: 'default', label: 'Shut down', canControl: false, canUnlock: false,
        detail: 'The device (or simulator, editor, browser) is not running.',
    },
};

const isUp = (status: DeviceStateType): boolean => status === DeviceStateType.Booted || status === DeviceStateType.Locked;

type Describe = (owner: string) => DeviceControllerInfo;

// only a person's own lease is "you": agents, tokens and anonymous never are
const isYou = (owner: DeviceControllerInfo, meId?: number): boolean =>
    meId !== undefined && !owner.agent && owner.userId === meId;

const controlled = (d: DeviceLiveState & { lockedBy: string }, describe: Describe, meId?: number): Availability => {
    const control = {canControl: isUp(d.status), canUnlock: true};
    const owner = describe(d.lockedBy);
    if (isYou(owner, meId)) {
        return {
            ...control, tone: 'success', label: 'Controlled by you',
            detail: "You control the device live. Release it on the live control page when you're done.",
        };
    }
    return {
        ...control, tone: 'warning', label: `Controlled by ${owner.label}`,
        detail: 'Someone controls the device live. Tests wait until the control ends.',
    };
};

// availability turns the device's live state into one answer; describe
// parses a lease owner (a person, a token or an agent), meId is the
// signed-in user's id
export const availability = (d: DeviceLiveState, describe: Describe, meId?: number): Availability => {
    if (d.lockedBy) return controlled({...d, lockedBy: d.lockedBy}, describe, meId);
    if (d.isLocked || d.status === DeviceStateType.Locked) return IN_USE;
    return BY_STATUS[d.status] ?? UNKNOWN;
};
