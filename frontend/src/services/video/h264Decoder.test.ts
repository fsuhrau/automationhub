import {beforeEach, expect, it, Mock, vi} from 'vitest';
import {createFakeBrowser, CODEC, fakeFrame, FakeBrowser, SPS} from './fakes';
import {createH264Decoder, H264Decoder, H264Frame} from './h264Decoder';

let browser: FakeBrowser;
let onFrame: Mock<(frame: VideoFrame) => void>;
let onError: Mock<(error: unknown) => void>;
let decoder: H264Decoder;
const sps = new Uint8Array(SPS);

beforeEach(() => {
    browser = createFakeBrowser();
    onFrame = vi.fn();
    onError = vi.fn();
    decoder = createH264Decoder({deps: browser.deps, onFrame, onError});
});

const frame = (key: boolean, data = [5]): H264Frame => ({key, pts: 9, data: new Uint8Array(data)});

it('configures the decoder from the SPS codec string', () => {
    decoder.configure(sps);
    expect(browser.decoders[0].configured).toEqual([{codec: CODEC, optimizeForLatency: true}]);
    expect(decoder.codec).toBe(CODEC);
});

it('ignores a config without an SPS', () => {
    decoder.configure(new Uint8Array([1, 2, 3]));
    expect(browser.decoders).toHaveLength(0);
    expect(decoder.codec).toBeNull();
});

it('reuses the decoder when it is reconfigured', () => {
    decoder.configure(sps);
    decoder.configure(sps);
    expect(browser.decoders).toHaveLength(1);
    expect(browser.decoders[0].configured).toHaveLength(2);
});

it('creates a new decoder when the old one was closed', () => {
    decoder.configure(sps);
    browser.decoders[0].close();
    decoder.configure(sps);
    expect(browser.decoders).toHaveLength(2);
});

it('sends the config in front of the first key frame, skipping delta frames before it', () => {
    decoder.configure(sps);
    decoder.decode(frame(false));
    expect(browser.decoders[0].decoded).toHaveLength(0);
    decoder.decode(frame(true, [7]));
    const [chunk] = browser.decoders[0].decoded;
    expect(chunk.type).toBe('key');
    expect(chunk.timestamp).toBe(9);
    expect(Array.from(chunk.data as Uint8Array)).toEqual([...SPS, 7]);
});

it('sends delta frames unchanged after the key frame', () => {
    decoder.configure(sps);
    decoder.decode(frame(true));
    decoder.decode(frame(false, [6]));
    const chunk = browser.decoders[0].decoded[1];
    expect(chunk.type).toBe('delta');
    expect(Array.from(chunk.data as Uint8Array)).toEqual([6]);
});

it('waits for a key frame again after a new config', () => {
    decoder.configure(sps);
    decoder.decode(frame(true));
    decoder.configure(sps);
    decoder.decode(frame(false));
    expect(browser.decoders[0].decoded).toHaveLength(1);
});

it('drops frames before the decoder is configured', () => {
    decoder.decode(frame(true));
    decoder.configure(sps);
    browser.decoders[0].state = 'unconfigured';
    decoder.decode(frame(true));
    expect(browser.decoders[0].decoded).toHaveLength(0);
});

it('forwards decoded frames and decoder errors', () => {
    decoder.configure(sps);
    const f = fakeFrame();
    browser.decoders[0].emitFrame(f);
    browser.decoders[0].emitError(new Error('boom'));
    expect(onFrame).toHaveBeenCalledWith(f);
    expect(onError).toHaveBeenCalledWith(new Error('boom'));
});

it('reports a configure error', () => {
    decoder.configure(sps);
    browser.decoders[0].configureError = new Error('unsupported');
    decoder.configure(sps);
    expect(onError).toHaveBeenCalledWith(new Error('unsupported'));
});

it('reports a decode error', () => {
    decoder.configure(sps);
    browser.decoders[0].decodeError = new Error('bad chunk');
    decoder.decode(frame(true));
    expect(onError).toHaveBeenCalledWith(new Error('bad chunk'));
});

it('reset closes the decoder and forgets the codec', () => {
    decoder.configure(sps);
    decoder.reset();
    expect(browser.decoders[0].state).toBe('closed');
    expect(decoder.codec).toBeNull();
    decoder.decode(frame(true));
    expect(browser.decoders[0].decoded).toHaveLength(0);
});

it('reset without a decoder, or with a closed one, does nothing', () => {
    decoder.reset();
    decoder.configure(sps);
    browser.decoders[0].close();
    expect(() => decoder.reset()).not.toThrow();
});
