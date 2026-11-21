import {useMemo} from 'react';
import {ApiClient} from '../api/ApiClient';
import {useApiClient} from '../api/ApiClientProvider';
import IProject from './project';

export interface ProjectService {
    list: () => Promise<IProject[]>,
    update: (id: string, data: IProject) => Promise<IProject>,
    create: (data: Pick<IProject, 'name'>) => Promise<IProject>,
    remove: (id: string) => Promise<void>,
}

export const projectService = (api: ApiClient): ProjectService => ({
    list: () => api.get('/projects'),
    update: (id, data) => api.put(`/project/${id}`, data),
    create: (data) => api.post('/project', data),
    remove: (id) => api.delete(`/project/${id}`),
});

export const useProjectService = (): ProjectService => {
    const api = useApiClient();
    return useMemo(() => projectService(api), [api]);
};
