import {act, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {Route, Routes, useLocation} from 'react-router-dom';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import ProjectDashboard from './ProjectDashboard';
import {ProjectProvider} from '../../hooks/ProjectProvider';
import {fakeApi, FakeApi} from '../../testing/fakeApi';
import {fakeAuthClient, signedIn} from '../../testing/fakeAuth';
import {FakeSse} from '../../testing/fakeSse';
import {aDevice, aProject, anApp} from '../../testing/fixtures';
import {renderWithProviders} from '../../testing/renderWithProviders';
import {DeviceStateType} from '../../types/deviceStateType';
import {IOverviewRun} from '../../types/overview.run';
import {IProjectOverview} from '../../types/overview.project';

const run = (id: number, fields: Partial<IOverviewRun> = {}): IOverviewRun => ({
    id, testId: 7, testName: `Test ${id}`, appId: 1, status: 'failed', state: 'finished', createdAt: new Date(Date.now() - 3 * 3600 * 1000),
    passed: 2, failed: 1, unstable: 0, open: 0, ...fields,
});

const overview = (fields: Partial<IProjectOverview> = {}): IProjectOverview =>
    ({apps: 2, tests: 4, bundles: 3, bundleStorage: 2048, failing: [run(1)], running: [], latest: [run(2, {status: 'passed'})], ...fields});

const Where: React.FC = () => <p>{`at ${useLocation().pathname}${useLocation().search}`}</p>;

const mount = async (api: FakeApi, devices = [aDevice({id: 1, status: DeviceStateType.Booted})]): Promise<void> => {
    renderWithProviders(
        <Routes>
            <Route path="/project/:project_identifier" element={<ProjectProvider><ProjectDashboard/></ProjectProvider>}/>
            <Route path="*" element={<Where/>}/>
        </Routes>,
        {url: '/project/p1', api, sse, auth: fakeAuthClient(signedIn), hub: {projects: [aProject('p1', 'Project One')], apps: [anApp(1, 'Android app')], devices}},
    );
    await act(async () => undefined);
};

let sse = new FakeSse();

beforeEach(() => {
    sse = new FakeSse();
    localStorage.clear();
});
afterEach(() => vi.useRealTimers());

describe('ProjectDashboard', () => {
    it('shows the project, its tiles and its runs', async () => {
        await mount(fakeApi({'GET /p1/overview': overview()}));
        expect(screen.getByRole('heading', {name: 'Project One', level: 1})).toBeInTheDocument();
        expect(screen.getByText('1 of 1')).toBeInTheDocument();
        expect(screen.getByText('none in use')).toBeInTheDocument();
        expect(screen.getByText('4')).toBeInTheDocument();
        expect(screen.getByText('2.00KB stored')).toBeInTheDocument();
        expect(screen.getByText('1 need attention')).toBeInTheDocument();
        expect(screen.getByText('Test 1')).toBeInTheDocument();
        expect(screen.getByText('Test 2')).toBeInTheDocument();
        expect(screen.getAllByText(/2 passed · 1 failed/)).toHaveLength(2);
    });

    it('opens a run when its row is clicked', async () => {
        await mount(fakeApi({'GET /p1/overview': overview()}));
        await userEvent.click(screen.getByText('Test 1'));
        expect(screen.getByText('at /project/p1/app:1/test/7/run/1')).toBeInTheDocument();
    });

    it('leads from the tiles to the devices and the apps', async () => {
        await mount(fakeApi({'GET /p1/overview': overview()}));
        await userEvent.click(screen.getByRole('link', {name: /Devices ready/}));
        expect(screen.getByText('at /project/p1/devices')).toBeInTheDocument();
    });

    it('says that nothing failed and that nothing ran', async () => {
        await mount(fakeApi({'GET /p1/overview': overview({failing: [], latest: []})}));
        expect(screen.getByText('No test failed in its latest run.')).toBeInTheDocument();
        expect(screen.getByText('Nothing ran yet. Start a test from Tests.')).toBeInTheDocument();
    });

});

describe('ProjectDashboard runs and devices', () => {
    it('shows the runs that are running, with the name of their app or a fallback', async () => {
        const running = [run(3, {status: 'running'}), run(4, {status: 'running', appId: 9})];
        await mount(fakeApi({'GET /p1/overview': overview({running})}));
        const section = within(screen.getByRole('heading', {name: 'Running now'}).closest('.MuiCard-root') as HTMLElement);
        expect(section.getByText(/Android app/)).toBeInTheDocument();
        expect(section.getByText(/app 9/)).toBeInTheDocument();
    });

    it('counts devices in use', async () => {
        const devices = [aDevice({id: 1, status: DeviceStateType.Booted}), aDevice({id: 2, status: DeviceStateType.Booted, lockedBy: 'someone'})];
        await mount(fakeApi({'GET /p1/overview': overview()}), devices);
        expect(screen.getByText('1 of 2')).toBeInTheDocument();
        expect(screen.getByText('1 in use')).toBeInTheDocument();
    });

    it('says when there are no devices yet', async () => {
        await mount(fakeApi({'GET /p1/overview': overview()}), []);
        expect(screen.getByText('no devices yet')).toBeInTheDocument();
    });
});

describe('ProjectDashboard for new projects', () => {
    it('asks for the first app', async () => {
        await mount(fakeApi({'GET /p1/overview': overview({apps: 0, tests: 0})}));
        await userEvent.click(screen.getByRole('link', {name: 'Add an app'}));
        expect(screen.getByText('at /project/p1/settings?tab=apps')).toBeInTheDocument();
    });

    it('offers to create a test for each app', async () => {
        await mount(fakeApi({'GET /p1/overview': overview({tests: 0})}));
        await userEvent.click(screen.getByRole('link', {name: /Test for Android app/}));
        expect(screen.getByText('at /project/p1/app:1/test/new')).toBeInTheDocument();
    });
});

describe('ProjectDashboard loading', () => {
    it('shows progress until the overview arrives', async () => {
        await mount(fakeApi({'GET /p1/overview': () => new Promise(() => undefined)}));
        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('shows the error and retries', async () => {
        let fail = true;
        const api = fakeApi({'GET /p1/overview': () => {
            if (fail) throw new Error('overview down');
            return overview();
        }});
        await mount(api);
        expect(screen.getByText('overview down')).toBeInTheDocument();
        fail = false;
        await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
        expect(await screen.findByText('Test 1')).toBeInTheDocument();
        expect(screen.queryByText('overview down')).toBeNull();
    });

    it('says when the master is too old', async () => {
        await mount(fakeApi({'GET /p1/overview': '<html/>'}));
        expect(screen.getByText(/This master is older than the page/)).toBeInTheDocument();
        expect(screen.queryByRole('progressbar')).toBeNull();
    });

});

describe('ProjectDashboard reloading', () => {
    it('reloads when a run of the project started', async () => {
        const api = fakeApi({'GET /p1/overview': overview()});
        await mount(api);
        vi.useFakeTimers();
        sse.emit('test_runs', {projectId: 'p1', appId: 1, testRunId: 5, state: 'started'});
        await act(async () => {
            await vi.advanceTimersByTimeAsync(300);
        });
        expect(api.calls).toHaveLength(2);
    });

    it('refreshes while a run is running', async () => {
        vi.useFakeTimers();
        const api = fakeApi({'GET /p1/overview': overview({running: [run(3, {status: 'running'})]})});
        await mount(api);
        await act(async () => {
            await vi.advanceTimersByTimeAsync(15000);
        });
        expect(api.calls).toHaveLength(2);
    });

    it('does not refresh when nothing runs', async () => {
        const api = fakeApi({'GET /p1/overview': overview()});
        await mount(api);
        vi.useFakeTimers();
        await act(async () => {
            await vi.advanceTimersByTimeAsync(60000);
        });
        expect(api.calls).toHaveLength(1);
    });
});
