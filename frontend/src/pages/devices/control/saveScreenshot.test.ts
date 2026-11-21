import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {fetchScreenshot, saveBlob, screenshotName} from './saveScreenshot';

afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe('screenshotName', () => {
    it('is the device name and the time, safe for a file', () => {
        expect(screenshotName('Lab phone/1', new Date(2026, 0, 2, 3, 4, 5).getTime())).toBe('Lab_phone_1-20260102-030405.png');
    });
});

describe('saveBlob', () => {
    beforeEach(() => {
        vi.useFakeTimers();
    });

    it('downloads the blob under the name and frees the url', () => {
        const createObjectURL = vi.fn(() => 'blob:x');
        const revokeObjectURL = vi.fn();
        vi.stubGlobal('URL', {createObjectURL, revokeObjectURL});
        const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) {
            expect(this.download).toBe('a.png');
            expect(this.href).toBe('blob:x');
        });
        const blob = new Blob(['x']);
        saveBlob(blob, 'a.png');
        expect(createObjectURL).toHaveBeenCalledWith(blob);
        expect(click).toHaveBeenCalled();
        expect(revokeObjectURL).not.toHaveBeenCalled();
        vi.advanceTimersByTime(1000);
        expect(revokeObjectURL).toHaveBeenCalledWith('blob:x');
    });
});

describe('fetchScreenshot', () => {
    const canvasWith = (blob: Blob | null): HTMLCanvasElement => {
        const canvas = document.createElement('canvas');
        canvas.toBlob = cb => cb(blob);
        return canvas;
    };
    const stubFetch = (response: Partial<Response>): ReturnType<typeof vi.fn> => {
        const fetchMock = vi.fn(async () => response as Response);
        vi.stubGlobal('fetch', fetchMock);
        return fetchMock;
    };

    it('takes the frame of a playing video', async () => {
        const frame = new Blob(['frame']);
        const fetchMock = stubFetch({});
        expect(await fetchScreenshot({canvas: canvasWith(frame), url: '/shot'})).toBe(frame);
        expect(fetchMock).not.toHaveBeenCalled();
    });

    it('asks the hub when there is no canvas or the canvas gives nothing', async () => {
        const shot = new Blob(['shot']);
        const fetchMock = stubFetch({ok: true, blob: async () => shot});
        expect(await fetchScreenshot({canvas: null, url: '/shot'})).toBe(shot);
        expect(await fetchScreenshot({canvas: canvasWith(null), url: '/shot'})).toBe(shot);
        expect(fetchMock).toHaveBeenCalledWith('/shot', {credentials: 'include'});
    });

    it('fails when the hub does', async () => {
        stubFetch({ok: false, status: 502});
        await expect(fetchScreenshot({canvas: null, url: '/shot'})).rejects.toThrow('screenshot failed: 502');
    });
});
