import {describe, expect, it} from 'vitest';
import {act, renderHook, waitFor} from '@testing-library/react';
import {useAsyncAction} from './useAsyncAction';

const deferred = <T,>(): { promise: Promise<T>, resolve: (v: T) => void, reject: (e: unknown) => void } => {
    let resolve!: (v: T) => void;
    let reject!: (e: unknown) => void;
    const promise = new Promise<T>((res, rej) => {
        resolve = res;
        reject = rej;
    });
    return {promise, resolve, reject};
};

describe('useAsyncAction', () => {
    it('is busy while running and returns the result', async () => {
        const d = deferred<number>();
        const {result} = renderHook(() => useAsyncAction((x: number) => d.promise.then(v => v + x - x)));
        let outcome: Promise<number | undefined> = Promise.resolve(undefined);
        act(() => {
            outcome = result.current.run(1);
        });
        expect(result.current.busy).toBe(true);
        await act(async () => d.resolve(7));
        expect(await outcome).toBe(7);
        expect(result.current.busy).toBe(false);
        expect(result.current.error).toBeNull();
    });

    it('keeps the failure as text and clears it on the next run', async () => {
        let fail = true;
        const {result} = renderHook(() => useAsyncAction(async () => {
            if (fail) throw new Error('nope');
            return 'ok';
        }));
        await act(async () => expect(await result.current.run()).toBeUndefined());
        expect(result.current.error).toBe('nope');
        expect(result.current.busy).toBe(false);
        fail = false;
        await act(async () => expect(await result.current.run()).toBe('ok'));
        expect(result.current.error).toBeNull();
    });

    it('ignores the outcome after unmount', async () => {
        const d = deferred<number>();
        const {result, unmount} = renderHook(() => useAsyncAction(() => d.promise));
        let outcome: Promise<number | undefined> = Promise.resolve(undefined);
        act(() => {
            outcome = result.current.run();
        });
        unmount();
        d.reject(new Error('late'));
        await waitFor(async () => expect(await outcome).toBeUndefined());
    });
});
