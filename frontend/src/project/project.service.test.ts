import {expect, it, vi} from 'vitest';
import {ApiClient} from '../api/ApiClient';
import {projectService} from './project.service';
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
