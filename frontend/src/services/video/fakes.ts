// Fakes of the browser APIs the video modules use, shared by their tests.
import {vi} from 'vitest';
import type {VideoDeps} from './types';

export class FakeSocket {
    readyState = 0;
    binaryType = '';
    sent: string[] = [];
    onmessage: ((event: { data: unknown }) => void) | null = null;
    onclose: (() => void) | null = null;

    constructor(readonly url: string) {}

    open(): void { this.readyState = 1; }
    send(text: string): void { this.sent.push(text); }
    // like the browser, closing ends in the close event
    close(): void { this.serverClose(); }
    serverClose(): void {
        this.readyState = 3;
        this.onclose?.();
    }
    emit(data: unknown): void { this.onmessage?.({data}); }
}

export class FakeDecoder {
    state: 'unconfigured' | 'configured' | 'closed' = 'unconfigured';
    configured: VideoDecoderConfig[] = [];
    decoded: EncodedVideoChunkInit[] = [];
    configureError: Error | null = null;
    decodeError: Error | null = null;

    constructor(readonly init: VideoDecoderInit) {}

    configure(config: VideoDecoderConfig): void {
        if (this.configureError) throw this.configureError;
        this.configured.push(config);
        this.state = 'configured';
    }
    decode(chunk: EncodedVideoChunkInit): void {
        if (this.decodeError) throw this.decodeError;
        this.decoded.push(chunk);
    }
    close(): void { this.state = 'closed'; }
    emitFrame(frame: VideoFrame): void { this.init.output(frame); }
    emitError(error: unknown): void { this.init.error(error as DOMException); }
}

export interface FakeBrowser {
    deps: VideoDeps,
    sockets: FakeSocket[],
    decoders: FakeDecoder[],
    bitmaps: Array<{ resolve: (b: ImageBitmap) => void, reject: (e: unknown) => void, blob: Blob }>,
    clock: { now: number },
}

export const createFakeBrowser = (): FakeBrowser => {
    const browser: FakeBrowser = {
        sockets: [], decoders: [], bitmaps: [], clock: {now: 100000},
        deps: {
            createSocket: url => {
                const socket = new FakeSocket(url);
                browser.sockets.push(socket);
                return socket as unknown as WebSocket;
            },
            createDecoder: init => {
                const decoder = new FakeDecoder(init);
                browser.decoders.push(decoder);
                return decoder as unknown as VideoDecoder;
            },
            createChunk: init => init as unknown as EncodedVideoChunk,
            createImageBitmap: blob => new Promise((resolve, reject) => browser.bitmaps.push({resolve, reject, blob})),
            now: () => browser.clock.now,
        },
    };
    return browser;
};

export const fakeBitmap = (width = 4, height = 3): ImageBitmap =>
    ({width, height, close: vi.fn()}) as unknown as ImageBitmap;

export const fakeFrame = (width = 4, height = 3): VideoFrame =>
    ({displayWidth: width, displayHeight: height, close: vi.fn()}) as unknown as VideoFrame;

export const fakeCanvas = (width = 0, height = 0): { canvas: HTMLCanvasElement, drawImage: ReturnType<typeof vi.fn> } => {
    const drawImage = vi.fn();
    const canvas = {
        width, height,
        getContext: () => ({drawImage}),
        getBoundingClientRect: () => ({left: 10, top: 20, width: 200, height: 100}),
    } as unknown as HTMLCanvasElement;
    return {canvas, drawImage};
};

export interface PacketInit { config?: boolean, key?: boolean, pts?: number, data?: number[] }

// [flags u8][pts u64 BE][data]
export const packet = ({config = false, key = false, pts = 0, data = []}: PacketInit = {}): Uint8Array => {
    const out = new Uint8Array(9 + data.length);
    out[0] = (config ? 1 : 0) | (key ? 2 : 0);
    new DataView(out.buffer).setBigUint64(1, BigInt(pts));
    out.set(data, 9);
    return out;
};

export const SPS = [0, 0, 0, 1, 0x67, 0x42, 0xc0, 0x1f, 0xaa];
export const CODEC = 'avc1.42c01f';
