import http from '../http-common';
import IProject from "./project";

export const getProjects = (): Promise<IProject[]> => {
    return http.get('/projects').then(response => response.data)
};

export const updateProject = (id: string, data: IProject): Promise<IProject> => {
    return http.put(`/project/${id}`, data).then(response => response.data)
};

export const createProject = (data: IProject): Promise<IProject> => {
    return http.post('/project', data).then(response => response.data)
};

export const deleteProject = (id: string): Promise<void> => {
    return http.delete(`/project/${id}`)
};
