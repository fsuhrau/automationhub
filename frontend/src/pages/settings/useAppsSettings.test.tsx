import React from 'react';
import {expect, it} from 'vitest';
import {act, renderHook, RenderHookResult, waitFor} from '@testing-library/react';
import {ApiClientProvider} from '../../api/ApiClientProvider';
import {ErrorProvider} from '../../ErrorProvider';
import {HubStateProvider} from '../../hooks/HubStateProvider';
import {createFakeApi, FakeApi} from './testKit';
import {parameterForm} from './settingsModel';
import {AppsSettings, useAppsSettings} from './useAppsSettings';

// without a project context the hook falls back to the empty identifier
const setup = (): { api: FakeApi } & RenderHookResult<AppsSettings, unknown> => {
    const api = createFakeApi({'GET //apps': () => []});
    const wrapper = ({children}: { children: React.ReactNode }): React.ReactElement =>
        <ErrorProvider><ApiClientProvider client={api.client}><HubStateProvider>{children}</HubStateProvider></ApiClientProvider></ErrorProvider>;
    return {api, ...renderHook(() => useAppsSettings(), {wrapper})};
};

it('useAppsSettings without an app: has no apps and no selection', async () => {
    const {result, api} = setup();
    await waitFor(() => expect(api.calls).toHaveLength(1));
    expect(result.current).toMatchObject({apps: [], app: null});
});

it('useAppsSettings without an app: does not call the master to rename, remove or save a parameter', async () => {
    const {result, api} = setup();
    await waitFor(() => expect(api.calls).toHaveLength(1));
    let outcomes: unknown[] = [];
    await act(async () => {
        outcomes = [
            await result.current.rename.run('x'),
            await result.current.removeParameter.run({id: 1, name: 'p', type: {type: 'string', defaultValue: ''}}),
            await result.current.saveParameter(null, parameterForm(null)),
        ];
    });
    expect(outcomes).toEqual([false, false, false]);
    expect(api.calls).toHaveLength(1);
});
