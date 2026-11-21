import {afterEach, describe, expect, it, vi} from 'vitest';
import {defaultVideoDeps} from './deps';

afterEach(() => vi.unstubAllGlobals());

describe('defaultVideoDeps', () => {
    it('builds the browser objects from the globals', async () => {
        const made: unknown[] = [];
        const make = (name: string): ReturnType<typeof vi.fn> => vi.fn(function (this: unknown, arg: unknown) { made.push([name, arg]); });
        vi.stubGlobal('WebSocket', make('ws'));
        vi.stubGlobal('VideoDecoder', make('decoder'));
        vi.stubGlobal('EncodedVideoChunk', make('chunk'));
        vi.stubGlobal('createImageBitmap', vi.fn(async () => 'bitmap'));
        const deps = defaultVideoDeps();
        deps.createSocket('ws://x');
        deps.createDecoder({} as VideoDecoderInit);
        deps.createChunk({} as EncodedVideoChunkInit);
        await expect(deps.createImageBitmap(new Blob())).resolves.toBe('bitmap');
        expect(made.map(m => (m as string[])[0])).toEqual(['ws', 'decoder', 'chunk']);
        expect(typeof deps.now()).toBe('number');
    });
});
