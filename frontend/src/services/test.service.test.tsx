import {renderHook} from '@testing-library/react';
import {describe, expect, it} from 'vitest';
import {fakeApi, FakeApi} from '../testing/fakeApi';
import {wrapperOf} from '../testing/renderWithProviders';
import {TestType} from '../types/test.type.enum';
import {UpdateTestData} from '../types/test.requests';
import {testService, useTestService} from './test.service';

const scope = {projectId: 'p', appId: 2};
const data = {name: 'x'} as UpdateTestData;

const master = (): FakeApi => fakeApi({
    'GET /p/app/2/tests': [{id: 7}],
    'GET /p/app/2/test/7': {id: 7},
    'POST /p/app/2/test': {id: 8},
    'PUT /p/app/2/test/7': {id: 7, name: 'x'},
});

describe('testService', () => {
    it('lists and gets tests of the app', async () => {
        const service = testService(master());
        expect(await service.list(scope)).toEqual([{id: 7}]);
        expect(await service.get(scope, '7')).toEqual({id: 7});
    });

    it('creates and updates a test with its settings', async () => {
        const api = master();
        const service = testService(api);
        expect(await service.create(scope, {type: TestType.Unity, ...data})).toEqual({id: 8});
        expect(await service.update(scope, 7, data)).toEqual({id: 7, name: 'x'});
        expect(api.bodies('POST /p/app/2/test')).toEqual([{type: TestType.Unity, name: 'x'}]);
        expect(api.bodies('PUT /p/app/2/test/7')).toEqual([data]);
    });
});

describe('useTestService', () => {
    it('is the service over the client of the provider, the same on every render', async () => {
        const api = master();
        const {result, rerender} = renderHook(() => useTestService(), {wrapper: wrapperOf({api})});
        const first = result.current;
        rerender();
        expect(result.current).toBe(first);
        expect(await first.list(scope)).toEqual([{id: 7}]);
    });
});
