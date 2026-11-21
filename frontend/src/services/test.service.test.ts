import {describe, expect, it, vi} from 'vitest';
import {ApiClient} from '../api/ApiClient';
import {createFakeApi} from '../pages/tests/definitions/testing/fakeApi';
import {aTest} from '../pages/tests/definitions/testing/fixtures';
import {TestType} from '../types/test.type.enum';
import {UpdateTestData} from '../types/test.requests';
import {cancelTestRun, executeTest, getTest, testService} from './test.service';

const fakeClient = (): ApiClient => ({
    get: vi.fn().mockResolvedValue('got'),
    post: vi.fn().mockResolvedValue('posted'),
    put: vi.fn().mockResolvedValue('put'),
    delete: vi.fn(),
});

const scope = {projectId: 'p', appId: 2};
const data = {name: 'x'} as UpdateTestData;

describe('testService', () => {
    it('lists and gets tests of the app', async () => {
        const api = fakeClient();
        const service = testService(api);
        expect(await service.list(scope)).toBe('got');
        expect(await service.get(scope, '7')).toBe('got');
        expect(api.get).toHaveBeenNthCalledWith(1, '/p/app/2/tests');
        expect(api.get).toHaveBeenNthCalledWith(2, '/p/app/2/test/7');
    });

    it('creates and updates a test with its settings', async () => {
        const api = fakeClient();
        const service = testService(api);
        expect(await service.create(scope, {type: TestType.Unity, ...data})).toBe('posted');
        expect(await service.update(scope, 7, data)).toBe('put');
        expect(api.post).toHaveBeenCalledWith('/p/app/2/test', {type: TestType.Unity, name: 'x'});
        expect(api.put).toHaveBeenCalledWith('/p/app/2/test/7', data);
    });

    it('starts a run and cancels one', async () => {
        const api = fakeClient();
        const service = testService(api);
        const run = {appBinaryId: 1, startUrl: null, params: 'a=1'};
        await service.execute(scope, 7, run);
        await service.cancelRun(scope, 7, 11);
        expect(api.post).toHaveBeenNthCalledWith(1, '/p/app/2/test/7/run', run);
        expect(api.post).toHaveBeenNthCalledWith(2, '/p/app/2/test/7/run/11/cancel');
    });
});

describe('legacy wrappers', () => {
    it('go through the shared axios instance', async () => {
        const api = createFakeApi({
            'GET /p/app/2/test/7': aTest(),
            'POST /p/app/2/test/7/run': {id: 11},
            'POST /p/app/2/test/7/run/11/cancel': null,
        });
        try {
            expect(await getTest('p', 2, '7')).toMatchObject({id: 7});
            expect(await executeTest('p', 2, 7, {appBinaryId: null, startUrl: null, params: ''})).toEqual({id: 11});
            await cancelTestRun('p', 2, 7, 11);
            expect(api.calls.map(c => `${c.method} ${c.url}`)).toEqual([
                'GET /p/app/2/test/7', 'POST /p/app/2/test/7/run', 'POST /p/app/2/test/7/run/11/cancel',
            ]);
        } finally {
            api.restore();
        }
    });
});
