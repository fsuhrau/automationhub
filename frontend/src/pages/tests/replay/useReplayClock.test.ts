import {afterEach, beforeEach, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {ReplayClock, TICK_MS, useReplayClock} from './useReplayClock';

const RANGE = {start: 1000, end: 11000};

const setup = (initial = 1000): RenderHookResult<ReplayClock, unknown> => renderHook(() => useReplayClock(RANGE, initial));
const advance = (ms: number): void => {
    act(() => {
        vi.advanceTimersByTime(ms);
    });
};

beforeEach(() => {
    vi.useFakeTimers();
});
afterEach(() => {
    vi.useRealTimers();
});

it('useReplayClock: starts paused at the initial time with rate 1', () => {
    const {result} = setup(3000);
    expect(result.current).toMatchObject({position: 3000, playing: false, rate: 1});
});

it('useReplayClock: does not move while paused', () => {
    const {result} = setup();
    advance(5000);
    expect(result.current.position).toBe(1000);
});

it('useReplayClock: advances with the wall time while playing', () => {
    const {result} = setup();
    act(() => result.current.toggle());
    advance(2000);
    expect(result.current.playing).toBe(true);
    expect(result.current.position).toBe(3000);
});

it('useReplayClock: updates ten times a second', () => {
    const {result} = setup();
    act(() => result.current.toggle());
    advance(TICK_MS - 1);
    expect(result.current.position).toBe(1000);
    advance(1);
    expect(result.current.position).toBe(1000 + TICK_MS);
});

it('useReplayClock: stops where paused', () => {
    const {result} = setup();
    act(() => result.current.toggle());
    advance(1000);
    act(() => result.current.toggle());
    advance(5000);
    expect(result.current).toMatchObject({playing: false, position: 2000});
});

it('useReplayClock: scales the time by the speed', () => {
    const {result} = setup();
    act(() => result.current.setRate(4));
    act(() => result.current.toggle());
    advance(1000);
    expect(result.current.position).toBe(5000);
    act(() => result.current.setRate(0.5));
    advance(1000);
    expect(result.current.position).toBe(5500);
});

it('useReplayClock: seeks within the range, clamping outside it', () => {
    const {result} = setup();
    act(() => result.current.seek(6000));
    expect(result.current.position).toBe(6000);
    act(() => result.current.seek(-5));
    expect(result.current.position).toBe(1000);
    act(() => result.current.seek(99999));
    expect(result.current.position).toBe(11000);
});

it('useReplayClock: continues from the sought time while playing', () => {
    const {result} = setup();
    act(() => result.current.toggle());
    advance(500);
    act(() => result.current.seek(8000));
    advance(1000);
    expect(result.current.position).toBe(9000);
});

it('useReplayClock: stops at the end and plays again from the start', () => {
    const {result} = setup(10500);
    act(() => result.current.toggle());
    advance(2000);
    expect(result.current).toMatchObject({playing: false, position: 11000});
    act(() => result.current.toggle());
    expect(result.current.position).toBe(1000);
    expect(result.current.playing).toBe(true);
});

it('useReplayClock: uses an injected time source', () => {
    let now = 0;
    const {result} = renderHook(() => useReplayClock(RANGE, 1000, () => now));
    act(() => result.current.toggle());
    now = 700;
    advance(TICK_MS);
    expect(result.current.position).toBe(1700);
});
