import {VideoInput, VideoMeta, VideoStatus} from '../services/device.video';

// video: H.264 or MJPEG stream, screenshots: stream of screenshots, off: manual refresh
export type ScreenMode = 'video' | 'screenshots' | 'off';

// The part of DeviceVideoPlayer the pages use; tests fake it.
export interface VideoPlayerLike {
    readonly canSendInput: boolean,

    sendInput(input: VideoInput, lease: string | null): boolean,

    toVideo(clientX: number, clientY: number): { x: number, y: number, width: number, height: number },

    close(): void,
}

export interface PlayerCallbacks {
    onStatus: (status: VideoStatus) => void,
    onMeta: (meta: VideoMeta) => void,
    onInputError: (message: string) => void,
}

export type VideoPlayerFactory = (url: string, canvas: HTMLCanvasElement, callbacks: PlayerCallbacks) => VideoPlayerLike;

// what the browser can decode, and how to start a player
export interface VideoSupport {
    decodesVideo: boolean,
    decodesMjpeg: boolean,
    createPlayer: VideoPlayerFactory,
}
