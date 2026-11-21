import {MutableRefObject, RefObject, useEffect, useRef, useState} from 'react';
import {VideoMeta, VideoStatus} from '../../../services/device.video';
import {VideoPlayerFactory, VideoPlayerLike} from '../../../types/device.control.video';
import {useLatest} from './useLatest';

export interface VideoPlayerOptions {
    // play only while true (a canvas is rendered)
    enabled: boolean,
    url: string,
    createPlayer: VideoPlayerFactory,
    onStatus?: (status: VideoStatus) => void,
    onInputError?: (message: string) => void,
}

export interface VideoPlayerState {
    canvasRef: RefObject<HTMLCanvasElement | null>,
    playerRef: MutableRefObject<VideoPlayerLike | null>,
    status: VideoStatus | null,
    meta: VideoMeta | null,
}

// useVideoPlayer plays the video of url on the canvas while enabled, and
// reports its status and stream meta. The player is closed when the url or
// enabled changes, and on unmount.
export const useVideoPlayer = (options: VideoPlayerOptions): VideoPlayerState => {
    const {enabled, url, createPlayer} = options;
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const playerRef = useRef<VideoPlayerLike | null>(null);
    const [status, setStatus] = useState<VideoStatus | null>(null);
    const [meta, setMeta] = useState<VideoMeta | null>(null);
    const latest = useLatest(options);

    useEffect(() => {
        if (!enabled || !canvasRef.current) return;
        const player = createPlayer(url, canvasRef.current, {
            onStatus: s => {
                setStatus(s);
                latest.current.onStatus?.(s);
            },
            onMeta: setMeta,
            onInputError: message => latest.current.onInputError?.(message),
        });
        playerRef.current = player;
        return () => {
            player.close();
            playerRef.current = null;
            setMeta(null);
        };
    }, [enabled, url, createPlayer, latest]);
    return {canvasRef, playerRef, status, meta};
};
