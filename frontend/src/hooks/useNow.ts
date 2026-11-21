import {useEffect, useState} from 'react';

// useNow returns the current time in ms and refreshes it every intervalMs,
// for live durations. While `active` is false it stops ticking.
export const useNow = (intervalMs: number, active = true): number => {
    const [now, setNow] = useState(() => Date.now());
    useEffect(() => {
        if (!active) return;
        const timer = setInterval(() => setNow(Date.now()), intervalMs);
        return () => clearInterval(timer);
    }, [intervalMs, active]);
    return now;
};
