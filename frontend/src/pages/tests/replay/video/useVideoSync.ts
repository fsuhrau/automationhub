import {RefObject, useEffect, useRef} from 'react';
import {SeekClock, videoSeekTarget} from '../../replayModel';

interface VideoSync extends SeekClock {
    rate: number,
    // the video has loaded its metadata
    ready: boolean,
}

// useVideoSync makes a video element follow the replay clock: it jumps only
// when off by more than a little, and plays, pauses and changes speed with it.
// It returns the ref for the element.
export const useVideoSync = (s: VideoSync): RefObject<HTMLVideoElement | null> => {
    const video = useRef<HTMLVideoElement>(null);
    const {position, segmentStart, playing, rate, ready} = s;
    useEffect(() => {
        const v = video.current;
        if (!v || v.readyState < 1) return;
        const target = videoSeekTarget(v.currentTime, {position, segmentStart, playing});
        if (target !== null) v.currentTime = target;
    }, [position, playing, segmentStart, ready]);

    useEffect(() => {
        const v = video.current;
        if (!v) return;
        v.playbackRate = rate;
        if (playing) {
            v.play().catch(() => undefined);
        } else {
            v.pause();
        }
    }, [playing, rate, ready]);
    return video;
};
