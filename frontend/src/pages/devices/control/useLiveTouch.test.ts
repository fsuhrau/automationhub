import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {FakePlayer} from './fakes';
import {pointer, screenElement} from './pointerFakes';
import {LiveTouch, useLiveTouch} from './useLiveTouch';

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

interface Setup extends RenderHookResult<LiveTouch, unknown> {
    el: HTMLCanvasElement,
    player: FakePlayer,
}

const setup = (): Setup => {
    const {el} = screenElement();
    const player = new FakePlayer('', el, {} as never);
    const hook = renderHook(() => useLiveTouch('tok'));
    return {el, player, ...hook};
};

const kinds = (player: FakePlayer): string[] => player.sent.map(s => (s.input as { action: string }).action);

describe('useLiveTouch', () => {
    it('does not start without a live player', () => {
        const {el, result} = setup();
        expect(result.current.begin(pointer(el, {x: 20, y: 20}), null)).toBe(false);
    });

    it('sends down, moves and up with the lease', () => {
        const {el, player, result} = setup();
        expect(result.current.begin(pointer(el, {x: 20, y: 20}), player)).toBe(true);
        expect(el.setPointerCapture).toHaveBeenCalledWith(1);
        act(() => result.current.move(pointer(el, {x: 30, y: 30})));
        act(() => {
            vi.advanceTimersByTime(20);
        });
        expect(result.current.end(pointer(el, {x: 40, y: 40}), 'up')).toBe(true);
        expect(kinds(player)).toEqual(['down', 'move', 'up']);
        expect(player.sent.every(s => s.lease === 'tok')).toBe(true);
    });

    it('sends at most one move per frame, from the latest position', () => {
        const {el, player, result} = setup();
        result.current.begin(pointer(el, {x: 20, y: 20}), player);
        act(() => {
            result.current.move(pointer(el, {x: 30, y: 30}));
            result.current.move(pointer(el, {x: 50, y: 60}));
            vi.advanceTimersByTime(20);
        });
        expect(player.sent.slice(1).map(s => (s.input as { x: number }).x)).toEqual([50]);
    });

    it('drops a pending move when the touch ends', () => {
        const {el, player, result} = setup();
        result.current.begin(pointer(el, {x: 20, y: 20}), player);
        act(() => result.current.move(pointer(el, {x: 30, y: 30})));
        result.current.end(pointer(el, {x: 30, y: 30}), 'cancel');
        act(() => {
            vi.advanceTimersByTime(20);
        });
        expect(kinds(player)).toEqual(['down', 'cancel']);
    });
});

describe('useLiveTouch other pointers', () => {
    it('does not start a second touch', () => {
        const {el, player, result} = setup();
        result.current.begin(pointer(el, {x: 20, y: 20}), player);
        expect(result.current.begin(pointer(el, {x: 20, y: 20}, {id: 2}), player)).toBe(false);
    });

    it('ignores other pointers and moves without a touch', () => {
        const {el, player, result} = setup();
        act(() => result.current.move(pointer(el, {x: 1, y: 1})));
        expect(result.current.end(pointer(el, {x: 1, y: 1}), 'up')).toBe(false);
        result.current.begin(pointer(el, {x: 20, y: 20}), player);
        act(() => result.current.move(pointer(el, {x: 1, y: 1}, {id: 9})));
        expect(result.current.end(pointer(el, {x: 1, y: 1}, {id: 9}), 'up')).toBe(false);
        expect(kinds(player)).toEqual(['down']);
    });

    it('sends nothing when the touch ended before the frame', () => {
        const {el, player, result} = setup();
        result.current.begin(pointer(el, {x: 20, y: 20}), player);
        act(() => result.current.move(pointer(el, {x: 30, y: 30})));
        act(() => {
            result.current.end(pointer(el, {x: 30, y: 30}), 'up');
            vi.advanceTimersByTime(20);
        });
        expect(kinds(player)).toEqual(['down', 'up']);
    });
});
