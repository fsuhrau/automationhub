import {useState} from 'react';
import {useLoader} from '../../../hooks/useLoader';
import IAppFunctionData from '../../../types/app.function';
import {useUnityService} from './useUnityService';

export interface UnityTests {
    // the tests Unity reported or the user pasted; null while nothing is known
    available: IAppFunctionData[] | null,
    loading: boolean,
    unreachable: boolean,
    reload: () => void,
    paste: (functions: IAppFunctionData[]) => void,
}

// useUnityTests asks the Unity editor on this computer for its tests; when
// it doesn't answer, a pasted list takes its place
export const useUnityTests = (): UnityTests => {
    const service = useUnityService();
    const [pasted, setPasted] = useState<IAppFunctionData[] | null>(null);
    const {data, error, loading, retry} = useLoader(() => service.testFunctions(), [service]);
    return {available: pasted ?? data, loading, unreachable: !!error, reload: retry, paste: setPasted};
};
