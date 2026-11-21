import {act, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {Route, Routes, useLocation} from 'react-router-dom';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import HubMainLayout from './HubMainLayout';
import ApplicationPage from '../hooks/ApplicationProvider';
import {fakeApi, FakeApi, Reply} from '../testing/fakeApi';
import {fakeAuthClient, signedIn} from '../testing/fakeAuth';
import {FakeSse, stubEventSource} from '../testing/fakeSse';
import {aDevice, aProject, anApp} from '../testing/fixtures';
import {renderWithProviders} from '../testing/renderWithProviders';
import {AuthClient} from '../app/authClient';

const masterReplies = (): Record<string, Reply> => ({
    'GET /projects': [aProject('p1', 'Project One'), aProject('p2', 'Project Two')],
    'GET /p1/apps': [anApp(1, 'Android app'), anApp(2, 'Other app')],
    'GET /p1/devices': [aDevice({id: 5})],
    'GET /p1/settings/nodes': [],
    'GET /p1/app/2/test/5': {name: 'Smoke test'},
});

const Where: React.FC<{ name: string }> = ({name}) => <p>{`${name} at ${useLocation().pathname}${useLocation().search}`}</p>;

const layout = (): React.ReactElement => <Routes>
    <Route path="/" element={<Where name="projects"/>}/>
    <Route path="/profile" element={<Where name="profile"/>}/>
    <Route path="/login" element={<Where name="login"/>}/>
    <Route path="project" element={<HubMainLayout/>}>
        <Route path=":project_identifier" element={<Where name="overview"/>}/>
        <Route path=":project_identifier/settings" element={<Where name="settings"/>}/>
        <Route path=":project_identifier/:appId" element={<ApplicationPage/>}>
            <Route path="tests" element={<Where name="tests"/>}/>
            <Route path="bundles" element={<Where name="bundles"/>}/>
            <Route path="test/:testId" element={<Where name="test"/>}/>
        </Route>
    </Route>
</Routes>;

interface OpenOptions {
    auth?: AuthClient,
    // answers of the master that differ from masterReplies
    replies?: Record<string, Reply>,
}

const mount = (path: string, {auth = fakeAuthClient(signedIn), replies = {}}: OpenOptions = {}): FakeApi => {
    const api = fakeApi({...masterReplies(), ...replies});
    renderWithProviders(layout(), {url: path, auth, api, sse});
    return api;
};

const open = async (path: string, options?: OpenOptions): Promise<FakeApi> => {
    const api = mount(path, options);
    await act(async () => undefined);
    return api;
};

const sse = new FakeSse();

beforeEach(() => {
    localStorage.clear();
    stubEventSource(sse);
});
afterEach(() => vi.unstubAllGlobals());

describe('project layout', () => {
    it('shows the loader, then the menu, the header and the page of the project', async () => {
        const api = mount('/project/p1');
        expect(screen.getByRole('progressbar', {hidden: true})).toBeInTheDocument();
        await act(async () => undefined);
        expect(screen.getByText('overview at /project/p1')).toBeInTheDocument();
        expect(within(screen.getByRole('navigation', {name: 'breadcrumb'})).getByText('Overview')).toBeInTheDocument();
        expect(screen.getByTitle('Project One')).toBeInTheDocument();
        expect(api.called('GET /projects')).toHaveLength(1);
    });

    it('sends an unknown project to the project list', async () => {
        await open('/project/zzz');
        expect(screen.getByText('projects at /')).toBeInTheDocument();
    });

    it('marks the menu entry of the current page', async () => {
        await open('/project/p1/settings');
        expect(screen.getByRole('button', {name: /Settings/})).toHaveAttribute('aria-current', 'page');
        expect(screen.getByRole('button', {name: /Overview/})).not.toHaveAttribute('aria-current');
    });

    it('opens other pages from the menu', async () => {
        await open('/project/p1');
        await userEvent.click(screen.getByRole('button', {name: /Settings/}));
        expect(screen.getByText('settings at /project/p1/settings')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: /Bundles/}));
        expect(screen.getByText('bundles at /project/p1/app:1/bundles')).toBeInTheDocument();
    });

    it('shows breadcrumbs for nested routes, with the name of the test', async () => {
        const api = await open('/project/p1/app:2/test/5');
        const crumbs = within(screen.getByRole('navigation', {name: 'breadcrumb'}));
        expect(await crumbs.findByText('Smoke test')).toBeInTheDocument();
        expect(crumbs.getByText('Project One').closest('a')).toHaveAttribute('href', '/project/p1');
        expect(crumbs.getByText('Other app').closest('a')).toHaveAttribute('href', '/project/p1/app:2/tests');
        expect(crumbs.getByText('Tests')).toBeInTheDocument();
        expect(api.called('GET /p1/app/2/test/5')).toHaveLength(1);
    });
});

describe('project and app selection', () => {
    it('opens the project list to switch the project', async () => {
        await open('/project/p1');
        await userEvent.click(screen.getByRole('link', {name: 'Switch project'}));
        expect(screen.getByText('projects at /')).toBeInTheDocument();
    });

    it('opens the tests of the app that is picked', async () => {
        await open('/project/p1');
        await userEvent.click(screen.getByRole('combobox', {name: /App/}));
        await userEvent.click(await screen.findByRole('option', {name: /Other app/}));
        expect(screen.getByText('tests at /project/p1/app:2/tests')).toBeInTheDocument();
        expect(localStorage.getItem('appId')).toBe('2');
    });

    it('stays on bundles when another app is picked there', async () => {
        await open('/project/p1/app:1/bundles');
        await userEvent.click(screen.getByRole('combobox', {name: /App/}));
        await userEvent.click(await screen.findByRole('option', {name: /Other app/}));
        expect(screen.getByText('bundles at /project/p1/app:2/bundles')).toBeInTheDocument();
    });

    it('opens the apps settings to add an app', async () => {
        await open('/project/p1');
        await userEvent.click(screen.getByRole('combobox', {name: /App/}));
        await userEvent.click(await screen.findByRole('option', {name: /Add app/}));
        expect(screen.getByText('settings at /project/p1/settings?tab=apps')).toBeInTheDocument();
    });

    it('points to the settings for a project without apps', async () => {
        await open('/project/p1', {replies: {'GET /p1/apps': []}});
        expect(screen.getByText(/This project has no apps yet/)).toBeInTheDocument();
        await userEvent.click(screen.getByRole('link', {name: 'Add one in Settings'}));
        expect(screen.getByText('settings at /project/p1/settings?tab=apps')).toBeInTheDocument();
    });
});

describe('menu collapse', () => {
    it('collapses to a rail of icons and remembers it', async () => {
        await open('/project/p1');
        await userEvent.click(screen.getByRole('button', {name: 'Collapse menu'}));
        expect(screen.getByRole('button', {name: 'Expand menu'})).toBeInTheDocument();
        expect(screen.queryByTitle('Project One')).toBeNull();
        expect(localStorage.getItem('mainMenuCollapsed')).toBe('true');
    });

    it('starts collapsed when it was left that way, and expands from the app entry', async () => {
        localStorage.setItem('mainMenuCollapsed', 'true');
        await open('/project/p1');
        await userEvent.click(screen.getByRole('button', {name: /App Android app: open the menu to switch/}));
        expect(screen.getByRole('button', {name: 'Collapse menu'})).toBeInTheDocument();
        expect(localStorage.getItem('mainMenuCollapsed')).toBe('false');
    });

    it('opens pages from the rail, and tells when there is no app to pick', async () => {
        localStorage.setItem('mainMenuCollapsed', 'true');
        await open('/project/p1', {replies: {'GET /p1/apps': []}});
        expect(screen.getByRole('button', {name: 'Pick an app'})).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Settings'}));
        expect(screen.getByText('settings at /project/p1/settings')).toBeInTheDocument();
    });
});

describe('account and small screens', () => {
    it('shows who is signed in and opens the profile from the account menu', async () => {
        await open('/project/p1');
        expect(screen.getByText('ada@example.com')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Account menu'}));
        await userEvent.click(await screen.findByRole('menuitem', {name: 'Profile and access tokens'}));
        expect(screen.getByText('profile at /profile')).toBeInTheDocument();
    });

    it('opens all projects and logs out from the account menu', async () => {
        const logout = vi.fn(async () => true);
        await open('/project/p1', {auth: fakeAuthClient(signedIn, {logout})});
        await userEvent.click(screen.getByRole('button', {name: 'Account menu'}));
        await userEvent.click(await screen.findByRole('menuitem', {name: 'Logout'}));
        expect(logout).toHaveBeenCalled();
    });

    it('opens all projects from the account menu', async () => {
        await open('/project/p1');
        await userEvent.click(screen.getByRole('button', {name: 'Account menu'}));
        await userEvent.click(await screen.findByRole('menuitem', {name: 'All projects'}));
        expect(screen.getByText('projects at /')).toBeInTheDocument();
    });

    it('opens the menu as a drawer, closes it on navigation, and opens the profile', async () => {
        await open('/project/p1');
        await userEvent.click(screen.getByRole('button', {name: 'Open navigation', hidden: true}));
        const drawer = within(await screen.findByRole('presentation'));
        await userEvent.click(drawer.getByRole('button', {name: /Settings/}));
        expect(screen.getByText('settings at /project/p1/settings')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Open navigation', hidden: true}));
        await userEvent.click(await screen.findByRole('button', {name: 'Profile'}));
        expect(screen.getByText('profile at /profile')).toBeInTheDocument();
    });

    it('logs out from the drawer', async () => {
        const logout = vi.fn(async () => true);
        await open('/project/p1', {auth: fakeAuthClient(signedIn, {logout})});
        await userEvent.click(screen.getByRole('button', {name: 'Open navigation', hidden: true}));
        await userEvent.click(await screen.findByRole('button', {name: 'Logout'}));
        expect(logout).toHaveBeenCalled();
    });
});
