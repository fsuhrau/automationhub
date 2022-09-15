import http from '../http-common';
import { AxiosResponse } from 'axios';
import IAccessTokenData from '../types/access.token';
import { Dayjs } from "dayjs";


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
