import {h264CodecString} from './h264Codec';
import type {VideoDeps} from './types';

export interface H264Frame {
    key: boolean,
    pts: number,
    data: Uint8Array,
}

export interface H264DecoderOptions {
    deps: Pick<VideoDeps, 'createDecoder' | 'createChunk'>,
    onFrame: (frame: VideoFrame) => void,
    onError: (error: unknown) => void,
}

export interface H264Decoder {
    readonly codec: string | null,
    // SPS/PPS: (re)configures, e.g. after a rotation
    configure: (data: Uint8Array) => void,
    decode: (frame: H264Frame) => void,
    reset: () => void,
}

const concat = (a: Uint8Array, b: Uint8Array): Uint8Array => {
    const out = new Uint8Array(a.length + b.length);
    out.set(a, 0);
    out.set(b, a.length);
    return out;
};

class WebCodecsH264Decoder implements H264Decoder {
    codec: string | null = null;
    private decoder: VideoDecoder | null = null;
    private pendingConfig: Uint8Array | null = null;
    private needKeyFrame = true;

    constructor(private readonly options: H264DecoderOptions) {}

    configure(data: Uint8Array): void {
        const codec = h264CodecString(data);
        if (!codec) return;
        try {
            this.ensureDecoder().configure({codec, optimizeForLatency: true});
            this.codec = codec;
        } catch (e) {
            this.options.onError(e);
        }
        // the config goes in front of the next key frame
        this.pendingConfig = data;
        this.needKeyFrame = true;
    }

    decode({key, pts, data}: H264Frame): void {
        const decoder = this.decoder;
        if (!decoder || decoder.state !== 'configured') return;
        if (this.needKeyFrame && !key) return;
        this.needKeyFrame = false;
        const chunk = this.pendingConfig ? concat(this.pendingConfig, data) : data;
        this.pendingConfig = null;
        try {
            decoder.decode(this.options.deps.createChunk({type: key ? 'key' : 'delta', timestamp: pts, data: chunk}));
        } catch (e) {
            this.options.onError(e);
        }
    }

    reset(): void {
        if (this.decoder && this.decoder.state !== 'closed') this.decoder.close();
        this.decoder = null;
        this.codec = null;
        this.pendingConfig = null;
        this.needKeyFrame = true;
    }

    private ensureDecoder(): VideoDecoder {
        if (!this.decoder || this.decoder.state === 'closed') {
            const {deps, onFrame, onError} = this.options;
            this.decoder = deps.createDecoder({output: onFrame, error: onError});
        }
        return this.decoder;
    }
}

export const createH264Decoder = (options: H264DecoderOptions): H264Decoder => new WebCodecsH264Decoder(options);
