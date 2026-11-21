import {useCallback, useRef} from 'react';
import {ErrorHandler} from './useGuardedAction';

// input errors of the video socket arrive per event; show one at a time
export const INPUT_ERROR_GAP_MS = 3000;

// useInputError returns the handler for input errors of the video socket: it
// reports one error per gap and reads the lock status again (the lease may be
// gone).
export const useInputError = (report: ErrorHandler, refresh: () => void): ((message: string) => void) => {
    const last = useRef(0);
    return useCallback((message: string) => {
        const now = Date.now();
        if (now - last.current < INPUT_ERROR_GAP_MS) return;
        last.current = now;
        report(message);
        refresh();
    }, [report, refresh]);
};
