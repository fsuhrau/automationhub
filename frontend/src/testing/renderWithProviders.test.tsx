import React from 'react';
import {render, renderHook, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {Link, Route, Routes} from 'react-router-dom';
import {describe, expect, it} from 'vitest';
import {useApiClient} from '../api/ApiClientProvider';
import {useAuth} from '../hooks/AuthProvider';
import {useHubState} from '../hooks/HubStateProvider';
import {useProjectContext} from '../hooks/ProjectProvider';
import {fakeApi} from './fakeApi';
import {fakeAuthClient, signedIn} from './fakeAuth';
import {aDevice, aProject} from './fixtures';
import {appRoute, currentLocation, Providers, projectRoute, renderAtRoot, renderWithProviders, wrapperOf} from './renderWithProviders';

const HubProbe: React.FC = () => {
    const {state} = useHubState();
    return <p>{`projects=${state.projects?.length ?? '-'} devices=${state.devices?.length ?? '-'} nodes=${state.nodes?.length ?? '-'} apps=${state.apps?.length ?? '-'}`}</p>;
};

describe('renderWithProviders', () => {
    it('gives the page the client it is given', async () => {
        const api = fakeApi({'GET /x': 'hi'});
        const {result} = renderHook(() => useApiClient(), {wrapper: wrapperOf({api})});
        expect(await result.current.get('/x')).toBe('hi');
    });

    it('shows its children when the seeded lists are in the hub state', () => {
        renderWithProviders(<HubProbe/>, {hub: {projects: [aProject()], devices: [aDevice()], nodes: [], apps: []}});
        expect(screen.getByText('projects=1 devices=1 nodes=0 apps=0')).toBeInTheDocument();
    });

    it('leaves the lists that are not seeded unloaded', () => {
        renderWithProviders(<HubProbe/>);
        expect(screen.getByText('projects=- devices=- nodes=- apps=-')).toBeInTheDocument();
    });

    it('provides the user when a session is given', async () => {
        const User: React.FC = () => <p>{useAuth().user?.name ?? 'nobody'}</p>;
        renderWithProviders(<User/>, {auth: fakeAuthClient(signedIn)});
        expect(await screen.findByText('Ada Lovelace')).toBeInTheDocument();
    });

    it('starts at the url, a string or a location', () => {
        renderWithProviders(<Routes><Route path="/a" element={<p>at a</p>}/></Routes>, {url: '/a'});
        expect(screen.getByText('at a')).toBeInTheDocument();
        renderWithProviders(<Routes><Route path="/b" element={<p>at b</p>}/></Routes>, {url: {pathname: '/b'}});
        expect(screen.getByText('at b')).toBeInTheDocument();
    });

    it('shows where the router went when asked to', async () => {
        renderWithProviders(<Routes><Route path="/" element={<Link to="/next?a=1">go</Link>}/></Routes>, {probe: true});
        expect(currentLocation()).toBe('/');
        await userEvent.click(screen.getByRole('link', {name: 'go'}));
        expect(currentLocation()).toBe('/next?a=1');
    });
});

describe('route helpers', () => {
    it('nest the project and the app routes with their providers', async () => {
        const Page: React.FC = () => <p>{useProjectContext().projectIdentifier}</p>;
        renderWithProviders(<Routes>{projectRoute(appRoute(<Route path="tests" element={<Page/>}/>))}</Routes>, {
            url: '/project/p1/app:2/tests', hub: {projects: [aProject()], apps: [{id: 2, name: 'App'} as never]},
        });
        expect(await screen.findByText('p1')).toBeInTheDocument();
    });

    it('renderAtRoot shows the page at / and the location anywhere else', () => {
        renderAtRoot(<p>root page</p>);
        expect(screen.getByText('root page')).toBeInTheDocument();
        renderAtRoot(<p>root page</p>, {url: '/elsewhere'});
        expect(currentLocation()).toBe('/elsewhere');
    });
});

describe('Providers', () => {
    it('keep the same fallback client when they render again', () => {
        const seen: unknown[] = [];
        const Spy: React.FC = () => {
            seen.push(useApiClient());
            return null;
        };
        const {rerender} = render(<Providers><Spy/></Providers>);
        rerender(<Providers><Spy/></Providers>);
        expect(new Set(seen).size).toBe(1);
    });
});
