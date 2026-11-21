// Live screen video of a device over a websocket. The messages are described
// in the Go package hub/videostream: JSON text messages (meta, error,
// inputError) and binary packets [flags u8][pts u64 BE][data].
// Android streams H.264 (Annex B, decoded with WebCodecs); iOS streams MJPEG
// from WebDriverAgent (one JPEG per packet, decoded with createImageBitmap).
// Input goes back as {type: 'input'} with the lease token, and is injected
// through scrcpy's control channel on Android. View-only streams (iOS) take no
// input; the page uses the REST API for them.
//
// This file composes the modules in ./video and keeps the public API.
import {defaultVideoDeps} from './video/deps';
import {createH264Decoder, H264Decoder} from './video/h264Decoder';
import {createMjpegDecoder, MjpegDecoder} from './video/mjpegDecoder';
import {parseMessage, parsePacket, VideoMessage} from './video/packetParser';
import {createRenderer, Renderer, toVideoPoint, VideoPoint} from './video/renderer';
import {createTransport, Transport} from './video/transport';
import type {VideoDeps, VideoInput, VideoMeta, VideoStatus} from './video/types';

export type {VideoDeps, VideoInput, VideoMeta, VideoOptions, VideoState, VideoStatus} from './video/types';
export {h264CodecString} from './video/h264Codec';
export {mjpegDecodingAvailable, videoDecodingAvailable, videoUrl} from './video/support';

// scrcpy's limit for one text message, in UTF-8 bytes
export const VIDEO_TEXT_MAX_BYTES = 300;

// a second decoder failure within this time reconnects instead of resetting
const FAILURE_WINDOW_MS = 5000;

/**
 * Plays the device video into a canvas and reconnects when the stream drops.
 * Devices without video support report 'unsupported' and are not retried.
 */
export class DeviceVideoPlayer {
    private readonly transport: Transport;
    private readonly renderer: Renderer;
    private readonly h264: H264Decoder;
    private readonly mjpeg: MjpegDecoder;
    private status: VideoStatus | null = null;
    private meta: VideoMeta | null = null;
    private lastFailure = 0;

    // The first five parameters are the public API used by the pages; deps are for tests.
    // eslint-disable-next-line max-params
    constructor(
        url: string,
        private readonly canvas: HTMLCanvasElement,
        private readonly onStatus: (status: VideoStatus) => void,
        private readonly onMeta: (meta: VideoMeta) => void,
        private readonly onInputError: (message: string) => void = () => undefined,
        private readonly deps: VideoDeps = defaultVideoDeps(),
    ) {
        this.renderer = createRenderer(canvas);
        this.h264 = createH264Decoder({deps, onFrame: frame => this.drawFrame(frame), onError: e => this.fail(e)});
        this.mjpeg = createMjpegDecoder({deps, onFrame: bitmap => this.drawBitmap(bitmap)});
        this.transport = createTransport({
            url,
            createSocket: deps.createSocket,
            handlers: {
                onConnecting: () => this.onConnecting(),
                onText: text => this.onText(text),
                onBinary: data => this.onBinary(data),
                onDisconnect: () => this.resetDecoders(),
            },
        });
    }

    // whether input can be sent now: the socket is open, frames arrive and
    // the stream takes input
    get canSendInput(): boolean {
        return this.transport.isOpen && this.status?.state === 'playing' && !this.meta?.viewOnly;
    }

    // sends input; false when the socket isn't ready (use the REST API then)
    sendInput(input: VideoInput, lease: string | null): boolean {
        return this.transport.send(JSON.stringify({type: 'input', input, lease: lease ?? undefined}));
    }

    // maps a client position to pixels of the current video frame
    toVideo(clientX: number, clientY: number): VideoPoint {
        return toVideoPoint(this.canvas, clientX, clientY);
    }

    close(): void {
        this.transport.stop();
        this.resetDecoders();
    }

    // report only changes; frames report 'playing' all the time
    private setStatus(status: VideoStatus): void {
        if (this.status?.state === status.state && this.status?.message === status.message) return;
        this.status = status;
        this.onStatus(status);
    }

    private onConnecting(): void {
        this.meta = null;
        this.setStatus({state: 'connecting'});
    }

    private onText(text: string): void {
        const message = parseMessage(text);
        if (message) this.onMessage(message);
    }

    private onMessage(message: VideoMessage): void {
        if (message.type === 'meta') {
            this.meta = message.meta;
            this.onMeta(message.meta);
        } else if (message.type === 'inputError') {
            this.onInputError(message.message);
        } else {
            if (message.unsupported) this.transport.giveUp();
            this.setStatus({state: message.unsupported ? 'unsupported' : 'error', message: message.message});
        }
    }

    private onBinary(buf: Uint8Array): void {
        const packet = parsePacket(buf);
        if (!packet) return;
        if (this.meta?.codec === 'mjpeg') this.mjpeg.push(packet.data);
        else if (packet.config) this.h264.configure(packet.data);
        else this.h264.decode(packet);
    }

    private resetDecoders(): void {
        this.h264.reset();
        this.mjpeg.reset();
    }

    private played(): void {
        this.transport.resetBackoff();
        this.setStatus({state: 'playing'});
    }

    private drawBitmap(bitmap: ImageBitmap): void {
        this.renderer.draw(bitmap, bitmap.width, bitmap.height);
        this.played();
    }

    private drawFrame(frame: VideoFrame): void {
        try {
            this.renderer.draw(frame, frame.displayWidth, frame.displayHeight);
        } finally {
            frame.close();
        }
        this.played();
    }

    private fail(error: unknown): void {
        const detail = error instanceof Error ? error.message : String(error);
        const codec = this.h264.codec;
        this.setStatus({state: 'error', message: `video decoder: ${detail}${codec ? ` (${codec})` : ''}`});
        const now = this.deps.now();
        if (now - this.lastFailure > FAILURE_WINDOW_MS) {
            // first failure in a while: ask the device for a new config and
            // key frame, which recreates the decoder
            this.lastFailure = now;
            this.h264.reset();
            this.sendInput({kind: 'reset_video'}, null);
            return;
        }
        // repeated failures: the socket's close handler resets and reconnects
        this.transport.drop();
    }
}
