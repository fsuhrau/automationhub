import type {VideoOptions} from './types';

// WebCodecs only exists in secure contexts (https or localhost); it decodes
// the H.264 video of Android devices.
export const videoDecodingAvailable = (): boolean =>
    typeof window !== 'undefined' && window.isSecureContext && typeof window.VideoDecoder === 'function';

// MJPEG (iOS) only needs createImageBitmap, which works over plain http too.
export const mjpegDecodingAvailable = (): boolean =>
    typeof window !== 'undefined' && typeof window.createImageBitmap === 'function';

const queryOf = (opts: VideoOptions): string => {
    const params = new URLSearchParams();
    if (opts.maxSize !== undefined) params.set('maxSize', String(opts.maxSize));
    if (opts.maxFps !== undefined) params.set('maxFps', String(opts.maxFps));
    if (opts.bitRate !== undefined) params.set('bitRate', String(opts.bitRate));
    const query = params.toString();
    return query ? `?${query}` : '';
};

export const videoUrl = (projectId: string, deviceId: number | string, opts: VideoOptions = {}): string => {
    const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const path = `/api/${encodeURIComponent(projectId)}/device/${encodeURIComponent(String(deviceId))}/video`;
    return `${scheme}//${window.location.host}${path}${queryOf(opts)}`;
};
