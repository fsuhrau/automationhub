import {createContext, useContext, useMemo} from 'react';
import {ApiClient} from '../../../api/ApiClient';
import {defaultUnityClient, unityService, UnityService} from '../../../services/unity.service';

// the client for the Unity editor; tests provide a fake through the context
export const UnityClientContext = createContext<ApiClient>(defaultUnityClient);

export const useUnityService = (): UnityService => {
    const client = useContext(UnityClientContext);
    return useMemo(() => unityService(client), [client]);
};
