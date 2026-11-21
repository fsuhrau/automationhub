import {describe, expect, it} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {useDeviceLog} from './useDeviceLog';
import {fakeSources} from './testKit';

describe('useDeviceLog', () => {
    it('opens the stream, collects lines and goes live', () => {
        const {create, sources} = fakeSources();
        const {result} = renderHook(() => useDeviceLog('/log', create));
        expect(sources[0].url).toBe('/log');
        expect(result.current.status).toBe('connecting');
        act(() => sources[0].emit('lines', ['a']));
        act(() => sources[0].emit('lines', ['b']));
        expect(result.current).toMatchObject({lines: ['a', 'b'], status: 'live'});
    });

    it('reports a failure and recovers with the next lines', () => {
        const {create, sources} = fakeSources();
        const {result} = renderHook(() => useDeviceLog('/log', create));
        act(() => sources[0].emit('failure', {message: 'adb is gone'}));
        expect(result.current).toMatchObject({status: 'failing', failure: 'adb is gone'});
        act(() => sources[0].emit('lines', ['x']));
        expect(result.current).toMatchObject({status: 'live', failure: null});
    });

    it('follows the connection: open, drop and reconnect', () => {
        const {create, sources} = fakeSources();
        const {result} = renderHook(() => useDeviceLog('/log', create));
        act(() => sources[0].emit('open'));
        expect(result.current.status).toBe('live');
        act(() => sources[0].emit('error'));
        expect(result.current.status).toBe('connecting');
    });

    it('clears the lines', () => {
        const {create, sources} = fakeSources();
        const {result} = renderHook(() => useDeviceLog('/log', create));
        act(() => sources[0].emit('lines', ['a']));
        act(() => result.current.clear());
        expect(result.current.lines).toEqual([]);
    });

});

describe('useDeviceLog restarts and closes', () => {
    it('starts over for another url and closes the old stream', () => {
        const {create, sources} = fakeSources();
        const {result, rerender} = renderHook(({url}) => useDeviceLog(url, create), {initialProps: {url: '/one'}});
        act(() => sources[0].emit('lines', ['a']));
        rerender({url: '/two'});
        expect(sources[0].closed).toBe(true);
        expect(result.current).toMatchObject({channel: '/two', lines: []});
        act(() => sources[1].emit('lines', ['b']));
        expect(result.current.lines).toEqual(['b']);
    });

    it('keeps a late event of a replaced stream out of the new log', () => {
        const {create, sources} = fakeSources();
        const {result, rerender} = renderHook(({url}) => useDeviceLog(url, create), {initialProps: {url: '/one'}});
        rerender({url: '/two'});
        act(() => sources[0].emit('lines', ['late']));
        expect(result.current.lines).toEqual([]);
        expect(result.current.channel).toBe('/two');
    });

});

describe('useDeviceLog lifecycle', () => {
    it('closes the stream on unmount', () => {
        const {create, sources} = fakeSources();
        renderHook(() => useDeviceLog('/log', create)).unmount();
        expect(sources[0].closed).toBe(true);
    });

    it('opens an EventSource by default', () => {
        const opened: string[] = [];
        class Stub {
            constructor(url: string) {
                opened.push(url);
            }

            addEventListener(): void {
            }

            close(): void {
            }
        }
        const original = globalThis.EventSource;
        globalThis.EventSource = Stub as unknown as typeof EventSource;
        try {
            renderHook(() => useDeviceLog('/default'));
        } finally {
            globalThis.EventSource = original;
        }
        expect(opened).toEqual(['/default']);
    });
});
