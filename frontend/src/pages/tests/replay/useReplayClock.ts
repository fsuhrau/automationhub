import {useCallback, useEffect, useRef, useState} from 'react';
import {ReplayRange} from '../replayModel';

export const RATES = [0.5, 1, 2, 4];
export const TICK_MS = 100;

export interface ReplayClock {
    // ms since the epoch
    position: number,
    playing: boolean,
    rate: number,
    setRate: (rate: number) => void,
    seek: (time: number) => void,
    toggle: () => void,
}

const systemNow = (): number => Date.now();

// useReplayClock is the time of the replay, in ms since the epoch. While
// playing it advances by the wall time passed times the rate, ten times a
// second, which is enough for the slider and the log. It stops at the end.
export const useReplayClock = ({start, end}: ReplayRange, initial: number, now: () => number = systemNow): ReplayClock => {
    const [position, setPosition] = useState(initial);
    const [playing, setPlaying] = useState(false);
    const [rate, setRate] = useState(1);
    const pos = useRef(initial);

    const moveTo = useCallback((time: number) => {
        pos.current = time;
        setPosition(time);
    }, []);

    useEffect(() => {
        if (!playing) return;
        let last = now();
        const id = setInterval(() => {
            const t = now();
            moveTo(Math.min(end, pos.current + (t - last) * rate));
            last = t;
            if (pos.current >= end) setPlaying(false);
        }, TICK_MS);
        return () => clearInterval(id);
    }, [playing, rate, end, now, moveTo]);

    const seek = useCallback((time: number) => moveTo(Math.max(start, Math.min(end, time))), [start, end, moveTo]);

    const toggle = useCallback(() => {
        if (!playing && pos.current >= end) moveTo(start);
        setPlaying(p => !p);
    }, [playing, start, end, moveTo]);

    return {position, playing, rate, setRate, seek, toggle};
};
