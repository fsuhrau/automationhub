import {expect, it, vi} from 'vitest';
import {InternalAxiosRequestConfig} from 'axios';
import axiosInstance from '../http-common';
import {ApiClient} from '../api/ApiClient';
import {getProjects, projectService} from './project.service';
import IProject from './project';

const fakeClient = (): ApiClient => ({
    get: vi.fn().mockResolvedValue([]),
    post: vi.fn().mockResolvedValue({}),
    put: vi.fn().mockResolvedValue({}),
    delete: vi.fn().mockResolvedValue(undefined),
});

it('projectService: lists, creates, updates and removes projects', async () => {
    const api = fakeClient();
    const service = projectService(api);
    await service.list();
    await service.create({name: 'N'});
    await service.update('p', {name: 'M'} as IProject);
    await service.remove('p');
    expect(api.get).toHaveBeenCalledWith('/projects');
    expect(api.post).toHaveBeenCalledWith('/project', {name: 'N'});
    expect(api.put).toHaveBeenCalledWith('/project/p', {name: 'M'});
    expect(api.delete).toHaveBeenCalledWith('/project/p');
});

it('legacy getProjects: reads the projects through the shared axios instance', async () => {
    const adapter = vi.fn(async (config: InternalAxiosRequestConfig) =>
        ({data: [{id: 1}], status: 200, statusText: '', headers: {}, config}));
    const original = axiosInstance.defaults.adapter;
    axiosInstance.defaults.adapter = adapter;
    try {
        await expect(getProjects()).resolves.toEqual([{id: 1}]);
        expect(adapter.mock.calls[0][0].url).toBe('/projects');
    } finally {
        axiosInstance.defaults.adapter = original;
    }
});
