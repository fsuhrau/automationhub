import {afterEach, beforeEach, describe, expect, it, Mock, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import type {KeyboardEvent} from 'react';
import {DeviceInput} from './deviceInput';
import {Typing, TYPE_BATCH_MS, useTyping} from './useTyping';

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

const keyEvent = (key: string, extra = {}): { event: KeyboardEvent, preventDefault: Mock } => {
    const preventDefault = vi.fn();
    return {event: {nativeEvent: {key, ...extra}, preventDefault} as unknown as KeyboardEvent, preventDefault};
};

interface Setup extends RenderHookResult<Typing, unknown> {
    input: DeviceInput,
}

const setup = (hasLease = true): Setup => {
    const input = {sendKey: vi.fn(), typeText: vi.fn()};
    const hook = renderHook(() => useTyping(input, hasLease));
    return {input, ...hook};
};

describe('useTyping', () => {
    it('sends characters as one text once typing pauses', () => {
        const {input, result} = setup();
        const a = keyEvent('a');
        act(() => {
            result.current.onScreenKey(a.event);
            result.current.onScreenKey(keyEvent('b').event);
        });
        expect(a.preventDefault).toHaveBeenCalled();
        expect(input.typeText).not.toHaveBeenCalled();
        act(() => {
            vi.advanceTimersByTime(TYPE_BATCH_MS);
        });
        expect(input.typeText).toHaveBeenCalledWith('ab');
    });

    it('sends what was typed before a named key, then the key', () => {
        const {input, result} = setup();
        act(() => {
            result.current.onScreenKey(keyEvent('a').event);
            result.current.onScreenKey(keyEvent('Enter').event);
        });
        expect(input.typeText).toHaveBeenCalledWith('a');
        expect(input.sendKey).toHaveBeenCalledWith('enter');
    });

    it('sends only the key when nothing was typed', () => {
        const {input, result} = setup();
        act(() => result.current.onScreenKey(keyEvent('Escape').event));
        expect(input.typeText).not.toHaveBeenCalled();
        expect(input.sendKey).toHaveBeenCalledWith('back');
    });
});

describe('useTyping browser keys', () => {
    it('leaves shortcuts to the browser', () => {
        const {input, result} = setup();
        const copy = keyEvent('c', {ctrlKey: true});
        act(() => result.current.onScreenKey(copy.event));
        expect(copy.preventDefault).not.toHaveBeenCalled();
        expect(input.sendKey).not.toHaveBeenCalled();
    });

    it('ignores the keyboard without a lease', () => {
        const {input, result} = setup(false);
        const a = keyEvent('a');
        act(() => result.current.onScreenKey(a.event));
        expect(a.preventDefault).not.toHaveBeenCalled();
        expect(input.typeText).not.toHaveBeenCalled();
    });

    it('flushes on demand, e.g. when the screen loses focus', () => {
        const {input, result} = setup();
        act(() => {
            result.current.onScreenKey(keyEvent('a').event);
            result.current.flush();
        });
        expect(input.typeText).toHaveBeenCalledWith('a');
        act(() => result.current.flush());
        expect(input.typeText).toHaveBeenCalledTimes(1);
    });

    it('drops the timer on unmount', () => {
        const {input, result, unmount} = setup();
        act(() => result.current.onScreenKey(keyEvent('a').event));
        unmount();
        vi.advanceTimersByTime(TYPE_BATCH_MS);
        expect(input.typeText).not.toHaveBeenCalled();
    });
});
