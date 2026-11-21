import {afterEach, describe, expect, it, vi} from 'vitest';
import {mjpegDecodingAvailable, videoDecodingAvailable, videoUrl} from './support';

afterEach(() => vi.unstubAllGlobals());

describe('videoUrl', () => {
    it('uses ws on http and escapes the ids', () => {
        expect(videoUrl('p 1', 7)).toBe(`ws://${window.location.host}/api/p%201/device/7/video`);
    });

    it('adds the given options as a query', () => {
        const url = videoUrl('p', 'd', {maxSize: 800, maxFps: 30, bitRate: 0});
        expect(url).toMatch(/\/video\?maxSize=800&maxFps=30&bitRate=0$/);
    });

    it('uses wss on https', () => {
        vi.stubGlobal('location', {protocol: 'https:', host: 'hub.test'});
        expect(videoUrl('p', 1)).toBe('wss://hub.test/api/p/device/1/video');
    });
});

describe('decoding availability', () => {
    it('needs a secure context and VideoDecoder for H.264', () => {
        vi.stubGlobal('VideoDecoder', function () { return undefined; });
        vi.stubGlobal('isSecureContext', true);
        expect(videoDecodingAvailable()).toBe(true);
        vi.stubGlobal('isSecureContext', false);
        expect(videoDecodingAvailable()).toBe(false);
    });

    it('is unavailable without VideoDecoder', () => {
        vi.stubGlobal('isSecureContext', true);
        expect(videoDecodingAvailable()).toBe(false);
    });

    it('needs createImageBitmap for MJPEG', () => {
        expect(mjpegDecodingAvailable()).toBe(false);
        vi.stubGlobal('createImageBitmap', () => undefined);
        expect(mjpegDecodingAvailable()).toBe(true);
    });

    it('is unavailable without a window', () => {
        vi.stubGlobal('window', undefined);
        expect(videoDecodingAvailable()).toBe(false);
        expect(mjpegDecodingAvailable()).toBe(false);
    });
});

