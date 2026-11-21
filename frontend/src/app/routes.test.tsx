import {act, screen} from '@testing-library/react';
import {Outlet} from 'react-router-dom';
import {describe, expect, it, vi} from 'vitest';
import {AppRoutes} from './routes';
import {renderInApp} from './testing/renderInApp';
import {fakeAuthClient, noAuth, signedIn, signedOut} from './testing/fakeAuth';
import {routes} from '../lib/routes';

// the pages are not what is tested here: each is replaced by a stand-in with its name
vi.mock('../project/ProjectsIndexPage', async () => ({default: (await import('./testing/pageStub')).pageStub('Projects page')}));
vi.mock('../pages/profile/ProfilePage', async () => ({default: (await import('./testing/pageStub')).pageStub('Profile page')}));
vi.mock('../pages/main/ProjectDashboard', async () => ({default: (await import('./testing/pageStub')).pageStub('Overview page')}));
vi.mock('../pages/tests/TestsIndexPage', async () => ({default: (await import('./testing/pageStub')).pageStub('Tests page')}));
vi.mock('../pages/apps/AppBundlesPage', async () => ({default: (await import('./testing/pageStub')).pageStub('Bundles page')}));
vi.mock('../pages/apps/AppBundlePage', async () => ({default: (await import('./testing/pageStub')).pageStub('Bundle page')}));
vi.mock('../pages/apps/AppBundleComparePage', async () => ({default: (await import('./testing/pageStub')).pageStub('Compare page')}));
vi.mock('../pages/tests/TestPageLoader', async () => ({default: (await import('./testing/pageStub')).pageStub('Test page')}));
vi.mock('../pages/tests/EditTestPage', async () => ({default: (await import('./testing/pageStub')).pageStub('New test page')}));
vi.mock('../pages/tests/TestProtocolPageLoader', async () => ({default: (await import('./testing/pageStub')).pageStub('Protocol page')}));
vi.mock('../pages/tests/TestRunPageLoader', async () => ({default: (await import('./testing/pageStub')).pageStub('Run page')}));
vi.mock('../pages/tests/TestReplayPage', async () => ({TestReplayPageLoader: (await import('./testing/pageStub')).pageStub('Replay page')}));
vi.mock('../pages/settings/SettingsPage', async () => ({default: (await import('./testing/pageStub')).pageStub('Settings page')}));
vi.mock('../pages/devices/DevicePageLoader', async () => ({default: (await import('./testing/pageStub')).pageStub('Device page')}));
vi.mock('../pages/devices/DeviceControlPage', async () => ({default: (await import('./testing/pageStub')).pageStub('Control page')}));
vi.mock('../pages/devices/DevicesPage', async () => ({default: (await import('./testing/pageStub')).pageStub('Devices page')}));
// the frames only pass the route on; they have their own tests
vi.mock('../application/HubMainLayout', () => ({default: () => <><p>project frame</p><Outlet/></>}));
vi.mock('../hooks/ApplicationProvider', () => ({default: () => <><p>app frame</p><Outlet/></>}));

const p = routes.project('p1');
const a = p.app(3);
const t = a.test(5);

const visit = async (path: string, auth = fakeAuthClient(signedIn)): Promise<void> => {
    renderInApp(<AppRoutes/>, {path, auth});
    await act(async () => undefined);
};

describe('AppRoutes of the signed in', () => {
    it.each([
        [routes.projects(), ['Projects page']],
        [routes.profile(), ['Profile page']],
        [p.overview(), ['project frame', 'Overview page']],
        [p.settings(), ['project frame', 'Settings page']],
        [p.devices(), ['project frame', 'Devices page']],
        [p.device(4).show(), ['Device page edit=false']],
        [p.device(4).edit(), ['Device page edit=true']],
        [p.device(4).control(), ['Control page']],
        [a.tests(), ['app frame', 'Tests page']],
        [a.bundles(), ['Bundles page']],
        [a.compareBundles(), ['Compare page']],
        [a.bundle(7), ['Bundle page']],
        [a.newTest(), ['New test page']],
        [t.show(), ['Test page edit=false']],
        [t.edit(), ['Test page edit=true']],
        [t.lastRun(), ['Run page']],
        [t.run(9).show(), ['Run page']],
        [t.run(9).replay(), ['Replay page']],
        [t.run(9).protocol(2), ['Protocol page']],
    ])('renders %s', async (path, texts) => {
        await visit(path);
        texts.forEach(text => expect(screen.getByText(text)).toBeInTheDocument());
    });

    it('sends the old app home link to the project overview', async () => {
        await visit(a.home());
        expect(screen.getByText('Overview page')).toBeInTheDocument();
    });

    it.each([[a.results()], [a.performance()], [p.users()]])('renders only the frame for %s, which has no page yet', async path => {
        await visit(path);
        expect(screen.getByText('project frame')).toBeInTheDocument();
    });

    it('shows nothing for an unknown address', async () => {
        await visit('/nowhere');
        expect(screen.queryByText('Projects page')).toBeNull();
    });
});

describe('AppRoutes authentication', () => {
    it.each([[routes.projects()], [p.overview()], [t.show()]])('sends anonymous visitors from %s to the sign in', async path => {
        await visit(path, fakeAuthClient(signedOut));
        expect(screen.getByRole('heading', {name: 'Sign in'})).toBeInTheDocument();
    });

    it('shows the sign in and sign up without a session', async () => {
        await visit(routes.signup(), fakeAuthClient(signedOut));
        expect(screen.getByRole('heading', {name: 'Create an account'})).toBeInTheDocument();
    });

    it('shows the pages when the master needs no authentication', async () => {
        await visit(p.devices(), fakeAuthClient(noAuth));
        expect(screen.getByText('Devices page')).toBeInTheDocument();
    });
});
