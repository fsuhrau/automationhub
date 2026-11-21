import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {ScreenMode} from '../../../types/device.control.video';
import {ScreenSource, SNAPSHOT_DELAY_MS, useScreenSource} from './useScreenSource';

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

const sourceOf = (mode: ScreenMode): RenderHookResult<ScreenSource, unknown> => renderHook(() => useScreenSource({projectId: 'p', deviceId: 7}, mode));

describe('useScreenSource', () => {
    it('streams screenshots in the screenshot mode', () => {
        expect(sourceOf('screenshots').result.current.src).toBe('/api/p/device/7/stream?fps=2');
    });

    it('shows one screenshot when paused, and a new one on refresh', () => {
        const {result} = sourceOf('off');
        const first = result.current.src;
        expect(first).toMatch(/\/screenshot\?t=\d+$/);
        vi.advanceTimersByTime(5);
        act(() => result.current.refresh());
        expect(result.current.src).not.toBe(first);
    });

    it('refreshes shortly after input when paused', () => {
        const {result} = sourceOf('off');
        const first = result.current.src;
        vi.advanceTimersByTime(5);
        act(() => result.current.afterInput());
        expect(result.current.src).toBe(first);
        act(() => {
            vi.advanceTimersByTime(SNAPSHOT_DELAY_MS);
        });
        expect(result.current.src).not.toBe(first);
    });

    it('leaves the screen alone after input in the other modes', () => {
        const {result} = sourceOf('screenshots');
        act(() => {
            result.current.afterInput();
            vi.advanceTimersByTime(SNAPSHOT_DELAY_MS);
        });
        expect(result.current.src).toBe('/api/p/device/7/stream?fps=2');
    });
});
