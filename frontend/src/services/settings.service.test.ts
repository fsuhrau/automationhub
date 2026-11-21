import {expect, it, vi} from 'vitest';
import {ApiClient} from '../api/ApiClient';
import {settingsService} from './settings.service';

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
