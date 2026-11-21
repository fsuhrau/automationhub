// Renders the control page for tests: a fake ApiClient that answers by
// "METHOD url", the providers of the real route, and a fake video player.
import React, {useEffect} from 'react';
import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {expect} from 'vitest';
import type {AxiosRequestConfig} from 'axios';
import {MemoryRouter, Outlet, Route, Routes} from 'react-router-dom';
import {ApiClient} from '../../../api/ApiClient';
import {ApiClientProvider} from '../../../api/ApiClientProvider';
import {HubStateActions} from '../../../application/HubState';
import {ErrorProvider} from '../../../ErrorProvider';
import {HubStateProvider, useHubState} from '../../../hooks/HubStateProvider';
import {ProjectProvider} from '../../../hooks/ProjectProvider';
import IProject from '../../../project/project';
import {DeviceLockStatus} from '../../../types/device.control.lease';
import DeviceControlPage from '../DeviceControlPage';
import {APPS, BUNDLES, DEVICE, fakeSupport, FakeSupport, FREE, leaseOf, heldBy} from './fakes';

export const BASE = '/p1/device/7';

export interface Call {
    method: string,
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
}

type Answer = unknown | ((data: unknown) => unknown);

export class FakeApi implements ApiClient {
    readonly calls: Call[] = [];

    constructor(private readonly answers: Record<string, Answer>) {
    }

    // the calls made to one endpoint
    to(method: string, url: string): Call[] {
        return this.calls.filter(c => c.method === method && c.url === url);
    }

    private respond<T>(call: Call): Promise<T> {
        this.calls.push(call);
        const answer = this.answers[`${call.method} ${call.url}`];
        if (answer === undefined) return Promise.reject(new Error(`unexpected ${call.method} ${call.url}`));
        const value = typeof answer === 'function' ? answer(call.data) : answer;
        return value instanceof Error ? Promise.reject(value) : Promise.resolve(value as T);
    }

    get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
        return this.respond({method: 'GET', url, config});
    }

    post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
        return this.respond({method: 'POST', url, data, config});
    }

    put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
        return this.respond({method: 'PUT', url, data, config});
    }

    delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
        return this.respond({method: 'DELETE', url, config});
    }
}

export const XML = `<Scene RectangleX="200" RectangleY="400">
  <Panel X="0" Y="0" RectangleX="200" RectangleY="200"><Button ID="b1" Name="Play" LabelText="Go" X="20" Y="40" RectangleX="40" RectangleY="20"/></Panel>
</Scene>`;

export const LOG = ['11-06 13:51:05.123  1  2 I App: hello', '11-06 13:51:06.000  1  2 E App: boom'];

// the hub of one device that is free, until somebody takes it
const hubAnswers = (world: { lock: DeviceLockStatus }): Record<string, Answer> => ({
    [`GET ${BASE}`]: DEVICE,
    'GET /p1/apps': APPS,
    'GET /p1/app/1/bundles': BUNDLES,
    'GET /p1/app/2/bundles': BUNDLES,
    [`GET ${BASE}/lock`]: () => world.lock,
    [`POST ${BASE}/lock`]: () => {
        world.lock = heldBy();
        return leaseOf();
    },
    [`DELETE ${BASE}/lock`]: () => {
        world.lock = FREE;
        return null;
    },
    [`GET ${BASE}/hierarchy`]: XML,
    [`GET ${BASE}/logs`]: {lines: LOG},
});

// the inputs answer with an empty body
const INPUTS = ['tap', 'swipe', 'text', 'key', 'element/touch', 'app/install', 'app/start', 'app/stop'];

const project = {id: 1, identifier: 'p1', name: 'P', companyId: 1, accessTokens: [], apps: []} as IProject;

const Seed: React.FC = () => {
    const {dispatch} = useHubState();
    useEffect(() => dispatch({type: HubStateActions.ProjectsUpdate, payload: [project]}), [dispatch]);
    return null;
};

export interface PageOptions {
    answers?: Record<string, Answer>,
    lock?: DeviceLockStatus,
    support?: FakeSupport,
}

export interface PageWorld {
    api: FakeApi,
    support: FakeSupport,
    world: { lock: DeviceLockStatus },
}

export const renderPage = ({answers = {}, lock = FREE, support = fakeSupport()}: PageOptions = {}): PageWorld => {
    const world = {lock};
    const inputs = Object.fromEntries(INPUTS.map(i => [`POST ${BASE}/${i}`, null]));
    const api = new FakeApi({...hubAnswers(world), ...inputs, ...answers});
    render(
        <ApiClientProvider client={api}><ErrorProvider><HubStateProvider><Seed/>
            <MemoryRouter initialEntries={['/project/p1/device/7/control']}>
                <Routes>
                    <Route path="/project/:project_identifier" element={<ProjectProvider><Outlet/></ProjectProvider>}>
                        <Route path="device/:deviceId/control" element={<DeviceControlPage support={support}/>}/>
                    </Route>
                </Routes>
            </MemoryRouter>
        </HubStateProvider></ErrorProvider></ApiClientProvider>,
    );
    return {api, support, world};
};

export const loaded = async (): Promise<void> => {
    await screen.findByRole('heading', {name: 'Lab phone'});
};

// waits until the hub has given the lease to this browser
export const takeControl = async (page: PageWorld): Promise<void> => {
    fireEvent.click(screen.getAllByRole('button', {name: 'Take control'})[0]);
    await screen.findByText('You control this device');
    await waitFor(() => expect(page.world.lock.locked).toBe(true));
};

export const controlled = async (options?: PageOptions): Promise<PageWorld> => {
    const page = renderPage(options);
    await loaded();
    await takeControl(page);
    return page;
};
