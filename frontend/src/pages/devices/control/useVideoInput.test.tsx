import React, {useEffect} from 'react';
import {describe, expect, it, vi} from 'vitest';
import {act, fireEvent, render, RenderResult} from '@testing-library/react';
import {ScreenMode} from '../../../types/device.control.video';
import {fakeSupport, FakeSupport} from './fakes';
import {useVideoInput, VideoInput, VideoInputOptions} from './useVideoInput';

const Harness: React.FC<{ options: VideoInputOptions, onInput: (v: VideoInput) => void }> = ({options, onInput}) => {
    const {canvasRef, status, meta, liveInput} = useVideoInput(options);
    useEffect(() => onInput({canvasRef, status, meta, liveInput}));
    return <canvas ref={canvasRef} data-testid="canvas"/>;
};

interface Setup {
    support: FakeSupport,
    options: VideoInputOptions,
    view: RenderResult,
    current: () => VideoInput,
}

const setup = (over: Partial<VideoInputOptions> = {}, mode: ScreenMode = 'video'): Setup => {
    const support: FakeSupport = fakeSupport();
    let input: VideoInput | undefined;
    const options: VideoInputOptions = {
        mode, target: {projectId: 'p', deviceId: 7}, support, lease: 'tok', onUnsupported: vi.fn(), onInputError: vi.fn(), ...over,
    };
    const view = render(<Harness options={options} onInput={i => {
        input = i;
    }}/>);
    return {support, options, view, current: () => input as VideoInput};
};

describe('useVideoInput', () => {
    it('plays the video of the device in the video mode', () => {
        const {support} = setup();
        expect(support.players[0].url).toMatch(/^wss?:\/\/.*\/api\/p\/device\/7\/video\?maxSize=1280&maxFps=30$/);
    });

    it('plays nothing in the other modes', () => {
        expect(setup({}, 'screenshots').support.players).toHaveLength(0);
    });

    it('offers the player for live input when it takes input and there is a lease', () => {
        const {support, current} = setup();
        expect(current().liveInput()).toBe(support.players[0]);
    });

    it.each([
        ['there is no lease', {lease: null}, true],
        ['the player takes no input', {}, false],
    ])('offers no player when %s', (_, over, canSend) => {
        const {support, current} = setup(over);
        support.players[0].canSendInput = canSend;
        expect(current().liveInput()).toBeNull();
    });

    it('offers no player outside the video mode', () => {
        expect(setup({}, 'off').current().liveInput()).toBeNull();
    });
});

describe('useVideoInput statuses', () => {
    it('says what to do when the device has no video, and when it has no reason', () => {
        const {support, options} = setup();
        act(() => support.players[0].callbacks.onStatus({state: 'unsupported', message: 'WDA failed'}));
        expect(options.onUnsupported).toHaveBeenLastCalledWith('Live video unavailable, showing screenshots: WDA failed');
        act(() => support.players[0].callbacks.onStatus({state: 'unsupported'}));
        expect(options.onUnsupported).toHaveBeenLastCalledWith(null);
    });

    it('ignores other statuses', () => {
        const {support, options} = setup();
        act(() => support.players[0].callbacks.onStatus({state: 'playing'}));
        expect(options.onUnsupported).not.toHaveBeenCalled();
    });

    it('passes input errors of the socket on', () => {
        const {support, options} = setup();
        support.players[0].callbacks.onInputError('lease lost');
        expect(options.onInputError).toHaveBeenCalledWith('lease lost');
    });
});

describe('useVideoInput wheel', () => {
    const wheel = (canvas: HTMLElement, init: WheelEventInit): WheelEvent => {
        const event = new WheelEvent('wheel', {cancelable: true, bubbles: true, ...init});
        act(() => {
            canvas.dispatchEvent(event);
        });
        return event;
    };
    const canvasOf = (view: RenderResult): HTMLElement => view.container.querySelector('canvas') as HTMLElement;

    it('scrolls on the device and keeps the page still', () => {
        const {support, view} = setup();
        const event = wheel(canvasOf(view), {deltaY: 100, deltaX: -200, clientX: 5, clientY: 6});
        expect(event.defaultPrevented).toBe(true);
        expect(support.players[0].sent).toEqual([{
            lease: 'tok',
            input: {kind: 'scroll', x: 5, y: 6, width: 360, height: 640, scrollX: 2, scrollY: -1},
        }]);
    });

    it('counts line deltas as lines', () => {
        const {support, view} = setup();
        wheel(canvasOf(view), {deltaY: 2, deltaMode: WheelEvent.DOM_DELTA_LINE});
        expect((support.players[0].sent[0].input as { scrollY: number }).scrollY).toBe(-2);
    });

    it('leaves the wheel to the page without a lease or when the player takes no input', () => {
        const noLease = setup({lease: null});
        expect(wheel(canvasOf(noLease.view), {deltaY: 100}).defaultPrevented).toBe(false);
        const noInput = setup();
        noInput.support.players[0].canSendInput = false;
        expect(wheel(canvasOf(noInput.view), {deltaY: 100}).defaultPrevented).toBe(false);
        expect(noLease.support.players[0].sent).toHaveLength(0);
    });

    it('stops listening outside the video mode', () => {
        const {support, view} = setup({}, 'off');
        fireEvent.wheel(canvasOf(view), {deltaY: 100});
        expect(support.players).toHaveLength(0);
    });
});
