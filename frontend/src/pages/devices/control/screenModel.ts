// Where a pointer is on the device screen and what a drag means.
import {Point} from '../../../types/device.control.api';
import {Rect} from './sceneModel';

export interface Size {
    width: number,
    height: number,
}

export interface ScreenSize {
    screenWidth: number,
    screenHeight: number,
}

export const TAP_DISTANCE = 12; // css px; a longer drag is a swipe
const MIN_SWIPE_MS = 100;
const MAX_SWIPE_MS = 2000;

// videoDeviceSize is the device size in pixels of a video: the stream may be
// scaled, so it takes the screen size from the stream meta, turned to the
// orientation of the current frame
export const videoDeviceSize = (canvas: Size, meta: ScreenSize | null): Size => {
    const sw = meta?.screenWidth ?? 0;
    const sh = meta?.screenHeight ?? 0;
    if (!sw || !sh) {
        return {width: canvas.width, height: canvas.height};
    }
    const landscape = canvas.width > canvas.height;
    return landscape === (sw > sh) ? {width: sw, height: sh} : {width: sh, height: sw};
};

// mapToDevice maps a pointer position on the screen element to device pixels
export const mapToDevice = (rect: Rect, size: Size, client: Point): Point => {
    const sx = (size.width || rect.width) / rect.width;
    const sy = (size.height || rect.height) / rect.height;
    return {x: Math.round((client.x - rect.left) * sx), y: Math.round((client.y - rect.top) * sy)};
};

export interface PointerStart extends Point {
    t: number,
}

export type Gesture =
    | { kind: 'tap', at: Point }
    | { kind: 'swipe', from: Point, to: Point, durationMs: number };

// classifyGesture says whether a press and release was a tap or a swipe;
// toDevice maps client positions to device pixels
export const classifyGesture = (start: PointerStart, end: PointerStart, toDevice: (p: Point) => Point): Gesture => {
    const from = toDevice(start);
    if (Math.hypot(end.x - start.x, end.y - start.y) < TAP_DISTANCE) {
        return {kind: 'tap', at: from};
    }
    const durationMs = Math.min(Math.max(end.t - start.t, MIN_SWIPE_MS), MAX_SWIPE_MS);
    return {kind: 'swipe', from, to: toDevice(end), durationMs};
};

const clamp = (v: number): number => Math.max(-16, Math.min(16, v));

// scrollAmount turns a wheel delta (pixels or lines) into the lines scrcpy
// scrolls; it scrolls up for positive values
export const scrollAmount = (delta: number, inLines: boolean): number => clamp(-delta * (inLines ? 1 : 1 / 100));
