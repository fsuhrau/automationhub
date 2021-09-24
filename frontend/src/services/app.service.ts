import http from '../http-common';
import IAppData from '../types/app';
import { AxiosResponse } from 'axios';

export const getAllApps = (): Promise<AxiosResponse<IAppData[]>> => {
    return http.get('/apps');
};

export const getApp = (id: string): Promise<AxiosResponse<IAppData | undefined>> => {
    return http.get(`/app/${id}`);
};

export const createApp = (data: IAppData): Promise<AxiosResponse<IAppData>> => {
    return http.post('/app', data);
};

export const updateApp = (data: IAppData, id: number): Promise<AxiosResponse<IAppData>> => {
    return http.put(`/app/${id}`, data);
};

export const deleteApp = (id: number): Promise<AxiosResponse<void>> => {
    return http.delete(`/app/${id}`);
};

export type AppFilter = {
    name?: string;
};

export const findApp = (filter?: AppFilter): Promise<AxiosResponse<IAppData[]>> => {
    return http.get('/apps', { params: filter });
};

export const uploadNewApp = (file: File,
    uploadProgress: (progressEvent: number) => void,
    finished: (finished: AxiosResponse<IAppData>) => void): void => {

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