import {describe, expect, it} from 'vitest';
import {act, renderHook, RenderHookResult, screen, waitFor} from '@testing-library/react';
import {Location, useLocation} from 'react-router-dom';
import {fakeApi, FakeApi} from '../../../testing/fakeApi';
import {run} from './harness/fixtures';
import {wrapperOf} from '../../../testing/renderWithProviders';
import {RunActions, useRunActions} from './useRunActions';

const test = {projectId: 'p1', appId: 1, testId: 2};
const base = '/p1/app/1/test/2';

interface Probe {
    actions: RunActions,
    location: Location,
}

interface Setup {
    api: FakeApi,
    result: RenderHookResult<Probe, unknown>['result'],
}

const setup = (answers: Record<string, unknown>): Setup => {
    const api = fakeApi(answers);
    const useProbe = (): Probe => ({actions: useRunActions(run({parameter: 'A=1', startUrl: 'u'}), test), location: useLocation()});
    return {api, result: renderHook(useProbe, {wrapper: wrapperOf({api, url: '/start'})}).result};
};

describe('useRunActions', () => {
    it('starts the test again with the bundle and parameters and opens the new run', async () => {
        const {api, result} = setup({[`POST ${base}/run`]: run({id: 4})});
        await act(async () => result.current.actions.rerun());
        expect(api.calls[0].data).toEqual({appBinaryId: 5, params: 'A=1', startUrl: 'u'});
        expect(result.current.location.pathname).toBe('/project/p1/app:1/test/2/run/4');
    });

    it('cancels the run and stays cancelling', async () => {
        const {api, result} = setup({[`POST ${base}/run/3/cancel`]: null});
        expect(result.current.actions.cancelling).toBe(false);
        await act(async () => result.current.actions.cancel());
        expect(api.calls[0].key).toBe(`POST ${base}/run/3/cancel`);
        expect(result.current.actions.cancelling).toBe(true);
    });

    it('shows the failure of an action in the error banner', async () => {
        const {result} = setup({[`POST ${base}/run/3/cancel`]: new Error('cannot cancel')});
        await act(async () => result.current.actions.cancel());
        await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('cannot cancel'));
        expect(result.current.actions.cancelling).toBe(false);
    });
});
