import {beforeEach, expect, it, Mock, vi} from 'vitest';
import {createFakeBrowser, fakeBitmap, FakeBrowser} from './fakes';
import {createMjpegDecoder, MjpegDecoder} from './mjpegDecoder';

let browser: FakeBrowser;
let onFrame: Mock<(bitmap: ImageBitmap) => void>;
let decoder: MjpegDecoder;

beforeEach(() => {
    browser = createFakeBrowser();
    onFrame = vi.fn();
    decoder = createMjpegDecoder({deps: browser.deps, onFrame});
});

const settle = (): Promise<unknown> => new Promise(resolve => setTimeout(resolve, 0));
const jpeg = (n: number): Uint8Array => new Uint8Array([n]);

it('decodes a JPEG blob and delivers the bitmap, then closes it', async () => {
    decoder.push(jpeg(1));
    const bitmap = fakeBitmap();
    expect(browser.bitmaps[0].blob.type).toBe('image/jpeg');
    browser.bitmaps[0].resolve(bitmap);
    await settle();
    expect(onFrame).toHaveBeenCalledWith(bitmap);
    expect(bitmap.close).toHaveBeenCalled();
});

it('keeps only the newest frame that arrives while one is decoding', async () => {
    decoder.push(jpeg(1));
    decoder.push(jpeg(2));
    decoder.push(jpeg(3));
    browser.bitmaps[0].resolve(fakeBitmap());
    await settle();
    expect(browser.bitmaps).toHaveLength(2);
    expect(browser.bitmaps[1].blob.size).toBe(1);
    browser.bitmaps[1].resolve(fakeBitmap());
    await settle();
    expect(onFrame).toHaveBeenCalledTimes(2);
    expect(browser.bitmaps).toHaveLength(2);
});

it('skips a broken frame with a warning and decodes the next', async () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    decoder.push(jpeg(1));
    browser.bitmaps[0].reject(new Error('bad jpeg'));
    await settle();
    expect(warn).toHaveBeenCalled();
    decoder.push(jpeg(2));
    expect(browser.bitmaps).toHaveLength(2);
    warn.mockRestore();
});

it('drops a frame that finishes after reset, and still closes it', async () => {
    decoder.push(jpeg(1));
    decoder.reset();
    const bitmap = fakeBitmap();
    browser.bitmaps[0].resolve(bitmap);
    await settle();
    expect(onFrame).not.toHaveBeenCalled();
    expect(bitmap.close).toHaveBeenCalled();
});

it('drops the queued frame on reset but decodes a frame pushed after it', async () => {
    decoder.push(jpeg(1));
    decoder.push(jpeg(2));
    decoder.reset();
    decoder.push(jpeg(3));
    browser.bitmaps[0].resolve(fakeBitmap());
    await settle();
    expect(browser.bitmaps).toHaveLength(2);
    browser.bitmaps[1].resolve(fakeBitmap());
    await settle();
    expect(onFrame).toHaveBeenCalledTimes(1);
});

it('closes the bitmap even when the frame handler throws', async () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    onFrame.mockImplementation(() => { throw new Error('draw'); });
    decoder.push(jpeg(1));
    const bitmap = fakeBitmap();
    browser.bitmaps[0].resolve(bitmap);
    await settle();
    expect(bitmap.close).toHaveBeenCalled();
    expect(warn).toHaveBeenCalled();
    warn.mockRestore();
});
