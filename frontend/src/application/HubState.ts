import IProject from "../project/project";
import {IAppData} from "../types/app";
import IAccessTokenData from "../types/access.token";
import {INodeData} from "../types/node";
import {IUserData} from "../types/user";
import IDeviceData from "../types/device";
import {HubStateAction, HubStateActions} from "./HubStateActions";
import * as r from "./HubStateReducers";

export {HubStateActions};
export type {HubStateAction};

export interface HubState {
    mobileOpen: boolean,
    projects: IProject[] | null,
    accessTokens: IAccessTokenData[] | null,
    nodes: INodeData[] | null,
    apps: IAppData[] | null,
    users: IUserData[] | null,
    devices: IDeviceData[] | null,
}

export const InitialHubState: HubState = {
    mobileOpen: false,
    projects: null,
    accessTokens: null,
    nodes: null,
    apps: null,
    users: null,
    devices: null,
}

// one pure transition per action type; the mapped type makes the compiler
// require an entry, with the matching payload, for every action
type Handlers = {
    [K in HubStateActions]: (state: HubState, payload: Extract<HubStateAction, { type: K }>['payload']) => HubState
};

const handlers: Handlers = {
    [HubStateActions.UsersUpdate]: (state, users) => ({...state, users}),
    [HubStateActions.AppsUpdate]: (state, apps) => ({...state, apps}),
    [HubStateActions.AppAdd]: r.appAdd,
    [HubStateActions.AppAttributeUpdate]: r.appAttributeUpdate,
    [HubStateActions.AddAppParameter]: r.addAppParameter,
    [HubStateActions.UpdateAppParameter]: r.updateAppParameter,
    [HubStateActions.DeleteAppParameter]: r.deleteAppParameter,
    [HubStateActions.ProjectAttributeUpdate]: r.projectAttributeUpdate,
    [HubStateActions.ProjectsUpdate]: (state, projects) => ({...state, projects}),
    [HubStateActions.ToggleMobileOpen]: (state) => ({...state, mobileOpen: !state.mobileOpen}),
    [HubStateActions.AccessTokensUpdate]: (state, accessTokens) => ({...state, accessTokens}),
    [HubStateActions.AccessTokenAdd]: (state, token) => ({...state, accessTokens: r.append(state.accessTokens, token)}),
    [HubStateActions.AccessTokenDelete]: r.accessTokenDelete,
    [HubStateActions.NodesUpdate]: (state, nodes) => ({...state, nodes}),
    [HubStateActions.NodeAdd]: (state, node) => ({...state, nodes: r.append(state.nodes, node)}),
    [HubStateActions.NodeDelete]: r.nodeDelete,
    [HubStateActions.UpdateDevices]: (state, devices) => ({...state, devices}),
    [HubStateActions.UpdateDevice]: r.updateDevice,
    [HubStateActions.UpdateDeviceState]: r.updateDeviceState,
    [HubStateActions.UnlockDevice]: (state, {deviceId}) => r.unlockDevice(state, deviceId),
};

type AnyHandler = (state: HubState, payload: unknown) => HubState | undefined;

export function hubReducer(state: HubState, action: HubStateAction): HubState {
    // the lookup may miss for a type that doesn't exist at runtime
    const handler = handlers[action.type] as AnyHandler | undefined;
    return handler?.(state, action.payload) ?? state;
}
