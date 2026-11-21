import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {useDebouncedValue} from './useDebouncedValue';

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

describe('useDebouncedValue', () => {
    it('follows the value once it stayed unchanged for the delay', () => {
        const {result, rerender} = renderHook(({value}) => useDebouncedValue(value, 300), {initialProps: {value: 'a'}});
        rerender({value: 'b'});
        act(() => vi.advanceTimersByTime(299));
        expect(result.current).toBe('a');
        rerender({value: 'c'});
        act(() => vi.advanceTimersByTime(299));
        expect(result.current).toBe('a');
        act(() => vi.advanceTimersByTime(1));
        expect(result.current).toBe('c');
    });
});
