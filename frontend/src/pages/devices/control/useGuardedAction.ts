import {useCallback} from 'react';
import {useAsyncAction} from '../../../hooks/useAsyncAction';

export type ErrorHandler = (error: unknown) => void;

// Runs one action of the page. A failure goes to the handler of the call, or
// else to the default one; run itself never rejects.
export type Runner = (action: () => Promise<unknown>, onError?: ErrorHandler) => Promise<unknown>;

// useGuardedAction tracks busy for all actions of the page (one at a time
// is shown), and sends failures to onError unless a call brings its own.
export const useGuardedAction = (onError: ErrorHandler): { run: Runner, busy: boolean } => {
    const guarded = useCallback(async (action: () => Promise<unknown>, handler: ErrorHandler = onError): Promise<void> => {
        try {
            await action();
        } catch (error) {
            handler(error);
        }
    }, [onError]);
    const {run, busy} = useAsyncAction(guarded);
    return {run, busy};
};
