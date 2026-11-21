import {expect, it, vi} from 'vitest';
import {InternalAxiosRequestConfig} from 'axios';
import axiosInstance from '../http-common';
import {ApiClient} from '../api/ApiClient';
import {appsSettingsService, getNodes, settingsService} from './settings.service';

const fakeClient = (): ApiClient => ({
    get: vi.fn().mockResolvedValue('got'),
    post: vi.fn().mockResolvedValue('posted'),
    put: vi.fn().mockResolvedValue('put'),
    delete: vi.fn().mockResolvedValue(undefined),
});

it('settingsService: reads tokens, nodes and users of a project', async () => {
    const api = fakeClient();
    const service = settingsService(api);
    await service.accessTokens('p');
    await service.nodes('p');
    await service.users('p');
    expect(api.get).toHaveBeenNthCalledWith(1, '/p/settings/access_tokens');
    expect(api.get).toHaveBeenNthCalledWith(2, '/p/settings/nodes');
    expect(api.get).toHaveBeenNthCalledWith(3, '/p/settings/users');
});

it('settingsService: creates tokens and nodes and issues a node token', async () => {
    const api = fakeClient();
    const service = settingsService(api);
    await service.createAccessToken('p', {name: 't', expiresAt: null});
    await service.createNode('p', {name: 'n'});
    await service.issueNodeToken('p', 4);
    expect(api.post).toHaveBeenNthCalledWith(1, '/p/settings/access_token', {name: 't', expiresAt: null});
    expect(api.post).toHaveBeenNthCalledWith(2, '/p/settings/node', {name: 'n'});
    expect(api.post).toHaveBeenNthCalledWith(3, '/p/settings/nodes/4/token');
});

it('settingsService: deletes tokens and nodes', async () => {
    const api = fakeClient();
    const service = settingsService(api);
    await service.deleteAccessToken('p', 3);
    await service.deleteNode('p', 4);
    expect(api.delete).toHaveBeenNthCalledWith(1, '/p/settings/access_token/3');
    expect(api.delete).toHaveBeenNthCalledWith(2, '/p/settings/nodes/4');
});

it('appsSettingsService: lists and creates apps of a project', async () => {
    const api = fakeClient();
    const service = appsSettingsService(api);
    await service.list('p');
    await service.create('p', {name: 'A', platform: 1, identifier: 'id'});
    expect(api.get).toHaveBeenCalledWith('/p/apps');
    expect(api.post).toHaveBeenCalledWith('/p/app', {name: 'A', platform: 1, identifier: 'id'});
});

it('appsSettingsService: updates an app and changes its parameters', async () => {
    const api = fakeClient();
    const service = appsSettingsService(api);
    const app = {projectId: 'p', appId: 2};
    const body = {name: 'N', type: {type: 'string' as const, defaultValue: ''}};
    await service.update(app, {id: 2} as never);
    await service.addParameter(app, body);
    await service.updateParameter(app, 5, body);
    await service.removeParameter(app, 5);
    expect(api.put).toHaveBeenNthCalledWith(1, '/p/app/2', {id: 2});
    expect(api.post).toHaveBeenCalledWith('/p/app/2/parameter', body);
    expect(api.put).toHaveBeenNthCalledWith(2, '/p/app/2/parameter/5', body);
    expect(api.delete).toHaveBeenCalledWith('/p/app/2/parameter/5');
});

it('legacy getNodes: reads the nodes through the shared axios instance', async () => {
    const adapter = vi.fn(async (config: InternalAxiosRequestConfig) =>
        ({data: [{id: 1}], status: 200, statusText: '', headers: {}, config}));
    const original = axiosInstance.defaults.adapter;
    axiosInstance.defaults.adapter = adapter;
    try {
        await expect(getNodes('p')).resolves.toEqual([{id: 1}]);
        expect(adapter.mock.calls[0][0].url).toBe('/p/settings/nodes');
    } finally {
        axiosInstance.defaults.adapter = original;
    }
});
