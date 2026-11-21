// Helpers of the functional tests of the settings, profile and project pages:
// a fake ApiClient that answers by "METHOD url", and the providers a page needs.
import React, {useEffect} from 'react';
import {render, RenderResult} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {MemoryRouter, Route, Routes, useLocation} from 'react-router-dom';
import {ApiClient} from '../../api/ApiClient';
import {ApiClientProvider} from '../../api/ApiClientProvider';
import {HubStateActions} from '../../application/HubState';
import {ErrorProvider} from '../../ErrorProvider';
import {HubStateProvider, useHubState} from '../../hooks/HubStateProvider';
import {ProjectProvider} from '../../hooks/ProjectProvider';
import IProject from '../../project/project';
import IAccessTokenData from '../../types/access.token';
import {AppParameter, IAppData} from '../../types/app';
import {INodeData} from '../../types/node';
import {PlatformType} from '../../types/platform.type.enum';

export type User = ReturnType<typeof userEvent.setup>;

// what a test gets after opening a page: the fake master and the person
export interface Opened {
    api: FakeApi,
    user: User,
}

export type Handler = (body: unknown) => unknown;

export interface Call {
    key: string,
    body: unknown,
}

export interface FakeApi {
    client: ApiClient,
    // every request in the order it was made
    calls: Call[],
    // the bodies sent with a request, e.g. bodies('POST /p1/app')
    bodies: (key: string) => unknown[],
}

// apiError is what a failing request rejects with: the server's message
export const apiError = (message: string): unknown => ({response: {data: {message}}, message: 'Request failed'});

// createFakeApi answers requests from handlers keyed "GET /path"; a handler
// returns the body (or throws / returns a rejected promise); a request
// without a handler fails
export const createFakeApi = (handlers: Record<string, Handler> = {}): FakeApi => {
    const calls: Call[] = [];
    const request = async <T, >(method: string, url: string, body?: unknown): Promise<T> => {
        const key = `${method} ${url}`;
        calls.push({key, body});
        const handler = handlers[key];
        if (!handler) throw new Error(`no fake answer for ${key}`);
        return await handler(body) as T;
    };
    const client: ApiClient = {
        get: (url) => request('GET', url),
        post: (url, body) => request('POST', url, body),
        put: (url, body) => request('PUT', url, body),
        delete: (url) => request('DELETE', url),
    };
    return {
        client, calls,
        bodies: key => calls.filter(c => c.key === key).map(c => c.body),
    };
};

export const project = (patch: Partial<IProject> = {}): IProject =>
    ({id: 1, identifier: 'p1', name: 'Heroes', companyId: 1, accessTokens: [], apps: [], ...patch});

export const app = (patch: Partial<IAppData> = {}): IAppData =>
    ({id: 1, name: 'Game Android', identifier: 'com.game', platform: PlatformType.Android, parameter: [], ...patch}) as IAppData;

export const parameter = (patch: Partial<AppParameter> = {}): AppParameter =>
    ({id: 10, name: 'MarketId', type: {type: 'string', defaultValue: 'eu'}, ...patch});

export const accessToken = (patch: Partial<IAccessTokenData> = {}): IAccessTokenData =>
    ({id: 5, name: 'CI nightly', token: 'abcdef123456789', expiresAt: null, createdAt: new Date('2024-02-03'), ...patch}) as IAccessTokenData;

export const node = (patch: Partial<INodeData> = {}): INodeData =>
    ({id: 3, identifier: 'node-1', name: 'Mac mini QA', status: 1, ...patch});

// puts the projects into the hub state, like the layout does after loading
const SeedProjects: React.FC<{ projects: IProject[], children: React.ReactNode }> = ({projects, children}) => {
    const {dispatch} = useHubState();
    useEffect(() => dispatch({type: HubStateActions.ProjectsUpdate, payload: projects}), [dispatch, projects]);
    return <>{children}</>;
};

export interface RenderOptions {
    api: FakeApi,
    // the page's URL; the project route matches /project/:project_identifier/*
    url?: string,
    projects?: IProject[],
}

const DEFAULT_PROJECTS = [project()];

// renderInProject renders a project page inside the real providers and route
export const renderInProject = (page: React.ReactElement, {api, url = '/project/p1/settings', projects = DEFAULT_PROJECTS}: RenderOptions): RenderResult =>
    render(
        <ErrorProvider><ApiClientProvider client={api.client}><HubStateProvider>
            <MemoryRouter initialEntries={[url]}>
                <SeedProjects projects={projects}>
                    <Routes><Route path="/project/:project_identifier/*" element={<ProjectProvider>{page}</ProjectProvider>}/></Routes>
                </SeedProjects>
            </MemoryRouter>
        </HubStateProvider></ApiClientProvider></ErrorProvider>,
    );

// shows where the app navigated to
const LocationProbe: React.FC = () => {
    const location = useLocation();
    return <p>{`at ${location.pathname}${location.search}`}</p>;
};

// renderOutsideProject renders a page that needs no project (the project list, the profile)
export const renderOutsideProject = (page: React.ReactElement, api: FakeApi): RenderResult =>
    render(
        <ErrorProvider><ApiClientProvider client={api.client}><HubStateProvider>
            <MemoryRouter initialEntries={['/']}>
                <Routes>
                    <Route path="/" element={page}/>
                    <Route path="*" element={<LocationProbe/>}/>
                </Routes>
            </MemoryRouter>
        </HubStateProvider></ApiClientProvider></ErrorProvider>,
    );
