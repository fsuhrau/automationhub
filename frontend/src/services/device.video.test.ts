import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { DeviceVideoPlayer, h264CodecString, VideoMeta, VideoStatus } from './device.video';
import { CODEC, createFakeBrowser, fakeBitmap, fakeCanvas, fakeFrame, FakeBrowser, FakeSocket, packet, SPS } from './video/fakes';

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

const h264Meta: VideoMeta = {codec: 'h264', width: 4, height: 3, screenWidth: 8, screenHeight: 6};
let browser: FakeBrowser;
let canvas: ReturnType<typeof fakeCanvas>;
let statuses: VideoStatus[];
let metas: VideoMeta[];
let inputErrors: string[];
let player: DeviceVideoPlayer;

beforeEach(() => {
    vi.useFakeTimers();
    browser = createFakeBrowser();
    canvas = fakeCanvas();
    statuses = [];
    metas = [];
    inputErrors = [];
    player = new DeviceVideoPlayer('ws://hub/video', canvas.canvas, s => statuses.push(s), m => metas.push(m), m => inputErrors.push(m), browser.deps);
});
afterEach(() => vi.useRealTimers());

const socket = (i = browser.sockets.length - 1): FakeSocket => browser.sockets[i];
const text = (message: unknown): void => socket().emit(JSON.stringify(message));
const binary = (bytes: Uint8Array): void => socket().emit(bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.length));
const states = (): string[] => statuses.map(s => s.state);
const startH264 = (): void => {
    socket().open();
    text({type: 'meta', meta: h264Meta});
    binary(packet({config: true, data: SPS}));
    binary(packet({key: true, pts: 1, data: [1]}));
};

it('plays H.264 packets into the canvas, then cleans up on close', () => {
    expect(states()).toEqual(['connecting']);
    startH264();
    expect(metas).toEqual([h264Meta]);
    const decoder = browser.decoders[0];
    expect(decoder.configured[0].codec).toBe(CODEC);
    expect(decoder.decoded).toHaveLength(1);
    const frame = fakeFrame(4, 3);
    decoder.emitFrame(frame);
    expect(canvas.drawImage).toHaveBeenCalledWith(frame, 0, 0);
    expect([canvas.canvas.width, canvas.canvas.height]).toEqual([4, 3]);
    expect(frame.close).toHaveBeenCalled();
    expect(states()).toEqual(['connecting', 'playing']);
    expect(player.canSendInput).toBe(true);

    player.close();
    expect(decoder.state).toBe('closed');
    expect(socket().readyState).toBe(3);
    vi.advanceTimersByTime(60000);
    expect(browser.sockets).toHaveLength(1);
    expect(player.sendInput({kind: 'reset_video'}, null)).toBe(false);
});

it('plays MJPEG packets with createImageBitmap', async () => {
    vi.useRealTimers();
    socket().open();
    text({type: 'meta', meta: {...h264Meta, codec: 'mjpeg'}});
    binary(packet({data: [0xff, 0xd8]}));
    const bitmap = fakeBitmap(6, 5);
    browser.bitmaps[0].resolve(bitmap);
    await new Promise(resolve => setTimeout(resolve, 0));
    expect(canvas.drawImage).toHaveBeenCalledWith(bitmap, 0, 0);
    expect(states()).toEqual(['connecting', 'playing']);
    expect(browser.decoders).toHaveLength(0);
    player.close();
});

it('sends input as JSON with the lease while the socket is open', () => {
    socket().open();
    const input = {kind: 'key', key: 'HOME'} as const;
    expect(player.sendInput(input, 'lease-1')).toBe(true);
    expect(JSON.parse(socket().sent[0])).toEqual({type: 'input', input, lease: 'lease-1'});
    player.sendInput(input, null);
    expect(JSON.parse(socket().sent[1]).lease).toBeUndefined();
});

it('takes no input before frames arrive or on a view-only stream', () => {
    socket().open();
    expect(player.canSendInput).toBe(false);
    text({type: 'meta', meta: {...h264Meta, viewOnly: true}});
    binary(packet({config: true, data: SPS}));
    binary(packet({key: true, data: [1]}));
    browser.decoders[0].emitFrame(fakeFrame());
    expect(player.canSendInput).toBe(false);
});

it('maps client positions to frame pixels', () => {
    canvas.canvas.width = 100;
    canvas.canvas.height = 50;
    expect(player.toVideo(110, 70)).toEqual({x: 50, y: 25, width: 100, height: 50});
});

it('reports input errors and ignores short packets and unknown messages', () => {
    socket().open();
    text({type: 'inputError', message: 'no lease'});
    text({type: 'whatever'});
    binary(new Uint8Array(3));
    expect(inputErrors).toEqual(['no lease']);
    expect(states()).toEqual(['connecting']);
});

it('reports a stream error and reconnects', () => {
    socket().open();
    text({type: 'error', message: 'device busy'});
    expect(statuses.at(-1)).toEqual({state: 'error', message: 'device busy'});
    socket().serverClose();
    vi.advanceTimersByTime(1000);
    expect(browser.sockets).toHaveLength(2);
    expect(states().at(-1)).toBe('connecting');
    expect(player.canSendInput).toBe(false);
});

it('does not retry an unsupported device', () => {
    text({type: 'error', message: 'no video', unsupported: true});
    socket().serverClose();
    vi.advanceTimersByTime(60000);
    expect(statuses.at(-1)).toEqual({state: 'unsupported', message: 'no video'});
    expect(browser.sockets).toHaveLength(1);
});

it('resets the decoder and asks the device for a new config on the first decoder failure', () => {
    startH264();
    browser.decoders[0].emitError(new Error('corrupt'));
    expect(statuses.at(-1)).toEqual({state: 'error', message: `video decoder: corrupt (${CODEC})`});
    expect(browser.decoders[0].state).toBe('closed');
    expect(JSON.parse(socket().sent[0]).input).toEqual({kind: 'reset_video'});
    expect(socket().readyState).toBe(1);
});

it('reconnects when the decoder fails again within five seconds', () => {
    startH264();
    browser.decoders[0].emitError('first');
    browser.clock.now += 1000;
    browser.decoders[0].emitError(new Error('second'));
    expect(socket(0).readyState).toBe(3);
    vi.advanceTimersByTime(1000);
    expect(browser.sockets).toHaveLength(2);
});

it('asks for a new config again once the failures are five seconds apart', () => {
    startH264();
    browser.decoders[0].emitError(new Error('a'));
    browser.clock.now += 6000;
    browser.decoders[0].emitError(new Error('b'));
    expect(socket().sent).toHaveLength(2);
    expect(socket().readyState).toBe(1);
});

it('ignores input errors when no handler is given', () => {
    const bare = new DeviceVideoPlayer('ws://hub/video', canvas.canvas, () => undefined, () => undefined, undefined, browser.deps);
    browser.sockets[1].emit(JSON.stringify({type: 'inputError', message: 'x'}));
    expect(bare.canSendInput).toBe(false);
});
