import {describe, expect, it} from 'vitest';
import {renderHook} from '@testing-library/react';
import {fakeApi} from '../testing/fakeApi';
import {wrapperOf} from '../testing/renderWithProviders';
import {at, run} from '../pages/tests/run/harness/fixtures';
import {testRunService, useTestRunService} from './test.run.service';

const test = {projectId: 'p1', appId: 1, testId: 2};
const ref = {...test, runId: 3, protocolId: 4};
const base = '/p1/app/1/test/2';

describe('testRunService', () => {
    it('pages through the runs with a limit and the run to start before', async () => {
        const api = fakeApi({[`GET ${base}/runs`]: {runs: [], more: true}});
        await expect(testRunService(api).runsPage(test, {before: 9})).resolves.toEqual({runs: [], more: true});
        expect(api.calls[0].config?.params).toEqual({limit: 25, before: 9});
    });

    it('asks for the default limit without paging', async () => {
        const api = fakeApi({[`GET ${base}/runs`]: {runs: [], more: false}});
        await testRunService(api).runsPage(test);
        expect(api.calls[0].config?.params).toEqual({limit: 25});
    });

    it('turns the plain list of an older master into a page, newest first', async () => {
        const api = fakeApi({[`GET ${base}/runs`]: [run({id: 1, parameter: 'a'}), run({id: undefined}), run({id: 5, parameter: 'b'})]});
        const page = await testRunService(api).runsPage(test);
        expect(page.more).toBe(false);
        expect(page.runs.map(r => [r.id, r.parameter])).toEqual([[5, 'b'], [1, 'a'], [undefined, '']]);
        expect(page.runs[0]).toMatchObject({state: 'not_running', passed: 0, bundle: null, endedAt: null, createdAt: at(0)});
    });

    it('loads the last run, a run and a protocol', async () => {
        const api = fakeApi({
            [`GET ${base}/runs/last`]: 'last', [`GET ${base}/run/3`]: 'run', [`GET ${base}/run/3/4`]: 'protocol',
        });
        const service = testRunService(api);
        expect(await service.lastRun(test)).toBe('last');
        expect(await service.run(ref)).toBe('run');
        expect(await service.protocol(ref)).toBe('protocol');
    });

});

describe('testRunService: baselines and actions', () => {
    it('lists baselines with only the filters that are set', async () => {
        const api = fakeApi({[`GET ${base}/run/3/4/baselines`]: {candidates: [], tags: []}});
        const service = testRunService(api);
        await service.baselines(ref, {q: 'x', tag: 't', anyDevice: true, limit: 5});
        await service.baselines(ref, {});
        expect(api.calls.map(r => r.config?.params)).toEqual([{q: 'x', tag: 't', anyDevice: 'true', limit: 5}, {}]);
    });

    it('loads a baseline on the given device', async () => {
        const api = fakeApi({[`GET ${base}/run/3/4/baseline/8`]: 'baseline'});
        const service = testRunService(api);
        expect(await service.baseline(ref, 8, 7)).toBe('baseline');
        await service.baseline(ref, 8);
        expect(api.calls.map(r => r.config?.params)).toEqual([{deviceId: 7}, {}]);
    });

    it('starts and cancels runs', async () => {
        const api = fakeApi({[`POST ${base}/run`]: 'started', [`POST ${base}/run/3/cancel`]: null});
        const service = testRunService(api);
        const data = {appBinaryId: 5, startUrl: '', params: ''};
        expect(await service.execute(test, data)).toBe('started');
        expect(api.calls[0].data).toEqual(data);
        await service.cancel(ref);
        expect(api.calls[1].key).toBe(`POST ${base}/run/3/cancel`);
    });
});

describe('useTestRunService', () => {
    it('is the service over the client of the provider', async () => {
        const api = fakeApi({[`GET ${base}/run/3`]: 'run'});
        const {result, rerender} = renderHook(() => useTestRunService(), {wrapper: wrapperOf({api})});
        const first = result.current;
        rerender();
        expect(result.current).toBe(first);
        expect(await result.current.run(ref)).toBe('run');
    });
});
