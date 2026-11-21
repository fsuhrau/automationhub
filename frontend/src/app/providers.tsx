import React, {ReactNode} from "react";
import {ApiClient, createApiClient} from "../api/ApiClient";
import {ApiClientProvider} from "../api/ApiClientProvider";
import http from "../http-common";
import {ErrorProvider} from "../ErrorProvider";
import {HubStateProvider} from "../hooks/HubStateProvider";

interface AppProvidersProps {
    // tests inject a fake; the app talks to the real server
    client?: ApiClient,
    children: ReactNode,
}

const realClient = createApiClient(http);

// what the whole app shares: the API client, the error display and the hub state
export const AppProviders: React.FC<AppProvidersProps> = ({client = realClient, children}) => (
    <ApiClientProvider client={client}>
        <ErrorProvider>
            <HubStateProvider>{children}</HubStateProvider>
        </ErrorProvider>
    </ApiClientProvider>
);
