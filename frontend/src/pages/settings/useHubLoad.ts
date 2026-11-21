import {DependencyList, useEffect} from 'react';
import {HubStateAction} from '../../application/HubState';
import {useError} from '../../ErrorProvider';
import {errorMessage} from '../../lib/errorMessage';
import {useHubState} from '../../hooks/HubStateProvider';
import {useLoader} from '../../hooks/useLoader';

// useHubLoad loads data when deps change and puts it into the hub state;
// a failure shows in the error toast
export function useHubLoad<T>(load: () => Promise<T>, toAction: (data: T) => HubStateAction, deps: DependencyList): void {
    const {dispatch} = useHubState();
    const {setError} = useError();
    const {error} = useLoader(() => load().then(data => {
        dispatch(toAction(data));
        return data;
    }), deps);
    useEffect(() => {
        if (error) setError(errorMessage(error));
    }, [error, setError]);
}
