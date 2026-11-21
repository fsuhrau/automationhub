import {act, cleanup, render, screen} from '@testing-library/react';
import React from 'react';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {ErrorProvider} from '../ErrorProvider';
import {HubStateProvider, useHubState} from '../hooks/HubStateProvider';
import HubStateUpdates from './HubStateUpdates';
import {SSE_CONNECTED_EVENT} from './sseSource';

// the event stream is faked: send(channel, data) feeds the stateReducer the
// component registered, like the real hook does for a server event
const channels = vi.hoisted(() => new Map<string, (data: unknown) => void>());
const services = vi.hoisted(() => ({
    getAllDevices: vi.fn(),
    getAllApps: vi.fn(),
    getNodes: vi.fn(),
}));

vi.mock('react-hooks-sse', async () => {
    const react = await import('react');
    return {
        useSSE: <S, A>(channel: string, initial: S, options: { stateReducer: (s: S, a: { data: A }) => S }): S => {
            const [state, dispatch] = react.useReducer(options.stateReducer, initial);
            channels.set(channel, (data) => dispatch({data: data as A}));
            return state;
        },
    };
});
vi.mock('../hooks/ProjectProvider', () => ({useProjectContext: () => ({projectIdentifier: 'p1'})}));
vi.mock('../services/device.service', () => ({getAllDevices: services.getAllDevices}));
vi.mock('../services/app.service', () => ({getAllApps: services.getAllApps}));
vi.mock('../services/settings.service', () => ({getNodes: services.getNodes}));

const send = (channel: string, data: unknown = {}): void => {
    const sink = channels.get(channel);
    if (!sink) throw new Error(`no listener on ${channel}`);
    act(() => sink(data));
};

const Probe: React.FC = () => {
    const {state} = useHubState();
    const devices = (state.devices ?? []).map(d => `${d.id}:${d.status}`).join(',');
    return <p>{`apps=${state.apps?.length ?? '-'} nodes=${state.nodes?.length ?? '-'} devices=${devices || '-'}`}</p>;
};

const mountHub = async (): Promise<void> => {
    render(<ErrorProvider><HubStateProvider><HubStateUpdates><Probe/></HubStateUpdates></HubStateProvider></ErrorProvider>);
    await act(async () => undefined);
};

const flush = async (ms = 0): Promise<void> => {
    await act(async () => {
        await vi.advanceTimersByTimeAsync(ms);
    });
};

beforeEach(() => {
    vi.useFakeTimers();
    channels.clear();
    services.getAllApps.mockResolvedValue([{id: 1}]);
    services.getNodes.mockResolvedValue([{id: 1}, {id: 2}]);
    services.getAllDevices.mockResolvedValue([{id: 9, status: 1}]);
});
afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.clearAllMocks();
});

describe('HubStateUpdates loading', () => {
    it('loads apps, nodes and devices of the project into the state', async () => {
        await mountHub();
        expect(screen.getByText('apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
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
        expect(screen.getByText('apps=1 nodes=- devices=9:1')).toBeInTheDocument();
    });

    it('drops results that arrive after unmount', async () => {
        let resolveApps: (v: unknown) => void = () => undefined;
        services.getAllApps.mockReturnValue(new Promise(r => {
            resolveApps = r;
        }));
        services.getAllDevices.mockRejectedValue(new Error('late'));
        render(<ErrorProvider><HubStateProvider><HubStateUpdates><Probe/></HubStateUpdates></HubStateProvider></ErrorProvider>);
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
        expect(screen.getByText('apps=1 nodes=1 devices=9:1')).toBeInTheDocument();
    });

    it('reloads the devices after a device list event', async () => {
        await mountHub();
        services.getAllDevices.mockResolvedValue([{id: 10, status: 2}]);
        send('device_list');
        await flush(200);
        expect(screen.getByText('apps=1 nodes=2 devices=10:2')).toBeInTheDocument();
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
        expect(screen.getByText('apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
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
        expect(screen.getByText('apps=1 nodes=2 devices=9:4')).toBeInTheDocument();
        send('devices', {deviceId: 9, deviceState: 2});
        send('devices', {deviceId: 9, deviceState: 3});
        expect(screen.getByText('apps=1 nodes=2 devices=9:3')).toBeInTheDocument();
    });

    it('ignores a device event without data', async () => {
        await mountHub();
        send('devices', null);
        expect(screen.getByText('apps=1 nodes=2 devices=9:1')).toBeInTheDocument();
    });

    it('keeps only the latest backlog of changes', async () => {
        await mountHub();
        for (let i = 0; i < 60; i++) send('devices', {deviceId: 9, deviceState: 2 + (i % 2)});
        expect(screen.getByText('apps=1 nodes=2 devices=9:3')).toBeInTheDocument();
    });
});
