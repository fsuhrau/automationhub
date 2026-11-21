import {act, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {Route, Routes, useLocation} from 'react-router-dom';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import HubMainLayout from './HubMainLayout';
import ApplicationPage from '../hooks/ApplicationProvider';
import {renderInApp} from '../app/testing/renderInApp';
import {resetSse} from '../app/testing/fakeSse';
import {aDevice, aProject, anApp} from '../app/testing/seedHub';
import {fakeAuthClient, signedIn} from '../app/testing/fakeAuth';

const services = vi.hoisted(() => ({
    getProjects: vi.fn(),
    getAllApps: vi.fn(),
    getAllDevices: vi.fn(),
    getNodes: vi.fn(),
    getTest: vi.fn(),
}));

vi.mock('react-hooks-sse', async () => (await import('../app/testing/fakeSse')).fakeSseModule);
vi.mock('../project/project.service', () => ({getProjects: services.getProjects}));
vi.mock('../services/app.service', () => ({getAllApps: services.getAllApps}));
vi.mock('../services/device.service', () => ({getAllDevices: services.getAllDevices}));
vi.mock('../services/settings.service', () => ({getNodes: services.getNodes}));
vi.mock('../services/test.service', () => ({getTest: services.getTest}));

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

const open = async (path: string, auth = fakeAuthClient(signedIn)): Promise<void> => {
    renderInApp(layout(), {path, auth});
    await act(async () => undefined);
};

beforeEach(() => {
    localStorage.clear();
    resetSse();
    services.getProjects.mockResolvedValue([aProject('p1', 'Project One'), aProject('p2', 'Project Two')]);
    services.getAllApps.mockResolvedValue([anApp(1, 'Android app'), anApp(2, 'Other app')]);
    services.getAllDevices.mockResolvedValue([aDevice({id: 5})]);
    services.getNodes.mockResolvedValue([]);
    services.getTest.mockResolvedValue({name: 'Smoke test'});
});

describe('project layout', () => {
    it('shows the loader, then the menu, the header and the page of the project', async () => {
        renderInApp(layout(), {path: '/project/p1'});
        expect(screen.getByRole('progressbar', {hidden: true})).toBeInTheDocument();
        await act(async () => undefined);
        expect(screen.getByText('overview at /project/p1')).toBeInTheDocument();
        expect(within(screen.getByRole('navigation', {name: 'breadcrumb'})).getByText('Overview')).toBeInTheDocument();
        expect(screen.getByTitle('Project One')).toBeInTheDocument();
        expect(services.getProjects).toHaveBeenCalledTimes(1);
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
        await open('/project/p1/app:2/test/5');
        const crumbs = within(screen.getByRole('navigation', {name: 'breadcrumb'}));
        expect(await crumbs.findByText('Smoke test')).toBeInTheDocument();
        expect(crumbs.getByText('Project One').closest('a')).toHaveAttribute('href', '/project/p1');
        expect(crumbs.getByText('Other app').closest('a')).toHaveAttribute('href', '/project/p1/app:2/tests');
        expect(crumbs.getByText('Tests')).toBeInTheDocument();
        expect(services.getTest).toHaveBeenCalledWith('p1', 2, '5');
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
        services.getAllApps.mockResolvedValue([]);
        await open('/project/p1');
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
        services.getAllApps.mockResolvedValue([]);
        await open('/project/p1');
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
        await open('/project/p1', fakeAuthClient(signedIn, {logout}));
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
        await open('/project/p1', fakeAuthClient(signedIn, {logout}));
        await userEvent.click(screen.getByRole('button', {name: 'Open navigation', hidden: true}));
        await userEvent.click(await screen.findByRole('button', {name: 'Logout'}));
        expect(logout).toHaveBeenCalled();
    });
});
