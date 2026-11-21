import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {FakePlayer, fakeService} from './fakes';
import {pointer, screenElement} from './pointerFakes';
import {DeviceControlApi} from '../../../types/device.control.api';
import {Runner} from './useGuardedAction';
import {PointerDeps, PointerHandlers, usePointerInput} from './usePointerInput';

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

interface Setup extends RenderHookResult<PointerHandlers, unknown> {
    el: HTMLCanvasElement,
    parent: HTMLElement,
    player: FakePlayer,
    service: DeviceControlApi,
    run: Runner,
    deps: PointerDeps,
}

const setup = (over: Partial<PointerDeps> = {}, live = false): Setup => {
    const {el, parent} = screenElement();
    const player = new FakePlayer('', el, {} as never);
    const service = fakeService();
    const run: Runner = vi.fn(async (action: () => Promise<unknown>) => {
        await action();
    });
    const deps: PointerDeps = {
        lease: 'tok', meta: null, service, run, onPick: null, onGesture: vi.fn(), liveInput: () => live ? player : null, ...over,
    };
    const hook = renderHook(() => usePointerInput(deps));
    return {el, parent, player, service, run, deps, ...hook};
};

describe('usePointerInput over the REST API', () => {
    it('taps at the device position when released close to where it was pressed', () => {
        const {el, result, service, deps} = setup();
        result.current.onPointerDown(pointer(el, {x: 60, y: 110}));
        result.current.onPointerUp(pointer(el, {x: 61, y: 110}));
        // the canvas is 360x640 device pixels, shown at 100x200
        expect(service.tap).toHaveBeenCalledWith('tok', {x: 180, y: 320});
        expect(deps.onGesture).toHaveBeenCalled();
    });

    it('swipes with the time it took when dragged', () => {
        const {el, result, service} = setup();
        result.current.onPointerDown(pointer(el, {x: 20, y: 20}));
        vi.advanceTimersByTime(300);
        result.current.onPointerUp(pointer(el, {x: 60, y: 110}));
        expect(service.swipe).toHaveBeenCalledWith('tok', expect.objectContaining({
            from: {x: 36, y: 32}, to: {x: 180, y: 320}, durationMs: 300,
        }));
    });

    it('uses the natural size of a screenshot', () => {
        const {parent, result, service} = setup();
        const img = document.createElement('img');
        parent.appendChild(img);
        Object.defineProperty(img, 'naturalWidth', {value: 1000});
        Object.defineProperty(img, 'naturalHeight', {value: 2000});
        img.getBoundingClientRect = () => ({left: 0, top: 0, width: 100, height: 200}) as DOMRect;
        result.current.onPointerDown(pointer(img, {x: 50, y: 100}));
        result.current.onPointerUp(pointer(img, {x: 50, y: 100}));
        expect(service.tap).toHaveBeenCalledWith('tok', {x: 500, y: 1000});
    });
});

describe('usePointerInput press handling', () => {
    it('focuses the screen on press and keeps the browser from selecting', () => {
        const {el, parent, result} = setup();
        const down = pointer(el, {x: 20, y: 20});
        result.current.onPointerDown(down);
        expect(down.preventDefault).toHaveBeenCalled();
        expect(document.activeElement).toBe(parent);
    });

    it('sends nothing without a lease', () => {
        const {el, result, run} = setup({lease: null});
        result.current.onPointerDown(pointer(el, {x: 20, y: 20}));
        result.current.onPointerUp(pointer(el, {x: 20, y: 20}));
        expect(run).not.toHaveBeenCalled();
    });

    it('sends nothing for a release without press, or after a cancel', () => {
        const {el, result, run} = setup();
        result.current.onPointerUp(pointer(el, {x: 20, y: 20}));
        result.current.onPointerDown(pointer(el, {x: 20, y: 20}));
        result.current.onPointerCancel(pointer(el, {x: 20, y: 20}));
        result.current.onPointerUp(pointer(el, {x: 20, y: 20}));
        expect(run).not.toHaveBeenCalled();
    });
});

describe('usePointerInput with live input', () => {
    it('sends the touch over the video socket, not through the REST API', () => {
        const {el, player, result, service} = setup({}, true);
        result.current.onPointerDown(pointer(el, {x: 20, y: 20}));
        act(() => result.current.onPointerMove(pointer(el, {x: 30, y: 30})));
        act(() => {
            vi.advanceTimersByTime(20);
        });
        result.current.onPointerUp(pointer(el, {x: 30, y: 30}));
        expect(player.sent.map(s => (s.input as { action: string }).action)).toEqual(['down', 'move', 'up']);
        expect(service.tap).not.toHaveBeenCalled();
    });

    it('cancels a live touch', () => {
        const {el, player, result} = setup({}, true);
        result.current.onPointerDown(pointer(el, {x: 20, y: 20}));
        result.current.onPointerCancel(pointer(el, {x: 20, y: 20}));
        expect(player.sent.map(s => (s.input as { action: string }).action)).toEqual(['down', 'cancel']);
    });
});

describe('usePointerInput while picking', () => {
    it('picks instead of touching the device', () => {
        const onPick = vi.fn();
        const {el, player, result, service} = setup({onPick}, true);
        result.current.onPointerDown(pointer(el, {x: 20, y: 30}));
        result.current.onPointerUp(pointer(el, {x: 20, y: 30}));
        expect(onPick).toHaveBeenCalledWith(el, {x: 20, y: 30});
        expect(player.sent).toHaveLength(0);
        expect(service.tap).not.toHaveBeenCalled();
    });
});
