import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {useNow} from './useNow';

describe('useNow', () => {
    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(1000);
    });
    afterEach(() => vi.useRealTimers());

    it('starts with the current time and ticks on the interval', () => {
        const {result} = renderHook(() => useNow(500));
        expect(result.current).toBe(1000);
        act(() => vi.advanceTimersByTime(500));
        expect(result.current).toBe(1500);
        act(() => vi.advanceTimersByTime(1000));
        expect(result.current).toBe(2500);
    });

    it('does not tick while inactive', () => {
        const {result} = renderHook(() => useNow(500, false));
        act(() => vi.advanceTimersByTime(2000));
        expect(result.current).toBe(1000);
    });

    it('stops the timer on unmount', () => {
        const {unmount} = renderHook(() => useNow(500));
        unmount();
        expect(vi.getTimerCount()).toBe(0);
    });
});
