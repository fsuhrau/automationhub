import http from '../http-common';
import { AxiosResponse } from 'axios';
import { IAppBinaryData } from "../types/app";

export const getAllApps = (projectId: string): Promise<AxiosResponse<IAppBinaryData[]>> => {
    return http.get(`/${projectId}/apps`);
};

export const getApp = (projectId: string, id: string): Promise<AxiosResponse<IAppBinaryData | undefined>> => {
    return http.get(`/${projectId}/app/${id}`);
};

export const createApp = (projectId: string, data: IAppBinaryData): Promise<AxiosResponse<IAppBinaryData>> => {
    return http.post('/${projectId}/app', data);
};

export const updateApp = (projectId: string, data: IAppBinaryData, id: number): Promise<AxiosResponse<IAppBinaryData>> => {
    return http.put(`/${projectId}/app/${id}`, data);
};

export const deleteApp = (projectId: string, id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/${projectId}/app/${id}`);
};

export type AppFilter = {
    name?: string;
};

export const findApp = (filter?: AppFilter): Promise<AxiosResponse<IAppBinaryData[]>> => {
    return http.get('/apps', { params: filter });
};

export const uploadNewApp = (file: File,
    uploadProgress: (progressEvent: number) => void,
    finished: (finished: AxiosResponse<IAppBinaryData>) => void): void => {

    const formData = new FormData();
    formData.append('test_target', file);

    http.request({
        method: 'post',
        url: 'app/upload',
        data: formData,
        headers: {
            'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: progressEvent => {
            uploadProgress((progressEvent.loaded / progressEvent.total) * 100);
        },
    }).then(data => {
        console.log(data);
        uploadProgress(100);
        finished(data);
    });
};