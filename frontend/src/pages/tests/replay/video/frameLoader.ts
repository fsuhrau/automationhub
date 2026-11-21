import {MjpegFrame} from '../../replayModel';

// the browser services the frame loading needs; tests inject fakes
export interface FrameIo {
    fetch: (url: string, init?: RequestInit) => Promise<Response>,
    createObjectUrl: (blob: Blob) => string,
    revokeObjectUrl: (url: string) => void,
}

export const browserFrameIo: FrameIo = {
    fetch: (url, init) => fetch(url, init),
    createObjectUrl: blob => URL.createObjectURL(blob),
    revokeObjectUrl: url => URL.revokeObjectURL(url),
};

export const FRAME_CACHE = 90;

interface FrameLoaderOptions {
    // the recording file
    url: string,
    frames: MjpegFrame[],
    io: FrameIo,
    limit?: number,
}

// FrameLoader fetches single JPEG frames of an MJPEG recording by byte range
// and keeps the newest as object URLs.
export class FrameLoader {
    // frame index -> object URL, oldest first
    private readonly cache = new Map<number, string>();
    private readonly pending = new Map<number, Promise<string>>();

    constructor(private readonly o: FrameLoaderOptions) {
    }

    load(index: number): Promise<string> {
        const cached = this.cache.get(index);
        if (cached) return Promise.resolve(cached);
        const running = this.pending.get(index);
        if (running) return running;
        const p = this.fetchFrame(index).finally(() => this.pending.delete(index));
        this.pending.set(index, p);
        return p;
    }

    dispose(): void {
        this.cache.forEach(u => this.o.io.revokeObjectUrl(u));
        this.cache.clear();
        this.pending.clear();
    }

    private async fetchFrame(index: number): Promise<string> {
        const {offset, length} = this.o.frames[index];
        const r = await this.o.io.fetch(this.o.url, {credentials: 'same-origin', headers: {Range: `bytes=${offset}-${offset + length - 1}`}});
        if (!r.ok) throw new Error(r.statusText);
        const body = await r.blob();
        // a server that ignores the range sends the whole file
        const jpeg = r.status === 206 ? body : body.slice(offset, offset + length);
        const url = this.o.io.createObjectUrl(new Blob([jpeg], {type: 'image/jpeg'}));
        this.remember(index, url);
        return url;
    }

    private remember(index: number, url: string): void {
        this.cache.set(index, url);
        while (this.cache.size > (this.o.limit ?? FRAME_CACHE)) {
            const [oldest, oldUrl] = this.cache.entries().next().value as [number, string];
            this.cache.delete(oldest);
            this.o.io.revokeObjectUrl(oldUrl);
        }
    }
}
