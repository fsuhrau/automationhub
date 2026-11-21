import {useNow} from '../../../hooks/useNow';

const MINUTE_MS = 60_000;

// useAgoClock is the time "x minutes ago" is counted from; it moves on
// every minute
export const useAgoClock = (): number => useNow(MINUTE_MS);
