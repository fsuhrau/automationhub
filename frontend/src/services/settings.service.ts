import {useMemo} from 'react';
import {ApiClient, createApiClient} from '../api/ApiClient';
import {useApiClient} from '../api/ApiClientProvider';
import http from '../http-common';
import IAccessTokenData from '../types/access.token';
import {AppParameter, IAppData} from '../types/app';
import {INodeData} from '../types/node';
import {IUser} from '../types/user';
import {NewAccessTokenRequest, NewAppRequest, NewNodeRequest, NodeWithToken, ParameterRequest} from '../types/settings.requests';

export interface SettingsService {
    accessTokens: (projectId: string) => Promise<IAccessTokenData[]>,
    createAccessToken: (projectId: string, data: NewAccessTokenRequest) => Promise<IAccessTokenData>,
    deleteAccessToken: (projectId: string, id: number) => Promise<void>,
    nodes: (projectId: string) => Promise<INodeData[]>,
    createNode: (projectId: string, data: NewNodeRequest) => Promise<NodeWithToken>,
    // gives a node a new token; its old token stops working
    issueNodeToken: (projectId: string, id: number) => Promise<NodeWithToken>,
    deleteNode: (projectId: string, id: number) => Promise<void>,
    users: (projectId: string) => Promise<IUser[]>,
}

export const settingsService = (api: ApiClient): SettingsService => ({
    accessTokens: (projectId) => api.get(`/${projectId}/settings/access_tokens`),
    createAccessToken: (projectId, data) => api.post(`/${projectId}/settings/access_token`, data),
    deleteAccessToken: (projectId, id) => api.delete(`/${projectId}/settings/access_token/${id}`),
    nodes: (projectId) => api.get(`/${projectId}/settings/nodes`),
    createNode: (projectId, data) => api.post(`/${projectId}/settings/node`, data),
    issueNodeToken: (projectId, id) => api.post(`/${projectId}/settings/nodes/${id}/token`),
    deleteNode: (projectId, id) => api.delete(`/${projectId}/settings/nodes/${id}`),
    users: (projectId) => api.get(`/${projectId}/settings/users`),
});

// an app of a project
export interface AppScope {
    projectId: string,
    appId: number,
}

// the app calls of the settings page: the apps, their name and parameters
export interface AppsSettingsService {
    list: (projectId: string) => Promise<IAppData[]>,
    create: (projectId: string, data: NewAppRequest) => Promise<IAppData>,
    update: (app: AppScope, data: IAppData) => Promise<IAppData>,
    addParameter: (app: AppScope, data: ParameterRequest) => Promise<AppParameter>,
    updateParameter: (app: AppScope, id: number, data: ParameterRequest) => Promise<AppParameter>,
    removeParameter: (app: AppScope, id: number) => Promise<void>,
}

export const appsSettingsService = (api: ApiClient): AppsSettingsService => ({
    list: (projectId) => api.get(`/${projectId}/apps`),
    create: (projectId, data) => api.post(`/${projectId}/app`, data),
    update: ({projectId, appId}, data) => api.put(`/${projectId}/app/${appId}`, data),
    addParameter: ({projectId, appId}, data) => api.post(`/${projectId}/app/${appId}/parameter`, data),
    updateParameter: ({projectId, appId}, id, data) => api.put(`/${projectId}/app/${appId}/parameter/${id}`, data),
    removeParameter: ({projectId, appId}, id) => api.delete(`/${projectId}/app/${appId}/parameter/${id}`),
});

export const useSettingsService = (): SettingsService => {
    const api = useApiClient();
    return useMemo(() => settingsService(api), [api]);
};

export const useAppsSettingsService = (): AppsSettingsService => {
    const api = useApiClient();
    return useMemo(() => appsSettingsService(api), [api]);
};

/** @deprecated legacy wrapper, removed in F13 */
export const getNodes = (projectId: string): Promise<INodeData[]> => settingsService(createApiClient(http)).nodes(projectId);
