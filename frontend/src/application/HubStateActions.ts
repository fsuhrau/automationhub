import IProject from "../project/project";
import {AppParameter, IAppData} from "../types/app";
import IAccessTokenData from "../types/access.token";
import {INodeData} from "../types/node";
import {IUserData} from "../types/user";
import IDeviceData from "../types/device";
import {DeviceStateType} from "../types/deviceStateType";

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
    UnlockDevice = "UnlockDevice",
    AddAppParameter = 'AddAppParameter',
    UpdateAppParameter = 'UpdateAppParameter',
    DeleteAppParameter = 'DeleteAppParameter',
}

// an app parameter as the reducer stores it: tagged with its app
export type AppParameterPayload = AppParameter & { appId: number };

export interface ProjectAttributePayload {
    projectIdentifier: string,
    attribute: keyof IProject,
    value: unknown,
}

export interface AppAttributePayload {
    appId: number,
    attribute: keyof IAppData,
    value: unknown,
}

export interface DeviceStatePayload {
    deviceId: number,
    deviceState: DeviceStateType,
}

export type HubStateAction =
    | { type: HubStateActions.ProjectsUpdate, payload: IProject[] }
    | { type: HubStateActions.ToggleMobileOpen, payload?: undefined }
    | { type: HubStateActions.ProjectAttributeUpdate, payload: ProjectAttributePayload }
    | { type: HubStateActions.AppsUpdate, payload: IAppData[] }
    | { type: HubStateActions.AppAdd, payload: IAppData }
    | { type: HubStateActions.AppAttributeUpdate, payload: AppAttributePayload }
    | { type: HubStateActions.AccessTokensUpdate, payload: IAccessTokenData[] }
    | { type: HubStateActions.AccessTokenAdd, payload: IAccessTokenData }
    | { type: HubStateActions.AccessTokenDelete, payload: number }
    | { type: HubStateActions.NodesUpdate, payload: INodeData[] }
    | { type: HubStateActions.NodeAdd, payload: INodeData }
    | { type: HubStateActions.NodeDelete, payload: number }
    | { type: HubStateActions.UsersUpdate, payload: IUserData[] }
    | { type: HubStateActions.UpdateDevices, payload: IDeviceData[] }
    | { type: HubStateActions.UpdateDevice, payload: IDeviceData }
    | { type: HubStateActions.UpdateDeviceState, payload: DeviceStatePayload }
    | { type: HubStateActions.UnlockDevice, payload: { deviceId: number } }
    | { type: HubStateActions.AddAppParameter, payload: AppParameterPayload }
    | { type: HubStateActions.UpdateAppParameter, payload: AppParameterPayload }
    | { type: HubStateActions.DeleteAppParameter, payload: { appId: number, id: number } };
