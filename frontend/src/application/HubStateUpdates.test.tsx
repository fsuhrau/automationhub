import {act, cleanup, render, screen} from '@testing-library/react';
import React from 'react';
import {MemoryRouter, Route, Routes} from 'react-router-dom';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {ErrorProvider} from '../ErrorProvider';
import {HubStateProvider, useHubState} from '../hooks/HubStateProvider';
import HubStateUpdates from './HubStateUpdates';
import {SSE_CONNECTED_EVENT} from './sseSource';
import {resetSse, sendSse as send} from '../app/testing/fakeSse';

// the services are not on ApiClient yet (their packages migrate them), so they are faked as modules
const services = vi.hoisted(() => ({
    getAllDevices: vi.fn(),
    getAllApps: vi.fn(),
    getNodes: vi.fn(),
    getProjects: vi.fn(),
}));

vi.mock('react-hooks-sse', async () => (await import('../app/testing/fakeSse')).fakeSseModule);
vi.mock('../services/device.service', () => ({getAllDevices: services.getAllDevices}));
vi.mock('../services/app.service', () => ({getAllApps: services.getAllApps}));
vi.mock('../services/settings.service', () => ({getNodes: services.getNodes}));
vi.mock('../project/project.service', () => ({getProjects: services.getProjects}));

const Probe: React.FC = () => {
    const {state} = useHubState();
    const devices = (state.devices ?? []).map(d => `${d.id}:${d.status}`).join(',');
    return <p>{`projects=${state.projects?.length ?? '-'} apps=${state.apps?.length ?? '-'} nodes=${state.nodes?.length ?? '-'} devices=${devices || '-'}`}</p>;
};

const hub = (path = '/project/p1'): React.ReactElement => <ErrorProvider><HubStateProvider>
    <MemoryRouter initialEntries={[path]}>
        <Routes><Route path="/project/:project_identifier?" element={<HubStateUpdates><Probe/></HubStateUpdates>}/></Routes>
    </MemoryRouter>
</HubStateProvider></ErrorProvider>;

const mountHub = async (path?: string): Promise<void> => {
    render(hub(path));
    await act(async () => undefined);
};

const flush = async (ms = 0): Promise<void> => {
    await act(async () => {
        await vi.advanceTimersByTimeAsync(ms);
    });
};

beforeEach(() => {
    vi.useFakeTimers();
    resetSse();
    services.getProjects.mockResolvedValue([{id: 1, identifier: 'p1'}]);
    services.getAllApps.mockResolvedValue([{id: 1}]);
    services.getNodes.mockResolvedValue([{id: 1}, {id: 2}]);
    services.getAllDevices.mockResolvedValue([{id: 9, status: 1}]);
});
afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
});

describe('HubStateUpdates loading', () => {
    it('loads the project list once', async () => {
        await mountHub();
        expect(services.getProjects).toHaveBeenCalledTimes(1);
    });

    it('loads the projects but not the data of an unknown project', async () => {
        await mountHub('/project/zzz');
        expect(screen.getByText('projects=1 apps=- nodes=- devices=-')).toBeInTheDocument();
        expect(services.getAllApps).not.toHaveBeenCalled();
    });

    it('loads the projects on pages without a project', async () => {
        await mountHub('/project');
        expect(screen.getByText('projects=1 apps=- nodes=- devices=-')).toBeInTheDocument();
    });

    it('shows the error when the projects cannot be loaded', async () => {
        services.getProjects.mockRejectedValue({message: 'projects down'});
        await mountHub();
        expect(screen.getByText('projects down')).toBeInTheDocument();
    });
});

describe('HubStateUpdates loading the project data', () => {

    it('loads apps, nodes and devices of the project into the state', async () => {
        await mountHub();
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
        expect(services.getAllApps).toHaveBeenCalledWith('p1');
        expect(services.getNodes).toHaveBeenCalledWith('p1');
    });

    it.each([
        ['apps', services.getAllApps],
        ['devices', services.getAllDevices],
    ])('shows the error when the %s cannot be loaded', async (name, failing) => {
        failing.mockRejectedValue({message: `${name} down`});
        await mountHub();
        expect(screen.getByText(`${name} down`)).toBeInTheDocument();
    });

    it('ignores a failing node load', async () => {
        services.getNodes.mockRejectedValue(new Error('x'));
        await mountHub();
        expect(screen.getByText('projects=1 apps=1 nodes=- devices=9:1')).toBeInTheDocument();
    });

    it('drops results that arrive after unmount', async () => {
        let resolveApps: (v: unknown) => void = () => undefined;
        services.getAllApps.mockReturnValue(new Promise(r => {
            resolveApps = r;
        }));
        services.getAllDevices.mockRejectedValue(new Error('late'));
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
        services.getAllDevices.mockClear();
        services.getNodes.mockResolvedValue([{id: 1}]);
        send('nodes');
        send('nodes');
        await flush(100);
        expect(services.getAllDevices).not.toHaveBeenCalled();
        await flush(200);
        expect(services.getAllDevices).toHaveBeenCalledTimes(1);
        expect(screen.getByText('projects=1 apps=1 nodes=1 devices=9:1')).toBeInTheDocument();
    });

    it('reloads the devices after a device list event', async () => {
        await mountHub();
        services.getAllDevices.mockResolvedValue([{id: 10, status: 2}]);
        send('device_list');
        await flush(200);
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=10:2')).toBeInTheDocument();
    });

    it('does not apply a debounced reload that was cancelled', async () => {
        await mountHub();
        services.getAllDevices.mockResolvedValue([{id: 10, status: 2}]);
        send('device_list');
        cleanup();
        await flush(200);
        expect(services.getAllDevices).toHaveBeenCalledTimes(1);
    });

    it('ignores a failing device reload', async () => {
        await mountHub();
        services.getAllDevices.mockRejectedValue(new Error('x'));
        send('device_list');
        await flush(200);
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
    });
});

describe('HubStateUpdates reconnecting and status changes', () => {
    it('reloads after a reconnect but not after the first connect', async () => {
        await mountHub();
        services.getNodes.mockClear();
        act(() => {
            window.dispatchEvent(new Event(SSE_CONNECTED_EVENT));
        });
        expect(services.getNodes).not.toHaveBeenCalled();
        act(() => {
            window.dispatchEvent(new Event(SSE_CONNECTED_EVENT));
        });
        expect(services.getNodes).toHaveBeenCalledTimes(1);
    });

    it('applies device status changes, each once', async () => {
        await mountHub();
        send('devices', {deviceId: 9, deviceState: 4});
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:4')).toBeInTheDocument();
        send('devices', {deviceId: 9, deviceState: 2});
        send('devices', {deviceId: 9, deviceState: 3});
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:3')).toBeInTheDocument();
    });

    it('ignores a device event without data', async () => {
        await mountHub();
        send('devices', null);
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
    });

    it('keeps only the latest backlog of changes', async () => {
        await mountHub();
        for (let i = 0; i < 60; i++) send('devices', {deviceId: 9, deviceState: 2 + (i % 2)});
        expect(screen.getByText('projects=1 apps=1 nodes=2 devices=9:3')).toBeInTheDocument();
    });
});
