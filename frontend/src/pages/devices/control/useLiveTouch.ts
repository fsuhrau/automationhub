import {PointerEvent, useRef} from 'react';
import {VideoPlayerLike} from '../../../types/device.control.video';
import {useLatest} from './useLatest';

interface Touch {
    pointerId: number,
    player: VideoPlayerLike,
}

type TouchEnd = 'up' | 'cancel';

export interface LiveTouch {
    // starts a touch over the video socket; false when there is no live player
    begin: (e: PointerEvent<HTMLElement>, player: VideoPlayerLike | null) => boolean,
    move: (e: PointerEvent<HTMLElement>) => void,
    // ends the touch; false when the event isn't part of one
    end: (e: PointerEvent<HTMLElement>, action: TouchEnd) => boolean,
}

// useLiveTouch sends a touch over the video socket as it happens: down,
// moves (at most one per animation frame) and up.
export const useLiveTouch = (lease: string | null): LiveTouch => {
    const leaseRef = useLatest(lease);
    const active = useRef<Touch | null>(null);
    const pending = useRef<{ x: number, y: number } | null>(null);
    const frame = useRef(0);

    const send = (touch: Touch, action: 'down' | 'move' | 'up' | 'cancel', at: { x: number, y: number }): void => {
        touch.player.sendInput({kind: 'touch', action, pointerId: touch.pointerId, ...touch.player.toVideo(at.x, at.y)}, leaseRef.current);
    };
    const begin: LiveTouch['begin'] = (e, player) => {
        if (!player || active.current) return false;
        e.currentTarget.setPointerCapture(e.pointerId);
        active.current = {pointerId: e.pointerId, player};
        send(active.current, 'down', {x: e.clientX, y: e.clientY});
        return true;
    };
    const move: LiveTouch['move'] = e => {
        if (active.current?.pointerId !== e.pointerId) return;
        pending.current = {x: e.clientX, y: e.clientY};
        if (frame.current) return;
        frame.current = requestAnimationFrame(() => {
            frame.current = 0;
            if (pending.current && active.current) send(active.current, 'move', pending.current);
        });
    };
    const end: LiveTouch['end'] = (e, action) => {
        const touch = active.current;
        if (!touch || touch.pointerId !== e.pointerId) return false;
        active.current = null;
        cancelAnimationFrame(frame.current);
        frame.current = 0;
        send(touch, action, {x: e.clientX, y: e.clientY});
        return true;
    };
    return {begin, move, end};
};
