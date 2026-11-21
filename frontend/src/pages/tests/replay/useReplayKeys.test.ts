import {afterEach, expect, it, vi} from 'vitest';
import {fireEvent, renderHook} from '@testing-library/react';
import {useReplayKeys} from './useReplayKeys';

interface Clock {
    position: number,
    seek: ReturnType<typeof vi.fn>,
    toggle: ReturnType<typeof vi.fn>,
}

const setup = (): { clock: Clock, unmount: () => void } => {
    const clock = {position: 20000, seek: vi.fn(), toggle: vi.fn()};
    const view = renderHook(() => useReplayKeys(clock));
    return {clock, unmount: view.unmount};
};

afterEach(() => {
    document.body.innerHTML = '';
});

it('useReplayKeys: toggles on space and skips 5 seconds with the arrows', () => {
    const {clock} = setup();
    fireEvent.keyDown(document.body, {key: ' '});
    fireEvent.keyDown(document.body, {key: 'ArrowLeft'});
    fireEvent.keyDown(document.body, {key: 'ArrowRight'});
    expect(clock.toggle).toHaveBeenCalledTimes(1);
    expect(clock.seek.mock.calls).toEqual([[15000], [25000]]);
});

it('useReplayKeys: ignores other keys', () => {
    const {clock} = setup();
    fireEvent.keyDown(document.body, {key: 'a'});
    expect(clock.toggle).not.toHaveBeenCalled();
    expect(clock.seek).not.toHaveBeenCalled();
});

it('useReplayKeys: leaves keys to buttons and inputs', () => {
    const {clock} = setup();
    const button = document.body.appendChild(document.createElement('button'));
    fireEvent.keyDown(button, {key: ' '});
    expect(clock.toggle).not.toHaveBeenCalled();
});

it('useReplayKeys: stops listening on unmount', () => {
    const {clock, unmount} = setup();
    unmount();
    fireEvent.keyDown(document.body, {key: ' '});
    expect(clock.toggle).not.toHaveBeenCalled();
});
