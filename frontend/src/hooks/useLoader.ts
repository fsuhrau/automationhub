import {DependencyList, useCallback, useEffect, useState} from 'react';

export interface Loaded<T> {
    data: T | null,
    error: unknown,
    // true while a load for the current deps is in flight; data may still be
    // the previous deps' result then
    loading: boolean,
    retry: () => void,
}

// useLoader runs load whenever deps change (or retry is called) and keeps
// the result or the error. Responses for outdated deps are ignored.
export function useLoader<T>(load: () => Promise<T> | null, deps: DependencyList): Loaded<T> {
    const [data, setData] = useState<T | null>(null);
    const [error, setError] = useState<unknown>(null);
    const [attempt, setAttempt] = useState(0);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        const promise = load();
        if (promise === null) {
            return;
        }
        let current = true;
        setError(null);
        setLoading(true);
        promise.then(result => {
            if (current) setData(result);
        }).catch(e => {
            if (current) setError(e);
        }).finally(() => {
            if (current) setLoading(false);
        });
        return () => {
            current = false;
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [...deps, attempt]);

    const retry = useCallback(() => setAttempt(a => a + 1), []);
    return {data, error, loading, retry};
}
