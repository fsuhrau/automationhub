import http from '../http-common';
import { AxiosResponse } from 'axios';
import { IAppBinaryData, IAppData } from "../types/app";

export const getAllApps = (projectId: string): Promise<AxiosResponse<IAppBinaryData[]>> => {
    return http.get(`/${projectId}/apps`);
};

export const getApp = (projectId: string, id: string): Promise<AxiosResponse<IAppBinaryData | undefined>> => {
    return http.get(`/${projectId}/app/${id}`);
};

export const createApp = (projectId: string, data: IAppData): Promise<AxiosResponse<IAppData>> => {
    return http.post(`/${projectId}/app`, data);
};

export const updateApp = (projectId: string, appId: number, data: IAppData): Promise<AxiosResponse<IAppData>> => {
    return http.put(`/${projectId}/app/${appId}`, data);
};

export const updateAppBundle = (projectId: string, appId: number, id: number, data: IAppBinaryData): Promise<AxiosResponse<IAppBinaryData>> => {
    return http.put(`/${projectId}/app/${appId}/bundle/${id}`, data);
};

export const deleteAppBundle = (projectId: string, appId: number, id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/${projectId}/app/${appId}/bundle/${id}`);
};

export const getAppBundles = (projectId: string, appId: number): Promise<AxiosResponse<IAppBinaryData[]>> => {
    return http.get(`/${projectId}/app/${appId}/bundles`);
};

export type AppFilter = {
    name?: string;
};

export const findApp = (filter?: AppFilter): Promise<AxiosResponse<IAppBinaryData[]>> => {
    return http.get('/apps', { params: filter });
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