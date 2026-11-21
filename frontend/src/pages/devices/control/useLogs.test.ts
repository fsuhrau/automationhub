import {describe, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {fakeService} from './fakes';
import {Runner} from './useGuardedAction';
import {DeviceControlApi} from '../../../types/device.control.api';
import {LogsState, useLogs} from './useLogs';

const run: Runner = async action => {
    await action();
};

const LINES = ['I/App( 1): hello', 'E/App( 1): boom'];

type Setup = RenderHookResult<LogsState, unknown> & { service: DeviceControlApi };

const setup = (): Setup => {
    const service = fakeService({logs: vi.fn(async () => LINES)});
    return {service, ...renderHook(() => useLogs(service, run))};
};

describe('useLogs', () => {
    it('has no log before loading', () => {
        const {result} = setup();
        expect(result.current).toMatchObject({logs: null, visible: [], lines: 300});
    });

    it('loads the chosen number of lines', async () => {
        const {result, service} = setup();
        act(() => result.current.setLines(1000));
        await act(async () => result.current.load());
        expect(service.logs).toHaveBeenCalledWith(1000);
        expect(result.current.visible).toEqual(LINES);
    });

    it('filters by text', async () => {
        const {result} = setup();
        await act(async () => result.current.load());
        act(() => result.current.setFilter('hello'));
        expect(result.current.visible).toEqual([LINES[0]]);
    });

    it('can keep only warnings and errors', async () => {
        const {result} = setup();
        await act(async () => result.current.load());
        act(() => result.current.toggleProblems());
        expect(result.current.problemsOnly).toBe(true);
        expect(result.current.visible).toEqual([LINES[1]]);
        act(() => result.current.toggleProblems());
        expect(result.current.visible).toEqual(LINES);
    });
});
