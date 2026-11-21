import {useCallback, useEffect, useRef, useState} from 'react';
import {errorMessage} from '../lib/errorMessage';

export interface AsyncAction<A extends unknown[], R> {
    // resolves with the result, or undefined if fn failed (see error)
    run: (...args: A) => Promise<R | undefined>,
    busy: boolean,
    // the failure as text (errorMessage), null while there is none
    error: string | null,
}

// useAsyncAction wraps an async function for a button: it tracks busy,
// keeps the last failure as text, and never rejects.
export function useAsyncAction<A extends unknown[], R>(fn: (...args: A) => Promise<R>): AsyncAction<A, R> {
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const mounted = useRef(true);
    useEffect(() => {
        mounted.current = true;
        return () => {
            mounted.current = false;
        };
    }, []);

    const run = useCallback(async (...args: A): Promise<R | undefined> => {
        setBusy(true);
        setError(null);
        try {
            return await fn(...args);
        } catch (e) {
            if (mounted.current) setError(errorMessage(e));
            return undefined;
        } finally {
            if (mounted.current) setBusy(false);
        }
    }, [fn]);
    return {run, busy, error};
}
