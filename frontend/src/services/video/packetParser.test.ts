import {describe, expect, it} from 'vitest';
import {packet} from './fakes';
import {parseMessage, parsePacket} from './packetParser';

describe('parsePacket', () => {
    it.each([
        ['a config packet', {config: true, pts: 0}, {config: true, key: false}],
        ['a key frame', {key: true, pts: 7}, {config: false, key: true}],
        ['a delta frame', {pts: 4294967297}, {config: false, key: false}],
    ])('reads %s', (_name, init, flags) => {
        const parsed = parsePacket(packet({...init, data: [9, 8]}));
        expect(parsed).toMatchObject({...flags, pts: init.pts});
        expect(Array.from(parsed?.data ?? [])).toEqual([9, 8]);
    });

    it('reads the pts of a packet that is a view into a larger buffer', () => {
        const inner = packet({pts: 5, data: [1]});
        const outer = new Uint8Array(inner.length + 3);
        outer.set(inner, 3);
        expect(parsePacket(outer.subarray(3))?.pts).toBe(5);
    });

    it('returns null for a packet shorter than the header', () => {
        expect(parsePacket(new Uint8Array(8))).toBeNull();
    });
});

describe('parseMessage', () => {
    const meta = {codec: 'h264', width: 1, height: 2, screenWidth: 3, screenHeight: 4};

    it.each([
        ['meta', {type: 'meta', meta}, {type: 'meta', meta}],
        ['inputError', {type: 'inputError', message: 'busy'}, {type: 'inputError', message: 'busy'}],
        ['error', {type: 'error', message: 'x', unsupported: true}, {type: 'error', message: 'x', unsupported: true}],
        ['error without unsupported', {type: 'error', message: 'x'}, {type: 'error', message: 'x', unsupported: false}],
    ])('parses %s', (_name, input, expected) => {
        expect(parseMessage(JSON.stringify(input))).toEqual(expected);
    });

    it.each([
        ['an unknown type', JSON.stringify({type: 'nope'})],
        ['invalid JSON', '{'],
        ['a JSON scalar', '42'],
        ['null', 'null'],
    ])('returns null for %s', (_name, text) => {
        expect(parseMessage(text)).toBeNull();
    });
});
