import {useEffect, useMemo, useState} from 'react';
import {frameIndexAt, MjpegFrame, parseMjpegIndex, ReplaySegment} from '../../replayModel';
import {FrameIo, FrameLoader} from './frameLoader';

const PREFETCH = 6;

interface IndexState {
    frames: MjpegFrame[] | null,
    error: boolean,
}

// the frame list of the recording, from its index file; the segment is keyed
// by the caller, so a new segment starts with a fresh state
const useMjpegIndex = (segment: ReplaySegment, io: FrameIo): IndexState => {
    const [state, setState] = useState<IndexState>({frames: null, error: !segment.indexUrl});
    useEffect(() => {
        let cancelled = false;
        const set = (next: IndexState): void => {
            if (!cancelled) setState(next);
        };
        if (!segment.indexUrl) return;
        io.fetch(segment.indexUrl, {credentials: 'same-origin'})
            .then(r => r.ok ? r.text() : Promise.reject(new Error(r.statusText)))
            .then(text => set({frames: parseMjpegIndex(text, segment.start), error: false}))
            .catch(() => set({frames: null, error: true}));
        return () => {
            cancelled = true;
        };
    }, [segment.indexUrl, segment.start, io]);
    return state;
};

export interface PlayClock {
    position: number,
    playing: boolean,
}

export interface MjpegFrameState {
    // the object URL of the frame on screen
    src: string | null,
    error: boolean,
}

// useMjpegFrame is the frame of an MJPEG recording at the replay position;
// while playing it fetches the next frames ahead.
export const useMjpegFrame = (segment: ReplaySegment, io: FrameIo, clock: PlayClock): MjpegFrameState => {
    const {position, playing} = clock;
    const {frames, error} = useMjpegIndex(segment, io);
    const [src, setSrc] = useState<string | null>(null);
    const loader = useMemo(() => frames ? new FrameLoader({url: segment.url, frames, io}) : null, [frames, segment.url, io]);
    useEffect(() => () => loader?.dispose(), [loader]);
    const index = frames ? frameIndexAt(frames, position) : -1;
    useEffect(() => {
        if (!loader || !frames || index < 0) return;
        let current = true;
        loader.load(index).then(url => current && setSrc(url)).catch(() => undefined);
        const ahead = playing ? Math.min(frames.length, index + 1 + PREFETCH) : 0;
        for (let i = index + 1; i < ahead; i++) loader.load(i).catch(() => undefined);
        return () => {
            current = false;
        };
    }, [loader, frames, index, playing]);
    return {src, error};
};
