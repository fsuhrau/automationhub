import {act, cleanup, render, screen} from '@testing-library/react';
import React from 'react';
import {Route, Routes} from 'react-router-dom';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {useHubState} from '../hooks/HubStateProvider';
import HubStateUpdates from './HubStateUpdates';
import {SSE_CONNECTED_EVENT} from './sseSource';
import {fakeApi, FakeApi, Reply} from '../testing/fakeApi';
import {FakeSse} from '../testing/fakeSse';
import {Providers} from '../testing/renderWithProviders';

const PROJECTS = 'GET /projects';
const APPS = 'GET /p1/apps';
const NODES = 'GET /p1/settings/nodes';
const DEVICES = 'GET /p1/devices';

// what the master answers; tests change it to make a load differ
let answers: Record<string, Reply> = {};
let api: FakeApi = fakeApi();
let sse = new FakeSse();

const Probe: React.FC = () => {
    const {state} = useHubState();
    const devices = (state.devices ?? []).map(d => `${d.id}:${d.status}`).join(',');
    return <p>{`projects=${state.projects?.length ?? '-'} apps=${state.apps?.length ?? '-'} nodes=${state.nodes?.length ?? '-'} devices=${devices || '-'}`}</p>;
};

const hub = (path = '/project/p1'): React.ReactElement => <Providers api={api} sse={sse} url={path}>
    <Routes><Route path="/project/:project_identifier?" element={<HubStateUpdates><Probe/></HubStateUpdates>}/></Routes>
</Providers>;

const mountHub = async (path?: string): Promise<void> => {
    render(hub(path));
    await act(async () => undefined);
};

const flush = async (ms = 0): Promise<void> => {
    await act(async () => {
        await vi.advanceTimersByTimeAsync(ms);
    });
};

const loads = (key: string): number => api.called(key).length;

beforeEach(() => {
    vi.useFakeTimers();
    sse = new FakeSse();
    answers = {
        [PROJECTS]: [{id: 1, identifier: 'p1'}],
        [APPS]: [{id: 1}],
        [NODES]: [{id: 1}, {id: 2}],
        [DEVICES]: [{id: 9, status: 1}],
    };
    api = fakeApi(answers);
});
afterEach(() => {
    vi.useRealTimers();
});

describe('HubStateUpdates loading', () => {
    it('loads the project list once', async () => {
        await mountHub();
        expect(loads(PROJECTS)).toBe(1);
    });

    it('loads the projects but not the data of an unknown project', async () => {
        await mountHub('/project/zzz');
        expect(screen.getByText('projects=1 apps=- nodes=- devices=-')).toBeInTheDocument();
        expect(loads(APPS)).toBe(0);
    });

    it('loads the projects on pages without a project', async () => {
        await mountHub('/project');
        expect(screen.getByText('projects=1 apps=- nodes=- devices=-')).toBeInTheDocument();
    });

    it('shows the error when the projects cannot be loaded', async () => {
        answers[PROJECTS] = new Error('projects down');
        await mountHub();
        expect(screen.getByText('projects down')).toBeInTheDocument();
    });
});

describe('HubStateUpdates loading the project data', () => {

    it('loads apps, nodes and devices of the project into the state', async () => {
        await mountHub();
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
        expect(loads(APPS)).toBe(1);
        expect(loads(NODES)).toBe(1);
    });

    it.each([
        ['apps', APPS],
        ['devices', DEVICES],
    ])('shows the error when the %s cannot be loaded', async (name, failing) => {
        answers[failing] = new Error(`${name} down`);
        await mountHub();
        expect(screen.getByText(`${name} down`)).toBeInTheDocument();
    });

    it('ignores a failing node load', async () => {
        answers[NODES] = new Error('x');
        await mountHub();
        expect(screen.getByText('projects=1 apps=1 nodes=- devices=9:1')).toBeInTheDocument();
    });

    it('drops results that arrive after unmount', async () => {
        let resolveApps: (v: unknown) => void = () => undefined;
        const pendingApps = new Promise(r => {
            resolveApps = r;
        });
        answers[APPS] = () => pendingApps;
        answers[DEVICES] = new Error('late');
        render(hub());
        cleanup();
        resolveApps([{id: 1}]);
        await flush();
        expect(screen.queryByRole('alert')).toBeNull();
    });
});

describe('HubStateUpdates reloading', () => {
    it('reloads nodes and, debounced once, devices after node events', async () => {
        await mountHub();
        answers[NODES] = [{id: 1}];
        sse.emit('nodes');
        sse.emit('nodes');
        await flush(100);
        expect(loads(DEVICES)).toBe(1);
        await flush(200);
        expect(loads(DEVICES)).toBe(2);
        expect(screen.getByText('projects=1 apps=1 nodes=1 devices=9:1')).toBeInTheDocument();
    });

    it('reloads the devices after a device list event', async () => {
        await mountHub();
        answers[DEVICES] = [{id: 10, status: 2}];
        sse.emit('device_list');
        await flush(200);
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=10:2')).toBeInTheDocument();
    });

    it('does not apply a debounced reload that was cancelled', async () => {
        await mountHub();
        answers[DEVICES] = [{id: 10, status: 2}];
        sse.emit('device_list');
        cleanup();
        await flush(200);
        expect(loads(DEVICES)).toBe(1);
    });

    it('ignores a failing device reload', async () => {
        await mountHub();
        answers[DEVICES] = new Error('x');
        sse.emit('device_list');
        await flush(200);
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
    });
});

describe('HubStateUpdates reconnecting and status changes', () => {
    it('reloads after a reconnect but not after the first connect', async () => {
        await mountHub();
        act(() => {
            window.dispatchEvent(new Event(SSE_CONNECTED_EVENT));
        });
        expect(loads(NODES)).toBe(1);
        act(() => {
            window.dispatchEvent(new Event(SSE_CONNECTED_EVENT));
        });
        expect(loads(NODES)).toBe(2);
    });

    it('applies device status changes, each once', async () => {
        await mountHub();
        sse.emit('devices', {deviceId: 9, deviceState: 4});
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:4')).toBeInTheDocument();
        sse.emit('devices', {deviceId: 9, deviceState: 2});
        sse.emit('devices', {deviceId: 9, deviceState: 3});
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:3')).toBeInTheDocument();
    });

    it('ignores a device event without data', async () => {
        await mountHub();
        sse.emit('devices', null);
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
    });

    it('keeps only the latest backlog of changes', async () => {
        await mountHub();
        for (let i = 0; i < 60; i++) sse.emit('devices', {deviceId: 9, deviceState: 2 + (i % 2)});
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:3')).toBeInTheDocument();
    });
});
