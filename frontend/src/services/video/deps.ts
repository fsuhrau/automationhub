import type {VideoDeps} from './types';

// Resolved at call time, so a test (or an old browser) can replace the globals.
export const defaultVideoDeps = (): VideoDeps => ({
    createSocket: url => new WebSocket(url),
    createDecoder: init => new VideoDecoder(init),
    createChunk: init => new EncodedVideoChunk(init),
    createImageBitmap: blob => createImageBitmap(blob),
    now: () => Date.now(),
});
