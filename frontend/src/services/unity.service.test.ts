import {describe, expect, it, vi} from 'vitest';
import {ApiClient} from '../api/ApiClient';
import {aFunction} from '../pages/tests/definitions/testing/fixtures';
import {unityService} from './unity.service';

describe('unityService', () => {
    it('asks the Unity editor for its tests', async () => {
        const api: ApiClient = {get: vi.fn().mockResolvedValue([aFunction('M')]), post: vi.fn(), put: vi.fn(), delete: vi.fn()};
        expect(await unityService(api).testFunctions()).toEqual([aFunction('M')]);
        expect(api.get).toHaveBeenCalledWith('/tests');
    });
});
