import {afterEach, describe, expect, it, vi} from 'vitest';
import {realVideoSupport} from './device.control.video';

class FakeSocket {
    static OPEN = 1;
    static instances: FakeSocket[] = [];
    readyState = 0;
    binaryType = '';
    onmessage: unknown = null;
    onclose: unknown = null;
    closed = false;

    constructor(readonly url: string) {
        FakeSocket.instances.push(this);
    }

    close(): void {
        this.closed = true;
    }
}

afterEach(() => {
    vi.unstubAllGlobals();
    FakeSocket.instances = [];
});

describe('realVideoSupport', () => {
    it('asks the browser what it can decode', () => {
        vi.stubGlobal('createImageBitmap', () => undefined);
        const support = realVideoSupport();
        expect(support.decodesMjpeg).toBe(true);
        expect(support.decodesVideo).toBeFalsy();
    });

    it('decodes video in a secure context with WebCodecs', () => {
        vi.stubGlobal('isSecureContext', true);
        vi.stubGlobal('VideoDecoder', function VideoDecoder() {
        });
        expect(realVideoSupport().decodesVideo).toBe(true);
    });

    it('plays with a DeviceVideoPlayer on the canvas', () => {
        vi.stubGlobal('WebSocket', FakeSocket);
        const states: string[] = [];
        const canvas = document.createElement('canvas');
        vi.spyOn(canvas, 'getContext').mockReturnValue(null);
        const player = realVideoSupport().createPlayer('ws://x/video', canvas, {
            onStatus: s => states.push(s.state), onMeta: () => undefined, onInputError: () => undefined,
        });
        expect(FakeSocket.instances[0].url).toBe('ws://x/video');
        expect(states).toEqual(['connecting']);
        player.close();
        expect(FakeSocket.instances[0].closed).toBe(true);
    });
});
