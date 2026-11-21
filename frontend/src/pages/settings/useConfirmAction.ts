import {useState} from 'react';
import {AsyncAction} from '../../hooks/useAsyncAction';

export interface ConfirmAction<T> {
    // what the open dialog asks about, null while it is closed
    target: T | null,
    ask: (target: T) => void,
    cancel: () => void,
    confirm: () => Promise<void>,
    busy: boolean,
}

// useConfirmAction drives a ConfirmDialog: ask opens it for a target, confirm
// runs the action and closes it when the action succeeded (resolved true)
export function useConfirmAction<T>(action: AsyncAction<[T], boolean>): ConfirmAction<T> {
    const [target, setTarget] = useState<T | null>(null);
    const confirm = async (): Promise<void> => {
        if (target !== null && await action.run(target)) setTarget(null);
    };
    return {target, ask: setTarget, cancel: () => setTarget(null), confirm, busy: action.busy};
}
