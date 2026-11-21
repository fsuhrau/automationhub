import React, {createContext, ReactNode, useContext} from 'react';
import axiosInstance from '../http-common';
import {ApiClient, createApiClient} from './ApiClient';

// without a provider, features talk to the real axios instance
const ApiClientContext = createContext<ApiClient>(createApiClient(axiosInstance));

interface ApiClientProviderProps {
    // tests inject a fake
    client: ApiClient,
    children: ReactNode,
}

export const ApiClientProvider: React.FC<ApiClientProviderProps> = ({client, children}) => {
    return <ApiClientContext.Provider value={client}>{children}</ApiClientContext.Provider>;
};

export const useApiClient = (): ApiClient => useContext(ApiClientContext);
