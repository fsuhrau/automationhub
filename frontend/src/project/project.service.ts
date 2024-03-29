import http from '../http-common';
import { AxiosResponse } from 'axios';
import IProject from "./project";

export const getProjects = (): Promise<AxiosResponse<IProject[]>> => {
    return http.get('/projects');
};

export const updateProject = (id: string, data: IProject): Promise<AxiosResponse<IProject>> => {
    return http.put(`/project/${id}`, data);
};

export const createProject = (data: IProject): Promise<AxiosResponse<IProject>> => {
    return http.post('/project', data);
};

export const deleteProject = (id: string): Promise<AxiosResponse<void>> => {
    return http.delete(`/project/${id}`);
};
