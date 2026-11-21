import {useEffect} from 'react';
import {useError} from '../../ErrorProvider';
import {AsyncAction, useAsyncAction} from '../../hooks/useAsyncAction';

// useReportedAction is useAsyncAction whose failure also shows in the
// page-wide error banner
export function useReportedAction<A extends unknown[], R>(fn: (...args: A) => Promise<R>): AsyncAction<A, R> {
    const action = useAsyncAction(fn);
    const {setError} = useError();
    const {error} = action;
    useEffect(() => {
        if (error) setError(error);
    }, [error, setError]);
    return action;
}
