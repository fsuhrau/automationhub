import {useEffect} from 'react';
import {useError} from '../../ErrorProvider';
import {AsyncAction, useAsyncAction} from '../../hooks/useAsyncAction';

// useReportedAction is useAsyncAction whose failure also shows in the
// page-wide error toast
export function useReportedAction<A extends unknown[], R>(fn: (...args: A) => Promise<R>): AsyncAction<A, R> {
    const {setError} = useError();
    const action = useAsyncAction(fn);
    useEffect(() => {
        if (action.error) setError(action.error);
    }, [action.error, setError]);
    return action;
}
