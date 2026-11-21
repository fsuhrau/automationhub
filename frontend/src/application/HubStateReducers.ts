import type {HubState} from "./HubState";
import type {
    AppAttributePayload,
    AppParameterPayload,
    DeviceStatePayload,
    ProjectAttributePayload,
} from "./HubStateActions";
import {IAppData} from "../types/app";
import IProject from "../project/project";
import IDeviceData from "../types/device";

// Pure state transitions, one per action. Lists that were never loaded
// (null) become a list holding just the change, or stay empty for changes
// to an item that can't exist yet.

export const append = <T, >(list: T[] | null, item: T): T[] => list ? [...list, item] : [item];

const mapApp = (state: HubState, appId: number, change: (app: IAppData) => IAppData): HubState => ({
    ...state,
    apps: state.apps ? state.apps.map(a => a.id === appId ? change(a) : a) : [],
});

const mapDevice = (state: HubState, deviceId: number, change: (device: IDeviceData) => IDeviceData): HubState => ({
    ...state,
    devices: state.devices ? state.devices.map(d => d.id === deviceId ? change(d) : d) : [],
});

export const appAdd = (state: HubState, app: IAppData): HubState =>
    ({...state, apps: append(state.apps, app)});

export const appAttributeUpdate = (state: HubState, {appId, attribute, value}: AppAttributePayload): HubState =>
    mapApp(state, appId, a => ({...a, [attribute]: value}));

// a new app has no parameter list yet
export const addAppParameter = (state: HubState, parameter: AppParameterPayload): HubState =>
    mapApp(state, parameter.appId, a => ({...a, parameter: [...(a.parameter ?? []), parameter]}));

export const updateAppParameter = (state: HubState, parameter: AppParameterPayload): HubState =>
    mapApp(state, parameter.appId, a => ({
        ...a,
        parameter: (a.parameter ?? []).map(p => p.id === parameter.id ? parameter : p),
    }));

export const deleteAppParameter = (state: HubState, {appId, id}: { appId: number, id: number }): HubState =>
    mapApp(state, appId, a => ({...a, parameter: (a.parameter ?? []).filter(p => p.id !== id)}));

export const projectAttributeUpdate = (
    state: HubState,
    {projectIdentifier, attribute, value}: ProjectAttributePayload,
): HubState => ({
    ...state,
    projects: (state.projects ?? []).map((p): IProject =>
        p.identifier === projectIdentifier ? {...p, [attribute]: value} : p),
});

export const accessTokenDelete = (state: HubState, id: number): HubState =>
    ({...state, accessTokens: state.accessTokens ? state.accessTokens.filter(t => t.id !== id) : []});

export const nodeDelete = (state: HubState, id: number): HubState =>
    ({...state, nodes: state.nodes ? state.nodes.filter(n => n.id !== id) : []});

export const updateDevice = (state: HubState, device: IDeviceData): HubState =>
    mapDevice(state, device.id, () => device);

export const updateDeviceState = (state: HubState, {deviceId, deviceState}: DeviceStatePayload): HubState =>
    mapDevice(state, deviceId, d => ({...d, status: deviceState}));

export const unlockDevice = (state: HubState, deviceId: number): HubState =>
    mapDevice(state, deviceId, d => ({...d, isLocked: false, lockedBy: undefined}));


