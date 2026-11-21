import type {VideoDeps} from './types';

export interface MjpegDecoderOptions {
    deps: Pick<VideoDeps, 'createImageBitmap'>,
    onFrame: (bitmap: ImageBitmap) => void,
}

export interface MjpegDecoder {
    push: (jpeg: Uint8Array) => void,
    // forgets queued and in-flight frames (new connection, stop)
    reset: () => void,
}

// Decodes one JPEG at a time and skips frames that arrive meanwhile, so a
// slow browser shows the newest screen instead of falling behind.
export const createMjpegDecoder = ({deps, onFrame}: MjpegDecoderOptions): MjpegDecoder => {
    let decoding = false;
    let next: Uint8Array | null = null;
    let epoch = 0;

    const deliver = (bitmap: ImageBitmap, startedIn: number): void => {
        try {
            if (startedIn === epoch) onFrame(bitmap);
        } finally {
            bitmap.close();
        }
    };

    const drain = (): void => {
        decoding = false;
        const queued = next;
        next = null;
        if (queued) push(queued);
    };

    function push(jpeg: Uint8Array): void {
        if (decoding) {
            next = jpeg;
            return;
        }
        decoding = true;
        const startedIn = epoch;
        deps.createImageBitmap(new Blob([jpeg as BlobPart], {type: 'image/jpeg'}))
            .then(bitmap => deliver(bitmap, startedIn))
            // a broken frame: the next one replaces it
            .catch(error => console.warn('MJPEG frame', error))
            .finally(drain);
    }

    return {push, reset: () => { epoch++; next = null; }};
};
