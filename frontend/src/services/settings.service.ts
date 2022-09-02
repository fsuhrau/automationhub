import http from '../http-common';
import { AxiosResponse } from 'axios';
import IAccessTokenData from '../types/access.token';
import { Dayjs } from "dayjs";


export const getAccessTokens = (): Promise<AxiosResponse<IAccessTokenData[]>> => {
    return http.get('/settings/access_tokens');
};

export interface NewAccessTokenRequest {
    Name: string,
    ExpiresAt: Dayjs | null,
}

export const createAccessToken = (data: NewAccessTokenRequest): Promise<AxiosResponse<IAccessTokenData>> => {
    return http.post('/settings/access_token', data);
};

export const deleteAccessToken = (id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/settings/access_token/${ id }`);
};
