import {render, RenderResult} from '@testing-library/react';
import React from 'react';
import {Location, MemoryRouter} from 'react-router-dom';
import {ApiClient} from '../../api/ApiClient';
import {AuthClient} from '../authClient';
import {AppProviders} from '../providers';
import {AuthProvider} from '../../hooks/AuthProvider';
import {fakeApi} from './fakeApi';
import {fakeAuthClient, signedIn} from './fakeAuth';

export interface InAppOptions {
    // a URL, or a location with the state a page was sent with
    path: string | Partial<Location>,
    api?: ApiClient,
    auth?: AuthClient,
}

// Renders ui like App does (providers, router, auth), but at a fake URL, with a fake API
// client and a fake auth session (signed in unless given).
export const renderInApp = (ui: React.ReactElement, {path, api = fakeApi({}), auth = fakeAuthClient(signedIn)}: InAppOptions): RenderResult =>
    render(<AppProviders client={api}>
        <MemoryRouter initialEntries={[path]}>
            <AuthProvider client={auth}>{ui}</AuthProvider>
        </MemoryRouter>
    </AppProviders>);
