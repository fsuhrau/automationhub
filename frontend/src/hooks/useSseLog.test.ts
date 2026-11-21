import {describe, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {SseLog, SseLogOptions, SseSource, useSseLog} from './useSseLog';

interface Entry {
    id: number,
}

class FakeSource implements SseSource {
    closed = false;
    listeners = new Map<string, ((e: Event) => void)[]>();
    constructor(readonly channel: string) {}
    addEventListener = ((type: string, l: (e: Event) => void) => {
        this.listeners.set(type, [...(this.listeners.get(type) ?? []), l]);
    }) as SseSource['addEventListener'];
    close = (): void => {
        this.closed = true;
    };
    emit(type: string, data?: unknown): void {
        const event = data === undefined ? new Event(type) : new MessageEvent(type, {data: JSON.stringify(data)});
        this.listeners.get(type)?.forEach(l => l(event));
    }
}

interface Setup extends RenderHookResult<SseLog<Entry>, { c: string }> {
    sources: FakeSource[],
}

const setup = (channel: string, options: SseLogOptions = {}): Setup => {
    const sources: FakeSource[] = [];
    const createSource = vi.fn((c: string): SseSource => {
        const s = new FakeSource(c);
        sources.push(s);
        return s;
    });
    const hook = renderHook(({c}) => useSseLog<Entry>(c, {createSource, ...options}), {initialProps: {c: channel}});
    return {...hook, sources};
};

describe('useSseLog', () => {
    it('appends entries with new ids, single or as list', () => {
        const {result, sources} = setup('/log');
        act(() => sources[0].emit('message', {id: 1}));
        act(() => sources[0].emit('message', [{id: 1}, {id: 2}]));
        expect(result.current.entries.map(e => e.id)).toEqual([1, 2]);
    });

    it('listens to the configured event and limits the entries', () => {
        const {result, sources} = setup('/log', {event: 'lines', limit: 2});
        act(() => sources[0].emit('message', {id: 9}));
        act(() => sources[0].emit('lines', [{id: 1}, {id: 2}, {id: 3}]));
        expect(result.current.entries.map(e => e.id)).toEqual([2, 3]);
    });

    it('tracks the connection state', () => {
        const {result, sources} = setup('/log');
        expect(result.current.connected).toBe(false);
        act(() => sources[0].emit('open'));
        expect(result.current.connected).toBe(true);
        act(() => sources[0].emit('error'));
        expect(result.current.connected).toBe(false);
    });

    it('closes the source on unmount', () => {
        const {unmount, sources} = setup('/log');
        unmount();
        expect(sources[0].closed).toBe(true);
    });

});

describe('useSseLog channel changes', () => {
    it('restarts with an empty list when the channel changes', () => {
        const {result, rerender, sources} = setup('/a');
        act(() => sources[0].emit('message', {id: 1}));
        rerender({c: '/b'});
        expect(result.current.entries).toEqual([]);
        expect(sources[0].closed).toBe(true);
        expect(sources[1].channel).toBe('/b');
        act(() => sources[1].emit('message', {id: 1}));
        expect(result.current.entries).toEqual([{id: 1}]);
    });

    it('drops events of the old channel after a change', () => {
        const {result, rerender, sources} = setup('/a');
        rerender({c: '/b'});
        act(() => sources[0].emit('message', {id: 5}));
        expect(result.current.entries).toEqual([]);
    });

});

describe('useSseLog default source', () => {
    it('opens an EventSource with credentials by default', () => {
        const created: unknown[] = [];
        class Stub extends FakeSource {
            constructor(url: string, init?: EventSourceInit) {
                super(url);
                created.push([url, init]);
            }
        }
        vi.stubGlobal('EventSource', Stub);
        const {unmount} = renderHook(() => useSseLog('/real'));
        expect(created).toEqual([['/real', {withCredentials: true}]]);
        unmount();
        vi.unstubAllGlobals();
    });
});
