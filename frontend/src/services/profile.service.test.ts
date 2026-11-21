import {expect, it, vi} from 'vitest';
import dayjs from 'dayjs';
import {ApiClient} from '../api/ApiClient';
import {profileService} from './profile.service';

const fakeClient = (): ApiClient => ({
    get: vi.fn().mockResolvedValue('got'),
    post: vi.fn().mockResolvedValue('posted'),
    put: vi.fn(),
    delete: vi.fn().mockResolvedValue('deleted'),
});

it('profileService: reads the profile and the tokens', async () => {
    const api = fakeClient();
    const service = profileService(api);
    await expect(service.profile()).resolves.toBe('got');
    await service.tokens();
    expect(api.get).toHaveBeenNthCalledWith(1, '/me');
    expect(api.get).toHaveBeenNthCalledWith(2, '/me/tokens');
});

it('profileService: creates a token that ends at the end of its expiry day', async () => {
    const api = fakeClient();
    const day = dayjs('2030-05-06');
    await profileService(api).createToken('ci', day);
    expect(api.post).toHaveBeenCalledWith('/me/tokens', {name: 'ci', expiresAt: day.endOf('day').toISOString()});
});

it('profileService: creates a token that never expires', async () => {
    const api = fakeClient();
    await profileService(api).createToken('ci', null);
    expect(api.post).toHaveBeenCalledWith('/me/tokens', {name: 'ci', expiresAt: null});
});

it('profileService: deletes a token and resolves without a value', async () => {
    const api = fakeClient();
    await expect(profileService(api).deleteToken(7)).resolves.toBeUndefined();
    expect(api.delete).toHaveBeenCalledWith('/me/tokens/7');
});
