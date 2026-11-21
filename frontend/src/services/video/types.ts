// Shared types of the live device video. The messages are described in the Go
// package hub/videostream.

export interface VideoMeta {
    codec: string,
    width: number,
    height: number,
    // device resolution in the natural orientation (0 when unknown)
    screenWidth: number,
    screenHeight: number,
    // the stream takes no input over the socket
    viewOnly?: boolean,
}

export type VideoState = 'connecting' | 'playing' | 'unsupported' | 'error';

export interface VideoStatus {
    state: VideoState,
    message?: string,
}

// Input over the video socket. Touch and scroll positions are pixels of the
// current video frame (width x height), see DeviceVideoPlayer.toVideo.
export type VideoInput =
    | { kind: 'touch', action: 'down' | 'move' | 'up' | 'cancel', pointerId: number, x: number, y: number, width: number, height: number }
    | { kind: 'key', action?: 'press' | 'down' | 'up', key: string }
    | { kind: 'text', text: string }
    | { kind: 'scroll', x: number, y: number, width: number, height: number, scrollX: number, scrollY: number }
    | { kind: 'reset_video' };

export interface VideoOptions {
    maxSize?: number,
    maxFps?: number,
    bitRate?: number,
}

// What the player needs from the browser; the defaults are the globals.
export interface VideoDeps {
    createSocket: (url: string) => WebSocket,
    createDecoder: (init: VideoDecoderInit) => VideoDecoder,
    createChunk: (init: EncodedVideoChunkInit) => EncodedVideoChunk,
    createImageBitmap: (blob: Blob) => Promise<ImageBitmap>,
    now: () => number,
}
