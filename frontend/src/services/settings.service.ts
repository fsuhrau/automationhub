import http from '../http-common';
import { AxiosResponse } from 'axios';
import IAccessTokenData from '../types/access.token';
import { Dayjs } from "dayjs";
import { INodeData } from "../types/node";


export const getAccessTokens = (projectId: string): Promise<AxiosResponse<IAccessTokenData[]>> => {
    return http.get(`/${projectId}/settings/access_tokens`);
};

export interface NewAccessTokenRequest {
    Name: string,
    ExpiresAt: Dayjs | null,
}

export const createAccessToken = (projectId: string, data: NewAccessTokenRequest): Promise<AxiosResponse<IAccessTokenData>> => {
    return http.post(`/${projectId}/settings/access_token`, data);
};

export const deleteAccessToken = (projectId: string, id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/${projectId}/settings/access_token/${ id }`);
};


export const getNodes = (projectId: string): Promise<AxiosResponse<INodeData[]>> => {
    return http.get(`/${projectId}/settings/nodes`);
};

export interface NewNodeRequest {
    Name: string,
}

export const createNode = (projectId: string, data: NewNodeRequest): Promise<AxiosResponse<INodeData>> => {
    return http.post(`/${projectId}/settings/node`, data);
};

export const getNode = (projectId: string, id: number): Promise<AxiosResponse<void>> => {
    return http.get(`/${projectId}/settings/nodes/${ id }`);
};

export const deleteNode = (projectId: string, id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/${projectId}/settings/nodes/${ id }`);
};

