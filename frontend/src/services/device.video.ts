// Live screen video of a device over a websocket. The messages are described
// in the Go package hub/videostream: JSON text messages (meta, error,
// inputError) and binary packets [flags u8][pts u64 BE][data].
// Android streams H.264 (Annex B, decoded with WebCodecs); iOS streams MJPEG
// from WebDriverAgent (one JPEG per packet, decoded with createImageBitmap).
// Input goes back as {type: 'input'} with the lease token, and is injected
// through scrcpy's control channel on Android. View-only streams (iOS) take no
// input; the page uses the REST API for them.

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

// scrcpy's limit for one text message, in UTF-8 bytes
export const VIDEO_TEXT_MAX_BYTES = 300;

export interface VideoOptions {
    maxSize?: number,
    maxFps?: number,
    bitRate?: number,
}

const FLAG_CONFIG = 1;
const FLAG_KEY_FRAME = 2;
const HEADER_SIZE = 9;

// WebCodecs only exists in secure contexts (https or localhost); it decodes
// the H.264 video of Android devices.
export const videoDecodingAvailable = (): boolean =>
    typeof window !== 'undefined' && window.isSecureContext && typeof window.VideoDecoder === 'function';

// MJPEG (iOS) only needs createImageBitmap, which works over plain http too.
export const mjpegDecodingAvailable = (): boolean =>
    typeof window !== 'undefined' && typeof window.createImageBitmap === 'function';

export const videoUrl = (projectId: string, deviceId: number | string, opts: VideoOptions = {}): string => {
    const params = new URLSearchParams();
    if (opts.maxSize !== undefined) params.set('maxSize', String(opts.maxSize));
    if (opts.maxFps !== undefined) params.set('maxFps', String(opts.maxFps));
    if (opts.bitRate !== undefined) params.set('bitRate', String(opts.bitRate));
    const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const query = params.toString();
    return `${scheme}//${window.location.host}/api/${encodeURIComponent(projectId)}/device/${encodeURIComponent(String(deviceId))}/video${query ? `?${query}` : ''}`;
};

// avc1.PPCCLL from the first SPS NAL unit of an Annex B buffer.
export const h264CodecString = (data: Uint8Array): string | null => {
    for (let i = 0; i + 3 < data.length; i++) {
        if (data[i] === 0 && data[i + 1] === 0 && data[i + 2] === 1) {
            const nal = i + 3;
            if ((data[nal] & 0x1f) === 7 && nal + 3 < data.length) {
                const hex = (b: number) => b.toString(16).padStart(2, '0');
                return `avc1.${hex(data[nal + 1])}${hex(data[nal + 2])}${hex(data[nal + 3])}`;
            }
            i = nal - 1;
        }
    }
    return null;
};

const concat = (a: Uint8Array, b: Uint8Array): Uint8Array => {
    const out = new Uint8Array(a.length + b.length);
    out.set(a, 0);
    out.set(b, a.length);
    return out;
};

const RECONNECT_MIN_MS = 1000;
const RECONNECT_MAX_MS = 10000;

/**
 * Plays the device video into a canvas and reconnects when the stream drops.
 * Devices without video support report 'unsupported' and are not retried.
 */
export class DeviceVideoPlayer {
    private ws: WebSocket | null = null;
    private decoder: VideoDecoder | null = null;
    private codec: string | null = null;
    private pendingConfig: Uint8Array | null = null;
    private needKeyFrame = true;
    private closed = false;
    private retryMs = RECONNECT_MIN_MS;
    private retryTimer: ReturnType<typeof setTimeout> | null = null;
    private ctx: CanvasRenderingContext2D | null;
    private status: VideoStatus | null = null;
    private lastFailure = 0;
    private meta: VideoMeta | null = null;
    // MJPEG: a frame is being decoded; the newest one that arrived meanwhile
    private decodingImage = false;
    private nextImage: Uint8Array | null = null;

    constructor(
        private readonly url: string,
        private readonly canvas: HTMLCanvasElement,
        private readonly onStatus: (status: VideoStatus) => void,
        private readonly onMeta: (meta: VideoMeta) => void,
        private readonly onInputError: (message: string) => void = () => undefined,
    ) {
        this.ctx = canvas.getContext('2d');
        this.connect();
    }

    // report only changes; draw() runs for every frame
    private setStatus(status: VideoStatus) {
        if (this.status?.state === status.state && this.status?.message === status.message) return;
        this.status = status;
        this.onStatus(status);
    }

    // whether input can be sent now: the socket is open, frames arrive and
    // the stream takes input
    get canSendInput(): boolean {
        return this.ws?.readyState === WebSocket.OPEN && this.status?.state === 'playing' && !this.meta?.viewOnly;
    }

    // sends input; false when the socket isn't ready (use the REST API then)
    sendInput(input: VideoInput, lease: string | null): boolean {
        if (this.ws?.readyState !== WebSocket.OPEN) return false;
        this.ws.send(JSON.stringify({type: 'input', input, lease: lease ?? undefined}));
        return true;
    }

    // maps a client position to pixels of the current video frame
    toVideo(clientX: number, clientY: number): { x: number, y: number, width: number, height: number } {
        const rect = this.canvas.getBoundingClientRect();
        const width = this.canvas.width;
        const height = this.canvas.height;
        const clamp = (v: number, max: number) => Math.min(Math.max(Math.round(v), 0), max - 1);
        return {
            x: clamp((clientX - rect.left) * width / rect.width, width),
            y: clamp((clientY - rect.top) * height / rect.height, height),
            width,
            height,
        };
    }

    close() {
        this.closed = true;
        if (this.retryTimer) clearTimeout(this.retryTimer);
        this.ws?.close();
        this.ws = null;
        this.resetDecoder();
    }

    private connect() {
        this.meta = null;
        this.setStatus({state: 'connecting'});
        const ws = new WebSocket(this.url);
        ws.binaryType = 'arraybuffer';
        this.ws = ws;
        let unsupported = false;

        ws.onmessage = (event) => {
            if (typeof event.data === 'string') {
                const message = JSON.parse(event.data);
                if (message.type === 'meta') {
                    this.meta = message.meta as VideoMeta;
                    this.onMeta(this.meta);
                } else if (message.type === 'inputError') {
                    this.onInputError(message.message);
                } else if (message.type === 'error') {
                    unsupported = !!message.unsupported;
                    this.setStatus({state: unsupported ? 'unsupported' : 'error', message: message.message});
                }
                return;
            }
            this.onPacket(new Uint8Array(event.data as ArrayBuffer));
        };
        ws.onclose = () => {
            if (this.ws !== ws) return;
            this.ws = null;
            this.resetDecoder();
            if (this.closed || unsupported) return;
            this.retryTimer = setTimeout(() => this.connect(), this.retryMs);
            this.retryMs = Math.min(this.retryMs * 2, RECONNECT_MAX_MS);
        };
    }

    private resetDecoder() {
        if (this.decoder && this.decoder.state !== 'closed') {
            this.decoder.close();
        }
        this.decoder = null;
        this.codec = null;
        this.pendingConfig = null;
        this.needKeyFrame = true;
        this.nextImage = null;
    }

    private onPacket(buf: Uint8Array) {
        if (buf.length < HEADER_SIZE) return;
        const flags = buf[0];
        const pts = Number(new DataView(buf.buffer, buf.byteOffset + 1, 8).getBigUint64(0));
        const data = buf.subarray(HEADER_SIZE);

        if (this.meta?.codec === 'mjpeg') {
            this.onImage(data);
            return;
        }
        if (flags & FLAG_CONFIG) {
            // SPS/PPS: (re)configure, e.g. after a rotation, and send it in
            // front of the next key frame
            const codec = h264CodecString(data);
            if (!codec) return;
            this.configure(codec);
            this.pendingConfig = data;
            this.needKeyFrame = true;
            return;
        }

        const key = (flags & FLAG_KEY_FRAME) !== 0;
        if (!this.decoder || this.decoder.state !== 'configured') return;
        if (this.needKeyFrame && !key) return;
        this.needKeyFrame = false;

        let chunkData = data;
        if (this.pendingConfig) {
            chunkData = concat(this.pendingConfig, data);
            this.pendingConfig = null;
        }
        try {
            this.decoder.decode(new EncodedVideoChunk({type: key ? 'key' : 'delta', timestamp: pts, data: chunkData}));
        } catch (e) {
            this.fail(e);
        }
    }

    private configure(codec: string) {
        if (!this.decoder || this.decoder.state === 'closed') {
            this.decoder = new VideoDecoder({
                output: frame => this.draw(frame),
                error: e => this.fail(e),
            });
        }
        try {
            this.decoder.configure({codec, optimizeForLatency: true});
            this.codec = codec;
        } catch (e) {
            this.fail(e);
        }
    }

    // decodes one JPEG at a time and skips frames that arrive meanwhile, so a
    // slow browser shows the newest screen instead of falling behind
    private onImage(data: Uint8Array) {
        if (this.decodingImage) {
            this.nextImage = data;
            return;
        }
        this.decodingImage = true;
        const ws = this.ws;
        createImageBitmap(new Blob([data as BlobPart], {type: 'image/jpeg'})).then(bitmap => {
            try {
                if (this.closed || this.ws !== ws) return;
                if (this.canvas.width !== bitmap.width || this.canvas.height !== bitmap.height) {
                    this.canvas.width = bitmap.width;
                    this.canvas.height = bitmap.height;
                }
                this.ctx?.drawImage(bitmap, 0, 0);
                this.played();
            } finally {
                bitmap.close();
            }
        }, error => {
            // a broken frame: the next one replaces it
            console.warn('MJPEG frame', error);
        }).finally(() => {
            this.decodingImage = false;
            const next = this.nextImage;
            this.nextImage = null;
            if (next && !this.closed && this.ws === ws) this.onImage(next);
        });
    }

    private played() {
        if (this.retryMs !== RECONNECT_MIN_MS) this.retryMs = RECONNECT_MIN_MS;
        this.setStatus({state: 'playing'});
    }

    private draw(frame: VideoFrame) {
        try {
            if (this.canvas.width !== frame.displayWidth || this.canvas.height !== frame.displayHeight) {
                this.canvas.width = frame.displayWidth;
                this.canvas.height = frame.displayHeight;
            }
            this.ctx?.drawImage(frame, 0, 0);
        } finally {
            frame.close();
        }
        this.played();
    }

    private fail(error: unknown) {
        this.setStatus({state: 'error', message: `video decoder: ${error instanceof Error ? error.message : String(error)}${this.codec ? ` (${this.codec})` : ''}`});
        const now = Date.now();
        if (now - this.lastFailure > 5000) {
            // first failure in a while: ask the device for a new config and
            // key frame, which recreates the decoder
            this.lastFailure = now;
            this.resetDecoder();
            this.sendInput({kind: 'reset_video'}, null);
            return;
        }
        // repeated failures: the socket's close handler resets and reconnects
        this.ws?.close();
    }
}
