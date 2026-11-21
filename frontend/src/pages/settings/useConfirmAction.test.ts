import {expect, it, vi} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {AsyncAction} from '../../hooks/useAsyncAction';
import {useConfirmAction} from './useConfirmAction';

const action = (result: boolean | undefined, busy = false): AsyncAction<[string], boolean> =>
    ({run: vi.fn().mockResolvedValue(result), busy, error: null});

it('useConfirmAction: opens for a target and closes when cancelled', () => {
    const {result} = renderHook(() => useConfirmAction(action(true)));
    expect(result.current.target).toBeNull();
    act(() => result.current.ask('a'));
    expect(result.current.target).toBe('a');
    act(() => result.current.cancel());
    expect(result.current.target).toBeNull();
});

it('useConfirmAction: runs the action on confirm and closes when it succeeded', async () => {
    const run = action(true);
    const {result} = renderHook(() => useConfirmAction(run));
    act(() => result.current.ask('a'));
    await act(() => result.current.confirm());
    expect(run.run).toHaveBeenCalledWith('a');
    expect(result.current.target).toBeNull();
});

it('useConfirmAction: stays open when the action failed', async () => {
    const {result} = renderHook(() => useConfirmAction(action(undefined)));
    act(() => result.current.ask('a'));
    await act(() => result.current.confirm());
    expect(result.current.target).toBe('a');
});

it('useConfirmAction: does nothing on confirm without a target and passes busy on', async () => {
    const run = action(true, true);
    const {result} = renderHook(() => useConfirmAction(run));
    await act(() => result.current.confirm());
    expect(run.run).not.toHaveBeenCalled();
    expect(result.current.busy).toBe(true);
});
