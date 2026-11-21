import axios from 'axios';
import {ApiClient, createApiClient} from '../api/ApiClient';
import IAppFunctionData from '../types/app.function';

export interface UnityService {
    // the tests of the Unity editor open on this computer
    testFunctions: () => Promise<IAppFunctionData[]>,
}

export const unityService = (api: ApiClient): UnityService => ({
    testFunctions: () => api.get('/tests'),
});

// the Unity editor plugin listens on this computer, not on the hub
export const defaultUnityClient: ApiClient = createApiClient(axios.create({
    baseURL: 'http://localhost:7109/',
    headers: {
        'Content-type': 'application/json',
    },
}));
