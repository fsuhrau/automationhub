import React, {useEffect, useMemo, useRef, useState} from 'react';
import {Box, CircularProgress, Typography} from '@mui/material';
import {frameIndexAt, MjpegFrame, parseMjpegIndex, ReplaySegment, segmentAt} from '../pages/tests/replayModel';

// ReplayVideo shows a device's recording at a time of the replay clock. The
// clock is the page's: the video follows it, and is only moved when it is off
// by more than a little, so playing stays smooth.
interface ReplayVideoProps {
    segments: ReplaySegment[],
    // ms since the epoch
    position: number,
    playing: boolean,
    rate: number,
    label: string,
}

const DRIFT_PLAYING_S = 0.5;
const DRIFT_PAUSED_S = 0.04;

const Frame: React.FC<{ width: number, height: number, children: React.ReactNode }> = ({width, height, children}) =>
    <Box sx={{
        position: 'relative', bgcolor: '#000', borderRadius: 1, overflow: 'hidden', mx: 'auto',
        // portrait phones would be very tall; keep the whole screen in view
        maxHeight: '70vh', aspectRatio: width > 0 && height > 0 ? `${width} / ${height}` : '9 / 16', maxWidth: '100%',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        '& > video, & > img': {width: '100%', height: '100%', objectFit: 'contain', display: 'block'},
    }}>{children}</Box>;

const Mp4Segment: React.FC<{ segment: ReplaySegment, position: number, playing: boolean, rate: number, label: string }> =
    ({segment, position, playing, rate, label}) => {
        const video = useRef<HTMLVideoElement>(null);
        const [error, setError] = useState(false);
        const [loading, setLoading] = useState(true);

        useEffect(() => {
            const v = video.current;
            if (!v || v.readyState < 1) return;
            const expected = Math.max(0, (position - segment.start) / 1000);
            if (Math.abs(v.currentTime - expected) > (playing ? DRIFT_PLAYING_S : DRIFT_PAUSED_S)) {
                v.currentTime = expected;
            }
        }, [position, playing, segment.start, loading]);

        useEffect(() => {
            const v = video.current;
            if (!v) return;
            v.playbackRate = rate;
            if (playing) {
                v.play().catch(() => undefined);
            } else {
                v.pause();
            }
        }, [playing, rate, loading]);

        if (error) {
            return <Typography variant="body2" sx={{color: 'grey.400', p: 2}}>This recording can't be played.</Typography>;
        }
        return <>
            <video ref={video} src={segment.url} muted playsInline preload="auto" aria-label={label}
                   onLoadedMetadata={() => setLoading(false)} onError={() => setError(true)}/>
            {loading && <CircularProgress size={28} sx={{position: 'absolute', color: 'grey.400'}}/>}
        </>;
    };

const FRAME_CACHE = 90;
const PREFETCH = 6;

const MjpegSegment: React.FC<{ segment: ReplaySegment, position: number, playing: boolean, label: string }> =
    ({segment, position, playing, label}) => {
        const [frames, setFrames] = useState<MjpegFrame[] | null>(null);
        const [error, setError] = useState(false);
        const [shown, setShown] = useState<string | null>(null);
        // frame index -> object URL, oldest first
        const cache = useRef(new Map<number, string>());
        const pending = useRef(new Map<number, Promise<string>>());

        useEffect(() => {
            let cancelled = false;
            setFrames(null);
            setError(false);
            if (!segment.indexUrl) {
                setError(true);
                return;
            }
            fetch(segment.indexUrl, {credentials: 'same-origin'})
                .then(r => r.ok ? r.text() : Promise.reject(new Error(r.statusText)))
                .then(text => !cancelled && setFrames(parseMjpegIndex(text, segment.start)))
                .catch(() => !cancelled && setError(true));
            const urls = cache.current;
            return () => {
                cancelled = true;
                urls.forEach(u => URL.revokeObjectURL(u));
                urls.clear();
                pending.current.clear();
            };
        }, [segment.indexUrl, segment.start]);

        const load = (index: number): Promise<string> => {
            const cached = cache.current.get(index);
            if (cached) return Promise.resolve(cached);
            const running = pending.current.get(index);
            if (running) return running;
            const f = frames![index];
            const p = fetch(segment.url, {credentials: 'same-origin', headers: {Range: `bytes=${f.offset}-${f.offset + f.length - 1}`}})
                .then(r => r.ok
                    // a server that ignores the range sends the whole file
                    ? r.blob().then(b => r.status === 206 ? b : b.slice(f.offset, f.offset + f.length))
                    : Promise.reject(new Error(r.statusText)))
                .then(blob => {
                    const url = URL.createObjectURL(new Blob([blob], {type: 'image/jpeg'}));
                    cache.current.set(index, url);
                    while (cache.current.size > FRAME_CACHE) {
                        const [oldest, oldUrl] = cache.current.entries().next().value as [number, string];
                        cache.current.delete(oldest);
                        URL.revokeObjectURL(oldUrl);
                    }
                    return url;
                })
                .finally(() => pending.current.delete(index));
            pending.current.set(index, p);
            return p;
        };

        const index = frames ? frameIndexAt(frames, position) : -1;
        useEffect(() => {
            if (!frames || index < 0) return;
            let current = true;
            load(index).then(url => current && setShown(url)).catch(() => undefined);
            if (playing) {
                for (let i = index + 1; i < Math.min(frames.length, index + 1 + PREFETCH); i++) {
                    load(i).catch(() => undefined);
                }
            }
            return () => {
                current = false;
            };
            // load only depends on frames and the segment
            // eslint-disable-next-line react-hooks/exhaustive-deps
        }, [frames, index, playing]);

        if (error) {
            return <Typography variant="body2" sx={{color: 'grey.400', p: 2}}>This recording can't be played.</Typography>;
        }
        return shown
            ? <img src={shown} alt={label}/>
            : <CircularProgress size={28} sx={{color: 'grey.400'}}/>;
    };

const ReplayVideo: React.FC<ReplayVideoProps> = ({segments, position, playing, rate, label}) => {
    const segment = segmentAt(segments, position);
    // keep the size of the frame between segments, so the page doesn't jump
    const sized = useMemo(() => segment ?? segments.find(s => s.start > position) ?? segments[segments.length - 1], [segment, segments, position]);
    return <Frame width={sized?.width ?? 0} height={sized?.height ?? 0}>
        {!segment
            ? <Typography variant="body2" sx={{color: 'grey.400', p: 2, textAlign: 'center'}}>
                {segments.length === 0 ? 'This device has no recording.' : 'No video at this point of the run.'}
            </Typography>
            : segment.codec === 'mp4'
                ? <Mp4Segment key={segment.id} segment={segment} position={position} playing={playing} rate={rate} label={label}/>
                : <MjpegSegment key={segment.id} segment={segment} position={position} playing={playing} label={label}/>}
    </Frame>;
};

export default ReplayVideo;
