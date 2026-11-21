import {describe, expect, it} from 'vitest';
import {classifyGesture, mapToDevice, scrollAmount, TAP_DISTANCE, videoDeviceSize} from './screenModel';

describe('videoDeviceSize', () => {
    const portrait = {width: 360, height: 640};

    it('takes the canvas size without stream meta', () => {
        expect(videoDeviceSize(portrait, null)).toEqual(portrait);
        expect(videoDeviceSize(portrait, {screenWidth: 0, screenHeight: 0})).toEqual(portrait);
    });

    it('uses the screen size in the orientation of the frame', () => {
        const meta = {screenWidth: 1080, screenHeight: 2400};
        expect(videoDeviceSize(portrait, meta)).toEqual({width: 1080, height: 2400});
        expect(videoDeviceSize({width: 640, height: 360}, meta)).toEqual({width: 2400, height: 1080});
    });

    it('keeps a landscape screen in a landscape frame', () => {
        expect(videoDeviceSize({width: 640, height: 360}, {screenWidth: 2400, screenHeight: 1080})).toEqual({width: 2400, height: 1080});
    });
});

describe('mapToDevice', () => {
    const rect = {left: 10, top: 10, width: 100, height: 200};

    it('scales to device pixels', () => {
        expect(mapToDevice(rect, {width: 1000, height: 2000}, {x: 60, y: 110})).toEqual({x: 500, y: 1000});
    });

    it('keeps element pixels without a device size', () => {
        expect(mapToDevice(rect, {width: 0, height: 0}, {x: 60, y: 110})).toEqual({x: 50, y: 100});
    });
});

describe('classifyGesture', () => {
    const identity = (p: { x: number, y: number }): { x: number, y: number } => ({x: p.x, y: p.y});
    const start = {x: 0, y: 0, t: 1000};

    it('is a tap for a short move', () => {
        expect(classifyGesture(start, {x: TAP_DISTANCE - 1, y: 0, t: 1100}, identity)).toEqual({kind: 'tap', at: {x: 0, y: 0}});
    });

    it('is a swipe for a long move, with the time it took', () => {
        expect(classifyGesture(start, {x: 50, y: 60, t: 1500}, identity))
            .toEqual({kind: 'swipe', from: {x: 0, y: 0}, to: {x: 50, y: 60}, durationMs: 500});
    });

    it.each([[1010, 100], [9000, 2000]])('limits the duration (took until %i)', (t, duration) => {
        expect(classifyGesture(start, {x: 50, y: 0, t}, identity)).toMatchObject({durationMs: duration});
    });
});

describe('scrollAmount', () => {
    it.each([
        [100, false, -1],
        [-300, false, 3],
        [3, true, -3],
        [100000, false, -16],
        [-100000, false, 16],
    ])('turns %i (lines: %s) into %i', (delta, lines, amount) => {
        expect(scrollAmount(delta, lines)).toBe(amount);
    });
});
