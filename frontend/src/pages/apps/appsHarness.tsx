// Renders the bundle pages the way the app does: the real providers and route
// shapes around them, with the API replaced by a fake that records its calls.
import React, {useEffect} from 'react';
import {render} from '@testing-library/react';
import {MemoryRouter, Outlet, Route, Routes} from 'react-router-dom';
import {AxiosRequestConfig} from 'axios';
import {ApiClient} from '../../api/ApiClient';
import {ApiClientProvider} from '../../api/ApiClientProvider';
import {ErrorProvider} from '../../ErrorProvider';
import {HubStateActions} from '../../application/HubState';
import ApplicationPage from '../../hooks/ApplicationProvider';
import {HubStateProvider, useHubState} from '../../hooks/HubStateProvider';
import {ProjectProvider} from '../../hooks/ProjectProvider';
import IProject from '../../project/project';
import {IAppData} from '../../types/app';
import AppBundleComparePage from './AppBundleComparePage';
import AppBundlePage from './AppBundlePage';
import AppBundlesPage from './AppBundlesPage';
import {androidApp, PROJECT} from './appsFixtures';

export interface Call {
    method: 'GET' | 'POST' | 'PUT' | 'DELETE',
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
}

// the signal a request was sent with
export const signalOf = (call: Call): AbortSignal | undefined => call.config?.signal as AbortSignal | undefined;

export type Handler = (call: Call) => unknown;

export interface FakeApi extends ApiClient {
    calls: Call[],
}

// fakeApi answers "METHOD url" with the handler of that key and fails
// every request nobody expects
export const fakeApi = (handlers: Record<string, Handler>): FakeApi => {
    const calls: Call[] = [];
    const respond = async <T, >(call: Call): Promise<T> => {
        calls.push(call);
        const handler = handlers[`${call.method} ${call.url}`];
        if (!handler) throw new Error(`unexpected ${call.method} ${call.url}`);
        return handler(call) as T;
    };
    return {
        calls,
        get: <T, >(url: string, config?: AxiosRequestConfig) => respond<T>({method: 'GET', url, config}),
        post: <T, >(url: string, data?: unknown, config?: AxiosRequestConfig) => respond<T>({method: 'POST', url, data, config}),
        put: <T, >(url: string, data?: unknown, config?: AxiosRequestConfig) => respond<T>({method: 'PUT', url, data, config}),
        delete: <T, >(url: string, config?: AxiosRequestConfig) => respond<T>({method: 'DELETE', url, config}),
    };
};

// a failure like axios raises it for an answer with a message
export const failure = (message: string, status = 500): unknown => ({response: {status, data: {message}}, message: 'Request failed'});

const project = {id: 1, identifier: PROJECT, name: 'Demo'} as IProject;

const SeedHub: React.FC<{ app: IAppData }> = ({app}) => {
    const {dispatch} = useHubState();
    useEffect(() => {
        dispatch({type: HubStateActions.ProjectsUpdate, payload: [project]});
        dispatch({type: HubStateActions.AppsUpdate, payload: [app]});
    }, [dispatch, app]);
    return null;
};

// renderBundles shows the page the url leads to, as /project/demo/app:3/...
export const renderBundles = (api: ApiClient, path: string, app: IAppData = androidApp): ReturnType<typeof render> =>
    render(
        <ApiClientProvider client={api}>
            <ErrorProvider>
                <HubStateProvider>
                    <SeedHub app={app}/>
                    <MemoryRouter initialEntries={[`/project/${PROJECT}/app:3/${path}`]}>
                        <Routes>
                            <Route path="/project/:project_identifier" element={<ProjectProvider><Outlet/></ProjectProvider>}>
                                <Route path=":appId" element={<ApplicationPage/>}>
                                    <Route path="bundles" element={<AppBundlesPage/>}/>
                                    <Route path="bundles/compare" element={<AppBundleComparePage/>}/>
                                    <Route path="bundle/:bundleId" element={<AppBundlePage/>}/>
                                </Route>
                            </Route>
                        </Routes>
                    </MemoryRouter>
                </HubStateProvider>
            </ErrorProvider>
        </ApiClientProvider>,
    );
