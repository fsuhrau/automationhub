import {afterEach, expect, it, vi} from 'vitest';
import {browserFrameIo, FrameIo, FrameLoader} from './frameLoader';

const FRAMES = [0, 1, 2].map(i => ({time: i * 40, offset: i * 10, length: 10}));

const response = (body: string, status = 206, ok = true): Response => ({
    ok, status, statusText: 'bad', blob: () => Promise.resolve(new Blob([body])),
} as unknown as Response);

const fakeIo = (res: Response = response('0123456789')): FrameIo => {
    let n = 0;
    const io: FrameIo = {
        fetch: vi.fn().mockResolvedValue(res),
        createObjectUrl: vi.fn(() => `blob:${n++}`),
        revokeObjectUrl: vi.fn(),
    };
    return io;
};

it('FrameLoader: fetches the byte range of a frame and caches its object url', async () => {
    const io = fakeIo();
    const loader = new FrameLoader({url: '/f.mjpeg', frames: FRAMES, io});
    expect(await loader.load(1)).toBe('blob:0');
    expect(await loader.load(1)).toBe('blob:0');
    expect(io.fetch).toHaveBeenCalledTimes(1);
    expect(io.fetch).toHaveBeenCalledWith('/f.mjpeg', {credentials: 'same-origin', headers: {Range: 'bytes=10-19'}});
});

it('FrameLoader: shares a running request', async () => {
    const io = fakeIo();
    const loader = new FrameLoader({url: '/f', frames: FRAMES, io});
    const urls = await Promise.all([loader.load(0), loader.load(0)]);
    expect(urls).toEqual(['blob:0', 'blob:0']);
    expect(io.fetch).toHaveBeenCalledTimes(1);
});

it('FrameLoader: cuts the frame out of a whole file when the server ignores the range', async () => {
    const io = fakeIo(response('a'.repeat(10) + 'b'.repeat(10) + 'c'.repeat(10), 200));
    const loader = new FrameLoader({url: '/f', frames: FRAMES, io});
    await loader.load(1);
    const blob = (io.createObjectUrl as ReturnType<typeof vi.fn>).mock.calls[0][0] as Blob;
    expect(blob.size).toBe(10);
    expect(blob.type).toBe('image/jpeg');
});

it('FrameLoader: rejects on a failed response and retries later', async () => {
    const io = fakeIo(response('', 500, false));
    const loader = new FrameLoader({url: '/f', frames: FRAMES, io});
    await expect(loader.load(0)).rejects.toThrow('bad');
    await expect(loader.load(0)).rejects.toThrow('bad');
    expect(io.fetch).toHaveBeenCalledTimes(2);
});

it('FrameLoader: revokes the oldest urls beyond the limit', async () => {
    const io = fakeIo();
    const loader = new FrameLoader({url: '/f', frames: FRAMES, io, limit: 2});
    await loader.load(0);
    await loader.load(1);
    await loader.load(2);
    expect(io.revokeObjectUrl).toHaveBeenCalledWith('blob:0');
    expect(io.revokeObjectUrl).toHaveBeenCalledTimes(1);
});

it('FrameLoader: revokes everything on dispose', async () => {
    const io = fakeIo();
    const loader = new FrameLoader({url: '/f', frames: FRAMES, io});
    await loader.load(0);
    await loader.load(1);
    loader.dispose();
    expect(io.revokeObjectUrl).toHaveBeenCalledTimes(2);
    await loader.load(0);
    expect(io.fetch).toHaveBeenCalledTimes(3);
});

afterEach(() => {
    vi.unstubAllGlobals();
});

it('browserFrameIo: uses the global fetch and object urls', async () => {
    const fetchFn = vi.fn().mockResolvedValue('res');
    vi.stubGlobal('fetch', fetchFn);
    const create = vi.fn().mockReturnValue('blob:x');
    const revoke = vi.fn();
    vi.stubGlobal('URL', {createObjectURL: create, revokeObjectURL: revoke});
    expect(await browserFrameIo.fetch('/u', {credentials: 'omit'})).toBe('res');
    expect(fetchFn).toHaveBeenCalledWith('/u', {credentials: 'omit'});
    const blob = new Blob(['x']);
    expect(browserFrameIo.createObjectUrl(blob)).toBe('blob:x');
    browserFrameIo.revokeObjectUrl('blob:x');
    expect(create).toHaveBeenCalledWith(blob);
    expect(revoke).toHaveBeenCalledWith('blob:x');
});
