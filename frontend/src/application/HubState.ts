import IProject from "../project/project";
import {IAppData} from "../types/app";
import IHubStatsData from "../types/hub.stats";
import IAccessTokenData from "../types/access.token";
import {INodeData} from "../types/node";
import {IUserData} from "../types/user";
import IDeviceData from "../types/device";

export enum HubStateActions {
    ProjectsUpdate = 'ProjectsUpdate',
    ToggleMobileOpen = 'ToggleMobileOpen',
    ProjectAttributeUpdate = 'ProjectAttributeUpdate',
    AppsUpdate = 'AppsUpdate',
    AppAdd = 'AppAdd',
    AppAttributeUpdate = 'AppAttributeUpdate',
    AccessTokensUpdate = 'AccessTokensUpdate',
    AccessTokenAdd = 'AccessTokenAdd',
    AccessTokenDelete = 'AccessTokenDelete',
    NodesUpdate = 'NodesUpdate',
    NodeAdd = 'NodeAdd',
    NodeDelete = 'NodeDelete',
    UsersUpdate = 'UsersUpdate',
    UpdateDevices = 'UpdateDevices',
    UpdateDevice = 'UpdateDevice',
    UpdateDeviceState = "UpdateDeviceState",
}

export interface HubStateAction {
    type: HubStateActions,
    payload?: any,
}

export interface HubState {
    mobileOpen: boolean,
    projects: IProject[] | null,
    stats: IHubStatsData | null,
    accessTokens: IAccessTokenData[] | null,
    nodes: INodeData[] | null,
    apps: IAppData[] | null,
    users: IUserData[] | null,
    devices: IDeviceData[] | null,
}

export const InitialHubState: HubState = {
    mobileOpen: false,
    projects: null,
    stats: null,
    accessTokens: null,
    nodes: null,
    apps: null,
    users: null,
    devices: null,
}

export function hubReducer(state: HubState, action: HubStateAction): HubState {

    const {type, payload} = action;

    switch (type) {
        case HubStateActions.UsersUpdate: {
            return {
                ...state,
                users: payload,
            }
        }
        case HubStateActions.AppsUpdate: {
            return {
                ...state,
                apps: payload,
            }
        }
        case HubStateActions.AppAdd: {
            return {
                ...state,
                apps: state.apps ? [...state.apps, payload] : [payload],
            }
        }
        case HubStateActions.AppAttributeUpdate: {
            return {
                ...state,
                apps: state.apps ? state.apps.map(a => (a.id as number) === (payload.appId as number) ? {
                    ...a,
                    [payload.attribute]: payload.value
                } as IAppData : a as IAppData) : [],
            }
        }
        case HubStateActions.ProjectAttributeUpdate: {
            return {
                ...state,
                projects: state.projects!.map(p => {
                    if (p.identifier == payload.projectIdentifier) {
                        return {
                            ...p,
                            [payload.attribute]: payload.value
                        }
                    }
                    return p
                })
            }
        }

        case HubStateActions.ProjectsUpdate: {
            return {
                ...state,
                projects: payload,
            }
        }
        case HubStateActions.ToggleMobileOpen:
            return {
                ...state,
                mobileOpen: !state.mobileOpen,
            }
        case HubStateActions.AccessTokensUpdate: {
            return {
                ...state,
                accessTokens: payload,
            }
        }
        case HubStateActions.AccessTokenAdd: {
            return {
                ...state,
                accessTokens: state.accessTokens ? [...state.accessTokens, payload] : [payload],
            }
        }
        case HubStateActions.AccessTokenDelete: {
            return {
                ...state,
                accessTokens: state.accessTokens ? state.accessTokens.filter(token => token.id !== payload) : [],
            }
        }
        case HubStateActions.NodesUpdate: {
            return {
                ...state,
                nodes: payload,
            }
        }
        case HubStateActions.NodeAdd: {
            return {
                ...state,
                nodes: state.nodes ? [...state.nodes, payload] : [payload],
            }
        }
        case HubStateActions.NodeDelete: {
            return {
                ...state,
                nodes: state.nodes ? state.nodes.filter(node => node.id !== payload) : [],
            }
        }
        case HubStateActions.UpdateDevices: {
            return {
                ...state,
                devices: payload,
            }
        }
        case HubStateActions.UpdateDevice: {
            return {
                ...state,
                devices: state.devices ? state.devices?.map(d => d.id === payload.id ? payload : d) : [],
            }
        }
        case HubStateActions.UpdateDeviceState: {
            return {
                ...state,
                devices: state.devices ? [...state.devices.map(d => d.id === payload.deviceId ? {
                    ...d,
                    status: payload.deviceState
                } : d)] : []
            }
        }
        default:
            return state
    }
}
