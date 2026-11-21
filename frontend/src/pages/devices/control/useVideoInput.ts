import {MutableRefObject, RefObject, useEffect} from 'react';
import {VideoMeta, VideoStatus, videoUrl} from '../../../services/device.video';
import {DeviceTarget} from '../../../types/device.control.api';
import {ScreenMode, VideoPlayerLike, VideoSupport} from '../../../types/device.control.video';
import {unsupportedNotice} from '../deviceControlModel';
import {scrollAmount} from './screenModel';
import {useLatest} from './useLatest';
import {useVideoPlayer} from './useVideoPlayer';

const VIDEO_OPTIONS = {maxSize: 1280, maxFps: 30};

export interface VideoInputOptions {
    mode: ScreenMode,
    target: DeviceTarget,
    support: VideoSupport,
    lease: string | null,
    // live video isn't available for this device: show screenshots, and tell
    // the user why when there is something to say
    onUnsupported: (notice: string | null) => void,
    onInputError: (message: string) => void,
}

export interface VideoInput {
    canvasRef: RefObject<HTMLCanvasElement | null>,
    status: VideoStatus | null,
    meta: VideoMeta | null,
    // the player when input can go over the video socket, else null
    // (screenshots, iOS, not connected yet, no lease): then the REST API is used
    liveInput: () => VideoPlayerLike | null,
}

interface WheelOptions {
    canvasRef: RefObject<HTMLCanvasElement | null>,
    playerRef: MutableRefObject<VideoPlayerLike | null>,
    leaseRef: MutableRefObject<string | null>,
    enabled: boolean,
}

// the mouse wheel scrolls on the device; a native listener, since React's
// wheel listener is passive and can't keep the page from scrolling
const useVideoWheel = ({canvasRef, playerRef, leaseRef, enabled}: WheelOptions): void => {
    useEffect(() => {
        const canvas = canvasRef.current;
        if (!enabled || !canvas) return;
        const onWheel = (e: WheelEvent): void => {
            const player = playerRef.current;
            const token = leaseRef.current;
            if (!player?.canSendInput || !token) return;
            e.preventDefault();
            const lines = e.deltaMode === WheelEvent.DOM_DELTA_LINE;
            player.sendInput({
                kind: 'scroll', ...player.toVideo(e.clientX, e.clientY),
                scrollX: scrollAmount(e.deltaX, lines), scrollY: scrollAmount(e.deltaY, lines),
            }, token);
        };
        canvas.addEventListener('wheel', onWheel, {passive: false});
        return () => canvas.removeEventListener('wheel', onWheel);
    }, [canvasRef, playerRef, leaseRef, enabled]);
};

// useVideoInput plays the live video of the device in the video mode and
// takes input over its socket: the wheel here, touches and keys through
// liveInput().
export const useVideoInput = (options: VideoInputOptions): VideoInput => {
    const {mode, target, support, lease, onUnsupported, onInputError} = options;
    const leaseRef = useLatest(lease);
    const enabled = mode === 'video';
    const {canvasRef, playerRef, status, meta} = useVideoPlayer({
        enabled, createPlayer: support.createPlayer, onInputError,
        url: videoUrl(target.projectId, target.deviceId, VIDEO_OPTIONS),
        onStatus: s => s.state === 'unsupported' && onUnsupported(unsupportedNotice(s.message)),
    });
    useVideoWheel({canvasRef, playerRef, leaseRef, enabled});
    const liveInput = (): VideoPlayerLike | null =>
        enabled && lease && playerRef.current?.canSendInput ? playerRef.current : null;
    return {canvasRef, status, meta, liveInput};
};
