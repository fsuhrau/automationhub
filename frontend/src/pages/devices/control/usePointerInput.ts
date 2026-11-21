import {PointerEvent, useRef} from 'react';
import {DeviceControlApi, Point} from '../../../types/device.control.api';
import {VideoMeta} from '../../../services/device.video';
import {VideoPlayerLike} from '../../../types/device.control.video';
import {classifyGesture, mapToDevice, PointerStart, Size, videoDeviceSize} from './screenModel';
import {Runner} from './useGuardedAction';
import {useLiveTouch} from './useLiveTouch';

export interface PointerDeps {
    lease: string | null,
    meta: VideoMeta | null,
    service: DeviceControlApi,
    run: Runner,
    liveInput: () => VideoPlayerLike | null,
    // set while picking an element on the screen: a press picks instead of
    // touching the device
    onPick: ((el: HTMLElement, at: Point) => void) | null,
    // after a tap or swipe went through the REST API
    onGesture: () => void,
}

export interface PointerHandlers {
    onPointerDown: (e: PointerEvent<HTMLElement>) => void,
    onPointerMove: (e: PointerEvent<HTMLElement>) => void,
    onPointerUp: (e: PointerEvent<HTMLElement>) => void,
    onPointerCancel: (e: PointerEvent<HTMLElement>) => void,
}

// device pixels of the shown screen: screenshots are in device pixels; the
// video may be scaled
const deviceSizeOf = (el: HTMLElement, meta: VideoMeta | null): Size =>
    el instanceof HTMLImageElement
        ? {width: el.naturalWidth, height: el.naturalHeight}
        : videoDeviceSize(el as HTMLCanvasElement, meta);

const start = (e: PointerEvent<HTMLElement>): PointerStart => ({x: e.clientX, y: e.clientY, t: Date.now()});

// usePointerInput turns pointer events on the screen into input: a live
// touch over the video socket when it takes input, else a tap or swipe
// through the REST API once the pointer is released.
export const usePointerInput = (deps: PointerDeps): PointerHandlers => {
    const {lease, meta, service, run} = deps;
    const touch = useLiveTouch(lease);
    const pressed = useRef<PointerStart | null>(null);

    const sendGesture = (el: HTMLElement, from: PointerStart, to: PointerStart): void => {
        if (!lease) return;
        const toDevice = (p: Point): Point => mapToDevice(el.getBoundingClientRect(), deviceSizeOf(el, meta), p);
        const gesture = classifyGesture(from, to, toDevice);
        run(() => gesture.kind === 'tap'
            ? service.tap(lease, gesture.at)
            : service.swipe(lease, gesture));
        deps.onGesture();
    };
    const onPointerDown = (e: PointerEvent<HTMLElement>): void => {
        e.preventDefault();
        (e.currentTarget.parentElement as HTMLElement | null)?.focus();
        if (deps.onPick) {
            deps.onPick(e.currentTarget, {x: e.clientX, y: e.clientY});
        } else if (!touch.begin(e, deps.liveInput())) {
            pressed.current = start(e);
        }
    };
    const onPointerUp = (e: PointerEvent<HTMLElement>): void => {
        if (touch.end(e, 'up')) return;
        const from = pressed.current;
        pressed.current = null;
        if (from) sendGesture(e.currentTarget, from, start(e));
    };
    const onPointerCancel = (e: PointerEvent<HTMLElement>): void => {
        if (!touch.end(e, 'cancel')) pressed.current = null;
    };
    return {onPointerDown, onPointerMove: touch.move, onPointerUp, onPointerCancel};
};
