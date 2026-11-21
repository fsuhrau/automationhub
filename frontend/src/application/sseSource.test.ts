import {afterEach, describe, expect, it, vi} from 'vitest';
import {createHubEventSource, SSE_CONNECTED_EVENT} from './sseSource';

class FakeEventSource {
    static instances: FakeEventSource[] = [];
    listeners: Record<string, () => void> = {};

    constructor(public url: string) {
        FakeEventSource.instances.push(this);
    }

    addEventListener(type: string, listener: () => void): void {
        this.listeners[type] = listener;
    }
}

afterEach(() => {
    vi.unstubAllGlobals();
    FakeEventSource.instances = [];
});

describe('createHubEventSource', () => {
    it('opens an event source on the endpoint', () => {
        vi.stubGlobal('EventSource', FakeEventSource);
        const source = createHubEventSource('/api/sse')();
        expect(source).toBe(FakeEventSource.instances[0]);
        expect(FakeEventSource.instances[0].url).toBe('/api/sse');
    });

    it('tells the window whenever the stream (re)connects', () => {
        vi.stubGlobal('EventSource', FakeEventSource);
        const connected = vi.fn();
        window.addEventListener(SSE_CONNECTED_EVENT, connected);
        createHubEventSource('/api/sse')();
        FakeEventSource.instances[0].listeners.open();
        FakeEventSource.instances[0].listeners.open();
        window.removeEventListener(SSE_CONNECTED_EVENT, connected);
        expect(connected).toHaveBeenCalledTimes(2);
    });
});
