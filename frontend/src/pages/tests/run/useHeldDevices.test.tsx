import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {fakeApi} from './testkit/fakes';
import {wrapperOf} from './testkit/wrapper';
import {useHeldDevices} from './useHeldDevices';

const run = {projectId: 'p1', appId: 1, testId: 2, runId: 3};
const key = 'GET /p1/app/1/test/2/run/3';

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

const poll = async (): Promise<void> => {
    await act(async () => {
        await vi.advanceTimersByTimeAsync(3000);
    });
};

describe('useHeldDevices', () => {
    it('asks again until the running run holds devices', async () => {
        const api = fakeApi({[key]: {devices: [7]}});
        const {result} = renderHook(() => useHeldDevices([], true, run), {wrapper: wrapperOf({api})});
        expect(result.current).toEqual([]);
        await poll();
        expect(result.current).toEqual([7]);
        await poll();
        expect(api.requests).toHaveLength(1);
    });

    it('treats an answer without devices as none held', async () => {
        const api = fakeApi({[key]: {}});
        const {result} = renderHook(() => useHeldDevices([], true, run), {wrapper: wrapperOf({api})});
        await poll();
        expect(result.current).toEqual([]);
        expect(api.requests).toHaveLength(1);
    });

    it('keeps polling after a failed request', async () => {
        const api = fakeApi({[key]: new Error('down')});
        renderHook(() => useHeldDevices([], true, run), {wrapper: wrapperOf({api})});
        await poll();
        await poll();
        expect(api.requests).toHaveLength(2);
    });

    it.each([
        ['the master does not say which devices are held', undefined, true],
        ['the run does not run', [], false],
        ['devices are held', [7], true],
    ])('does not ask when %s', async (_name, initial, running) => {
        const api = fakeApi({[key]: {devices: [1]}});
        const {result} = renderHook(() => useHeldDevices(initial, running, run), {wrapper: wrapperOf({api})});
        await poll();
        expect(api.requests).toHaveLength(0);
        expect(result.current).toEqual(initial);
    });
});
