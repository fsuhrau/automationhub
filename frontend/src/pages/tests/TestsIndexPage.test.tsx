import {act, screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, describe, expect, it} from 'vitest';
import ITestData, {RunSummaryStatus} from '../../types/test';
import {createFakeApi, failure, FakeApi, Reply} from './definitions/testing/fakeApi';
import {FakeEventSource} from './definitions/testing/fakeEventSource';
import {aSummary, aTest, anApp} from './definitions/testing/fixtures';
import {renderTestsApp} from './definitions/testing/renderTestsApp';

const PATH = '/project/proj/app:1/tests';
const LIST = 'GET /proj/app/1/tests';
let api: FakeApi;

const open = (tests: Reply, apps = [anApp()], sse?: FakeEventSource): void => {
    api = createFakeApi({[LIST]: tests, 'POST /proj/app/1/test/7/run': {id: 12}});
    renderTestsApp(PATH, {api, apps, sse});
};
const location = (): string | null => screen.getByTestId('location').textContent;
const withRuns = (id: number, name: string, status: RunSummaryStatus): ITestData =>
    aTest({id, name, recentRuns: [aSummary({id: id * 10, status})]});

afterEach(() => api.restore());

describe('TestsIndexPage', () => {
    it('lists the tests with what they run and how the last run went', async () => {
        open([aTest({recentRuns: [aSummary({status: 'failed', failed: 2, passed: 1}), aSummary({id: 10})]}), aTest({id: 8, name: 'Nightly', recentRuns: []})]);
        const table = await screen.findByRole('table', {name: 'Tests'});
        expect(within(table).getByText('Smoke')).toBeInTheDocument();
        expect(within(table).getAllByText('Unity · all tests · on all devices · one device after the other')).toHaveLength(2);
        expect(within(table).getByText('Failed')).toBeInTheDocument();
        expect(within(table).getByText('Never ran')).toBeInTheDocument();
        expect(within(table).getByRole('list', {name: 'Latest runs, oldest first'})).toBeInTheDocument();
        expect(screen.getByText(/Test setups of Hero/)).toBeInTheDocument();
    });

    it('shows a running test with its progress chip', async () => {
        open([withRuns(7, 'Smoke', 'running')]);
        expect(await screen.findByText('Running', {selector: '.MuiChip-label'})).toBeInTheDocument();
    });

    it('tells the counts and the time of a run in the history tooltip', async () => {
        open([aTest({recentRuns: [aSummary({id: 70, failed: 1, unstable: 2, passed: 0})]})]);
        await userEvent.hover(await screen.findByRole('listitem', {name: 'Run 70: Passed'}));
        expect(await screen.findByRole('tooltip')).toHaveTextContent('Run 70: Passed (1 failed, 2 unstable)2026/01/02');
    });

    it('shows dashes when the master sends no history', async () => {
        open([aTest()]);
        await screen.findByText('Smoke');
        expect(screen.getAllByText('–').length).toBeGreaterThan(0);
        expect(screen.queryByRole('group', {name: /Show tests/})).not.toBeInTheDocument();
    });

});

describe('TestsIndexPage filtering', () => {
    it('filters by how the tests are doing and resets on a second click', async () => {
        open([withRuns(1, 'Alpha', 'passed'), withRuns(2, 'Beta', 'failed')]);
        await screen.findByText('Alpha');
        await userEvent.click(screen.getByRole('button', {name: 'Failing 1'}));
        expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        expect(screen.getByText('Beta')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Failing 1'}));
        expect(screen.getByText('Alpha')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Passing 1'}));
        expect(screen.queryByText('Beta')).not.toBeInTheDocument();
    });

    it('searches from six tests on and says when nothing matches', async () => {
        open(['A', 'B', 'C', 'D', 'E', 'Needle'].map((name, i) => aTest({id: i + 1, name})));
        const search = await screen.findByRole('textbox', {name: 'Find a test'});
        await userEvent.type(search, 'needle');
        expect(screen.getAllByRole('row')).toHaveLength(2);
        await userEvent.clear(search);
        await userEvent.type(search, 'zzz');
        expect(screen.getByText('No test matches.')).toBeInTheDocument();
    });

    it('has no search for a short list', async () => {
        open([aTest()]);
        await screen.findByText('Smoke');
        expect(screen.queryByRole('textbox', {name: 'Find a test'})).not.toBeInTheDocument();
    });

});

describe('TestsIndexPage empty and error states', () => {
    it('offers to create a test when there is none', async () => {
        open([]);
        expect(await screen.findByText('No tests yet')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('link', {name: 'Create a test'}));
        expect(location()).toBe('/project/proj/app:1/test/new');
    });

    it('opens the new test form from the header', async () => {
        open([]);
        await userEvent.click(await screen.findByRole('link', {name: 'New test'}));
        expect(location()).toBe('/project/proj/app:1/test/new');
    });

    it('shows the load error and loads again on retry', async () => {
        let fail = true;
        open(() => {
            if (fail) throw failure(500, 'master is down');
            return [aTest()];
        });
        expect(await screen.findByText('master is down')).toBeInTheDocument();
        fail = false;
        await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
        expect(await screen.findByText('Smoke')).toBeInTheDocument();
    });

});

describe('TestsIndexPage navigation', () => {
    it('opens a test, its last run and a run of its history', async () => {
        open([withRuns(7, 'Smoke', 'passed')]);
        await userEvent.click(await screen.findByText('Smoke'));
        expect(location()).toBe('/project/proj/app:1/test/7');
    });

    it.each([
        ['the last run chip', () => screen.getByRole('link', {name: 'Passed'})],
        ['a run in the history', () => screen.getByRole('listitem', {name: 'Run 70: Passed'})],
    ])('opens the run from %s without opening the test', async (_, target) => {
        open([withRuns(7, 'Smoke', 'passed')]);
        await screen.findByText('Smoke');
        await userEvent.click(target());
        expect(location()).toBe('/project/proj/app:1/test/7/run/70');
    });

    it('starts a test from its row and opens the new run', async () => {
        open([aTest()]);
        await userEvent.click(await screen.findByRole('button', {name: 'Run'}));
        await userEvent.click(await screen.findByRole('button', {name: 'Start run'}));
        await waitFor(() => expect(location()).toBe('/project/proj/app:1/test/7/run/12'));
    });

    it('cannot start a test of an app that is not known', async () => {
        open([aTest()], [anApp({id: 2})]);
        await screen.findByText('Smoke');
        expect(screen.getByRole('button', {name: 'Run'})).toBeDisabled();
        expect(screen.getByText(/Test setups of this app/)).toBeInTheDocument();
    });

    it('reloads when a run of the app finished', async () => {
        const sse = new FakeEventSource();
        open([aTest()], [anApp()], sse);
        await screen.findByText('Smoke');
        act(() => sse.emit('test_runs', {projectId: 'proj', appId: 1, testRunId: 3, state: 'finished'}));
        await waitFor(() => expect(api.calledWith(LIST)).toHaveLength(2));
    });
});
