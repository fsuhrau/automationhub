import {describe, expect, it, vi} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {useGuardedAction} from './useGuardedAction';

describe('useGuardedAction', () => {
    it('runs the action', async () => {
        const onError = vi.fn();
        const {result} = renderHook(() => useGuardedAction(onError));
        const action = vi.fn(async () => 1);
        await act(() => result.current.run(action));
        expect(action).toHaveBeenCalled();
        expect(onError).not.toHaveBeenCalled();
    });

    it('is busy while an action runs', async () => {
        const {result} = renderHook(() => useGuardedAction(vi.fn()));
        let finish = (): void => undefined;
        let pending: Promise<unknown> = Promise.resolve();
        act(() => {
            pending = result.current.run(() => new Promise<void>(resolve => {
                finish = resolve;
            }));
        });
        expect(result.current.busy).toBe(true);
        await act(async () => {
            finish();
            await pending;
        });
        expect(result.current.busy).toBe(false);
    });
});

describe('useGuardedAction failures', () => {
    it('sends a failure to the default handler and never rejects', async () => {
        const onError = vi.fn();
        const {result} = renderHook(() => useGuardedAction(onError));
        const failure = new Error('boom');
        await act(() => result.current.run(() => Promise.reject(failure)));
        expect(onError).toHaveBeenCalledWith(failure);
    });

    it('sends a failure to the handler of the call when there is one', async () => {
        const onError = vi.fn();
        const own = vi.fn();
        const {result} = renderHook(() => useGuardedAction(onError));
        await act(() => result.current.run(() => Promise.reject(new Error('x')), own));
        expect(own).toHaveBeenCalled();
        expect(onError).not.toHaveBeenCalled();
    });
});
