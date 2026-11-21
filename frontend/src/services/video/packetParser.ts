import type {VideoMeta} from './types';

// Binary packets are [flags u8][pts u64 BE][data].
const FLAG_CONFIG = 1;
const FLAG_KEY_FRAME = 2;
export const HEADER_SIZE = 9;

export interface VideoPacket {
    config: boolean,
    key: boolean,
    pts: number,
    data: Uint8Array,
}

// null for a packet shorter than its header
export const parsePacket = (buf: Uint8Array): VideoPacket | null => {
    if (buf.length < HEADER_SIZE) return null;
    const flags = buf[0];
    return {
        config: (flags & FLAG_CONFIG) !== 0,
        key: (flags & FLAG_KEY_FRAME) !== 0,
        pts: Number(new DataView(buf.buffer, buf.byteOffset + 1, 8).getBigUint64(0)),
        data: buf.subarray(HEADER_SIZE),
    };
};

export type VideoMessage =
    | { type: 'meta', meta: VideoMeta }
    | { type: 'inputError', message: string }
    | { type: 'error', message: string, unsupported: boolean };

const isRecord = (v: unknown): v is Record<string, unknown> => typeof v === 'object' && v !== null;

// null for invalid JSON and for unknown message types
export const parseMessage = (text: string): VideoMessage | null => {
    let raw: unknown;
    try {
        raw = JSON.parse(text);
    } catch {
        return null;
    }
    if (!isRecord(raw)) return null;
    if (raw.type === 'meta') return {type: 'meta', meta: raw.meta as VideoMeta};
    if (raw.type === 'inputError') return {type: 'inputError', message: String(raw.message)};
    if (raw.type === 'error') return {type: 'error', message: String(raw.message), unsupported: !!raw.unsupported};
    return null;
};
