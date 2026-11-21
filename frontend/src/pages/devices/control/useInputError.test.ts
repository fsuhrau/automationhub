import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {renderHook} from '@testing-library/react';
import {INPUT_ERROR_GAP_MS, useInputError} from './useInputError';

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

describe('useInputError', () => {
    it('reports an error and reads the lock status again', () => {
        const report = vi.fn();
        const refresh = vi.fn();
        const {result} = renderHook(() => useInputError(report, refresh));
        result.current('lease lost');
        expect(report).toHaveBeenCalledWith('lease lost');
        expect(refresh).toHaveBeenCalledTimes(1);
    });

    it('shows one error per gap', () => {
        const report = vi.fn();
        const {result} = renderHook(() => useInputError(report, vi.fn()));
        result.current('first');
        vi.advanceTimersByTime(INPUT_ERROR_GAP_MS - 1);
        result.current('second');
        vi.advanceTimersByTime(1);
        result.current('third');
        expect(report.mock.calls.map(c => c[0])).toEqual(['first', 'third']);
    });
});
