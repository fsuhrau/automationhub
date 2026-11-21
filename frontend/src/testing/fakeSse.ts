import {act} from '@testing-library/react';
import {vi} from 'vitest';
import {SseSource, SseSourceFactory} from '../hooks/useSseLog';

type Listener = (event: MessageEvent) => void;

// FakeSse stands in for the hub's event stream: pages subscribe through the
// SSEProvider (`source`), tests push an event with emit. Data is sent as JSON.
export class FakeSse {
    private readonly listeners = new Map<string, Set<Listener>>();

    // the source the SSEProvider asks for
    readonly source = (): EventSource => ({
        addEventListener: (name: string, listener: Listener) => this.listenersOf(name).add(listener),
        removeEventListener: (name: string, listener: Listener) => this.listenersOf(name).delete(listener),
        close: () => undefined,
    }) as unknown as EventSource;

    emit(name: string, data: unknown = {}): void {
        act(() => this.listenersOf(name).forEach(listener => listener({data: JSON.stringify(data)} as MessageEvent)));
    }

    private listenersOf(name: string): Set<Listener> {
        const known = this.listeners.get(name) ?? new Set<Listener>();
        this.listeners.set(name, known);
        return known;
    }
}

// makes `new EventSource(url)` of the code under test return the fake stream,
// for components that create their own SSEProvider; undo with vi.unstubAllGlobals
export const stubEventSource = (sse: FakeSse): void => {
    vi.stubGlobal('EventSource', function EventSourceStub() {
        return sse.source();
    });
};

// a stream a component opened itself (useSseLog, useDeviceLog), not through the SSEProvider
export class FakeSource {
    closed = false;
    private listeners: Record<string, Listener[]> = {};

    constructor(readonly url: string) {
    }

    addEventListener(type: string, listener: EventListenerOrEventListenerObject): void {
        (this.listeners[type] ??= []).push(listener as Listener);
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

// fakeSources is the createSource of the hooks that open their own stream
export const fakeSources = (): FakeSources => {
    const sources: FakeSource[] = [];
    const create: SseSourceFactory = url => {
        const source = new FakeSource(url);
        sources.push(source);
        return source as unknown as SseSource;
    };
    return {create: vi.fn(create), sources};
};
