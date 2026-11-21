import { describe, expect, it } from 'vitest';
import { h264CodecString } from './device.video';

const startCode = [0, 0, 1];

describe('h264CodecString', () => {
    it('builds the codec string from the first SPS NAL unit', () => {
        const data = new Uint8Array([0, 0, 0, 1, 0x67, 0x42, 0xc0, 0x1f, 0xaa]);
        expect(h264CodecString(data)).toBe('avc1.42c01f');
    });

    it('skips NAL units that are not an SPS', () => {
        const data = new Uint8Array([...startCode, 0x68, 0x01, ...startCode, 0x67, 0x64, 0x00, 0x28, 0x00]);
        expect(h264CodecString(data)).toBe('avc1.640028');
    });

    it('pads single digit hex values', () => {
        const data = new Uint8Array([...startCode, 0x67, 0x04, 0x00, 0x0a, 0x00]);
        expect(h264CodecString(data)).toBe('avc1.04000a');
    });

    it.each([
        ['an empty buffer', []],
        ['a buffer without a start code', [0x67, 0x42, 0xc0, 0x1f]],
        ['an SPS cut off before its profile bytes', [...startCode, 0x67, 0x42, 0xc0]],
    ])('returns null for %s', (_name, bytes) => {
        expect(h264CodecString(new Uint8Array(bytes))).toBeNull();
    });
});
