import {useMemo} from 'react';
import {ApiClient} from '../api/ApiClient';
import {useApiClient} from '../api/ApiClientProvider';
import IAccessTokenData from '../types/access.token';
import {INodeData} from '../types/node';
import {IUser} from '../types/user';
import {NewAccessTokenRequest, NewNodeRequest, NodeWithToken} from '../types/settings.requests';

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

export const useSettingsService = (): SettingsService => {
    const api = useApiClient();
    return useMemo(() => settingsService(api), [api]);
};
