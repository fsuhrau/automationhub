// Fakes and builders for the tests of the device pages: a fake ApiClient, a
// fake EventSource, device builders and a render that wires the providers
// and routes the pages run in.
import React, {ReactElement, useEffect} from 'react';
import {render, RenderResult} from '@testing-library/react';
import {MemoryRouter, Outlet, Route, Routes, useLocation} from 'react-router-dom';
import {vi} from 'vitest';
import {ApiClient} from '../../api/ApiClient';
import {ApiClientProvider} from '../../api/ApiClientProvider';
import {HubStateAction, HubStateActions} from '../../application/HubState';
import {HubStateProvider, useHubState} from '../../hooks/HubStateProvider';
import {ProjectProvider} from '../../hooks/ProjectProvider';
import {SseSource, SseSourceFactory} from '../../hooks/useSseLog';
import IDeviceData from '../../types/device';
import {INodeData} from '../../types/node';
import {DeviceStateType} from '../../types/deviceStateType';
import {DeviceType} from '../../types/device.type.enum';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {PlatformType} from '../../types/platform.type.enum';

export interface Call {
    method: string,
    url: string,
    data?: unknown,
    config?: unknown,
}

export type Reply = unknown | ((call: Call) => unknown);

export interface FakeApi extends ApiClient {
    calls: Call[],
}

const answer = async (reply: Reply, call: Call): Promise<unknown> => {
    const value = typeof reply === 'function' ? (reply as (c: Call) => unknown)(call) : reply;
    if (value instanceof Error) throw value;
    return value;
};

// fakeApi answers by "METHOD /url"; an Error reply rejects, a function
// reply gets the call; every call is recorded
export const fakeApi = (replies: Record<string, Reply>): FakeApi => {
    const calls: Call[] = [];
    const handle = (method: string) => (url: string, data?: unknown, config?: unknown): Promise<never> => {
        const call = {method, url, data, config};
        calls.push(call);
        const key = `${method} ${url}`;
        if (!(key in replies)) return Promise.reject(new Error(`no fake answer for ${key}`));
        return answer(replies[key], call) as Promise<never>;
    };
    return {
        calls,
        get: (url, config) => handle('GET')(url, undefined, config),
        post: handle('POST'),
        put: handle('PUT'),
        delete: (url, config) => handle('DELETE')(url, undefined, config),
    };
};

export class FakeSource {
    closed = false;
    private listeners: Record<string, ((e: Event) => void)[]> = {};

    constructor(readonly url: string) {
    }

    addEventListener(type: string, listener: EventListenerOrEventListenerObject): void {
        (this.listeners[type] ??= []).push(listener as (e: Event) => void);
    }

    close(): void {
        this.closed = true;
    }

    // delivers an event; data is sent as JSON like the master does
    emit(type: string, data?: unknown): void {
        const event = new MessageEvent(type, {data: JSON.stringify(data)});
        (this.listeners[type] ?? []).forEach(l => l(event));
    }
}

export interface FakeSources {
    create: SseSourceFactory,
    sources: FakeSource[],
}

export const fakeSources = (): FakeSources => {
    const sources: FakeSource[] = [];
    const create: SseSourceFactory = url => {
        const source = new FakeSource(url);
        sources.push(source);
        return source as unknown as SseSource;
    };
    return {create: vi.fn(create), sources};
};

export const device = (change: Partial<IDeviceData> = {}): IDeviceData => ({
    id: 1, nodeId: 0, companyId: 1, deviceIdentifier: 'serial-1', deviceType: DeviceType.Phone,
    platformType: PlatformType.iOS, name: 'Pixel', alias: '', os: 'Android', osVersion: '14',
    status: DeviceStateType.Booted, isLocked: false, manager: '', isAcknowledged: true,
    connectionType: DeviceConnectionType.USB,
    connectionParameter: {connectionType: DeviceConnectionType.USB, ip: '', port: 0},
    customParameter: [], deviceParameter: [],
    ...change,
});

export const node = (change: Partial<INodeData> = {}): INodeData =>
    ({id: 5, identifier: 'node-a-id', name: 'node-a', status: 1, ...change});

const project = {id: 1, identifier: 'p1', name: 'P1', companyId: 1, accessTokens: [], apps: []};

// puts the hub state in place once, as the app's loaders would
const Seed: React.FC<{ actions: HubStateAction[] }> = ({actions}) => {
    const {dispatch} = useHubState();
    useEffect(() => actions.forEach(dispatch), [actions, dispatch]);
    return null;
};

const Probe: React.FC = () => <div data-testid="location">{useLocation().pathname}</div>;

export interface PageOptions {
    api?: ApiClient,
    // the route of the page below /project/:project_identifier, and the url opened
    route: string,
    url: string,
    devices?: IDeviceData[],
    nodes?: INodeData[],
}

// renders the page in its route inside the providers; a navigation away
// from it shows the new path in the element "location"
export const renderPage = (page: ReactElement, options: PageOptions): RenderResult => {
    const {api = fakeApi({}), route, url, devices, nodes} = options;
    const actions: HubStateAction[] = [
        {type: HubStateActions.ProjectsUpdate, payload: [project]},
        ...(devices ? [{type: HubStateActions.UpdateDevices, payload: devices} as HubStateAction] : []),
        ...(nodes ? [{type: HubStateActions.NodesUpdate, payload: nodes} as HubStateAction] : []),
    ];
    return render(
        <ApiClientProvider client={api}>
            <HubStateProvider>
                <Seed actions={actions}/>
                <MemoryRouter initialEntries={[url]}>
                    <Routes>
                        <Route path="/project/:project_identifier" element={<ProjectProvider><Outlet/></ProjectProvider>}>
                            <Route path={route} element={page}/>
                            <Route path="*" element={<Probe/>}/>
                        </Route>
                        <Route path="*" element={<Probe/>}/>
                    </Routes>
                </MemoryRouter>
            </HubStateProvider>
        </ApiClientProvider>,
    );
};
