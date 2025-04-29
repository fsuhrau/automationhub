import { AxiosResponse } from 'axios';
import http from '../http-common';
import {AppParameter, AppParameterType, IAppBinaryData, IAppData, Parameter} from "../types/app";

export const getAllApps = (projectId: string): Promise<IAppData[]> => {
    return http.get(`/${projectId}/apps`).then(resp => resp.data)
};

export const getApp = (projectId: string, id: string): Promise<IAppData | undefined> => {
    return http.get(`/${projectId}/app/${id}`).then(resp => resp.data)
};

export const createApp = (projectId: string, data: IAppData): Promise<IAppData> => {
    return http.post(`/${projectId}/app`, data).then(resp => resp.data)
};

export const updateApp = (projectId: string, appId: number, data: IAppData): Promise<IAppData> => {
    return http.put(`/${projectId}/app/${appId}`, data).then(resp => resp.data)
};

export const addAppParameter = (projectId: string, appId: number, data: AppParameter): Promise<AppParameter> => {
    return http.post(`/${projectId}/app/${appId}/parameter`, data).then(resp => resp.data)
};

export const updateAppParameter = (projectId: string, appId: number, id: number, data: AppParameter): Promise<AppParameter> => {
    return http.put(`/${projectId}/app/${appId}/parameter/${id}`, data).then(resp => resp.data)
};

export const removeAppParameter = (projectId: string, appId: number, id: number): Promise<void> => {
    return http.delete(`/${projectId}/app/${appId}/parameter/${id}`).then(resp => resp.data)
};

export const updateAppBundle = (projectId: string, appId: number, id: number, data: IAppBinaryData): Promise<IAppBinaryData> => {
    return http.put(`/${projectId}/app/${appId}/bundle/${id}`, data).then(resp => resp.data)
};

export const deleteAppBundle = (projectId: string, appId: number, id: number): Promise<void> => {
    return http.delete(`/${projectId}/app/${appId}/bundle/${id}`);
};

export const getAppBundles = (projectId: string, appId: number): Promise<IAppBinaryData[]> => {
    return http.get(`/${projectId}/app/${appId}/bundles`).then(resp => resp.data)
};

export type AppFilter = {
    name?: string;
};

export const findApp = (filter?: AppFilter): Promise<IAppBinaryData[]> => {
    return http.get('/apps', {params: filter}).then(resp => resp.data)
};

export const uploadNewApp = (file: File,
    projectId: string,
    appId: number,
    uploadProgress: (progressEvent: number) => void,
    finished: (finished: AxiosResponse<IAppBinaryData>) => void): void => {

    const formData = new FormData();
    formData.append('test_target', file);

    http.request({
        method: 'post',
        url: `/${projectId}/app/${appId}/upload`,
        data: formData,
        headers: {
            'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: progressEvent => {
            if (progressEvent.total) {
                uploadProgress((progressEvent.loaded / progressEvent.total) * 100);
            } else return ''
        },
    }).then(data => {
        console.log(data);
        uploadProgress(100);
        finished(data);
    }).catch(ex => {
        alert(ex)
    });
};