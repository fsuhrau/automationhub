import http from '../http-common';
import IAccessTokenData from '../types/access.token';
import {Dayjs} from "dayjs";
import {INodeData} from "../types/node";
import {IUser} from '../types/user';

export const getAccessTokens = (projectId: string): Promise<IAccessTokenData[]> => {
    return http.get(`/${projectId}/settings/access_tokens`).then(response => response.data);
};

export interface NewAccessTokenRequest {
    name: string,
    expiresAt: Dayjs | null,
}

export const createAccessToken = (projectId: string, data: NewAccessTokenRequest): Promise<IAccessTokenData> => {
    return http.post(`/${projectId}/settings/access_token`, data).then(response => response.data);
};

export const deleteAccessToken = (projectId: string, id: number): Promise<void> => {
    return http.delete(`/${projectId}/settings/access_token/${id}`);
};

export const getNodes = (projectId: string): Promise<INodeData[]> => {
    return http.get(`/${projectId}/settings/nodes`).then(response => response.data)
};

export interface NewNodeRequest {
    name: string,
}

export const createNode = (projectId: string, data: NewNodeRequest): Promise<INodeData> => {
    return http.post(`/${projectId}/settings/node`, data).then(response => response.data);
};

export const getNode = (projectId: string, id: number): Promise<INodeData> => {
    return http.get(`/${projectId}/settings/nodes/${id}`).then(response => response.data)
};

export const deleteNode = (projectId: string, id: number): Promise<void> => {
    return http.delete(`/${projectId}/settings/nodes/${id}`);
};

export const getUsers = (projectId: string): Promise<IUser[]> => {
    return http.get(`/${projectId}/settings/users`).then(response => response.data)
};