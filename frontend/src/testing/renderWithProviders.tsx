import React, {ReactElement, ReactNode, useState} from 'react';
import {render, RenderResult, screen} from '@testing-library/react';
import {SSEProvider} from 'react-hooks-sse';
import {Location, MemoryRouter, Outlet, Route, Routes, useLocation} from 'react-router-dom';
import {AuthClient} from '../app/authClient';
import {ApiClient} from '../api/ApiClient';
import {AppProviders} from '../app/providers';
import {AuthProvider} from '../hooks/AuthProvider';
import ApplicationPage from '../hooks/ApplicationProvider';
import {ProjectProvider} from '../hooks/ProjectProvider';
import {fakeApi} from './fakeApi';
import {FakeSse} from './fakeSse';
import {HubSeed, SeedHub} from './SeedHub';

export interface ProviderOptions {
    api?: ApiClient,
    sse?: FakeSse,
    // the url the router starts at, or a location with the state a page was sent with
    url?: string | Partial<Location>,
    // the session of the user; without it there is no AuthProvider
    auth?: AuthClient,
    // the hub state the page finds, as after the app's first load
    hub?: HubSeed,
    // shows where the router is, see currentLocation
    probe?: boolean,
}

// LocationProbe shows where the router is, so a test can see a navigation
export const LocationProbe: React.FC = () => {
    const {pathname, search} = useLocation();
    return <output aria-label="location" data-testid="location">{pathname + search}</output>;
};

export const currentLocation = (): string => screen.getByTestId('location').textContent ?? '';

// the providers and router of a page, as App wires them, over fakes
export const Providers: React.FC<ProviderOptions & { children: ReactNode }> = (options) => {
    // the fallbacks are made once, so a rerender keeps the same client and stream
    const [fallback] = useState(() => ({api: fakeApi(), sse: new FakeSse()}));
    const {api = fallback.api, sse = fallback.sse, url = '/', auth, hub = {}, probe = false, children} = options;
    const content = auth ? <AuthProvider client={auth}>{children}</AuthProvider> : children;
    return <AppProviders client={api}><SSEProvider source={sse.source}>
        <SeedHub {...hub}><MemoryRouter initialEntries={[url]}>
            {content}
            {probe && <LocationProbe/>}
        </MemoryRouter></SeedHub>
    </SSEProvider></AppProviders>;
};

// renderWithProviders renders ui, usually a <Routes>, inside Providers
export const renderWithProviders = (ui: ReactElement, options: ProviderOptions = {}): RenderResult =>
    render(<Providers {...options}>{ui}</Providers>);

// the route of a project; its pages are the children
export const projectRoute = (children: ReactNode): ReactElement =>
    <Route path="/project/:project_identifier" element={<ProjectProvider><Outlet/></ProjectProvider>}>{children}</Route>;

// the route of an app inside a project route
export const appRoute = (children: ReactNode): ReactElement =>
    <Route path=":appId" element={<ApplicationPage/>}>{children}</Route>;

// a page that needs no project (the project list, the profile) at "/"; any
// other location shows where the router went, see currentLocation
export const renderAtRoot = (page: ReactElement, options: ProviderOptions = {}): RenderResult =>
    renderWithProviders(<Routes><Route path="/" element={page}/><Route path="*" element={<LocationProbe/>}/></Routes>, options);

// the wrapper of renderHook for hooks that need the providers and the router
export const wrapperOf = (options: ProviderOptions = {}): React.FC<{ children: ReactNode }> => {
    const Wrapper: React.FC<{ children: ReactNode }> = ({children}) => <Providers {...options}>{children}</Providers>;
    return Wrapper;
};
