import {describe, expect, it} from 'vitest';
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {TestResultState as S} from '../../types/test.result.state.enum';
import {fakeApi} from '../../testing/fakeApi';
import {FakeSse} from '../../testing/fakeSse';
import {protocol, run} from './run/harness/fixtures';
import {renderAt} from './run/harness/render';
import {base, fullRun, openRun, url} from './run/harness/runPage';
import {currentLocation} from '../../testing/renderWithProviders';

describe('the test run page: the run', () => {
    it('shows the run, its status and its devices', async () => {
        openRun(fullRun());
        expect(await screen.findByRole('heading', {level: 1})).toHaveTextContent('Smoke · Run 3');
        expect(screen.getByText('Failed', {selector: '.MuiChip-label'})).toBeInTheDocument();
        expect(screen.getByText(/took 10s · 1 device · Game 1.4/)).toBeInTheDocument();
        expect(screen.getByText('Login')).toBeInTheDocument();
        expect(screen.getByText('Shop')).toBeInTheDocument();
        expect(screen.getByText('iPhone 12', {selector: 'a'})).toHaveAttribute('href', '/project/p1/device/7');
        expect(screen.getByText('3.0 s')).toBeInTheDocument();
        expect(screen.getByText(/usually 2.0 s/)).toBeInTheDocument();
    });

    it('shows the parameters, the bundle and the log', async () => {
        openRun(fullRun());
        expect(await screen.findByText('game://start')).toBeInTheDocument();
        expect(screen.getByText('(empty)')).toBeInTheDocument();
        expect(screen.getByText('abc123')).toBeInTheDocument();
        expect(screen.getByText('addon')).toBeInTheDocument();
        expect(screen.getByRole('link', {name: 'All bundles'})).toHaveAttribute('href', '/project/p1/app:1/bundles');
        expect(screen.getByText('run started')).toBeInTheDocument();
    });

});

describe('the test run page: navigation', () => {
    it('shows the last run on the last run url', async () => {
        const last = {testRun: fullRun(), nextRunId: null, prevRunId: null, state: 'finished'};
        openRun(fullRun(), {answers: {[`GET ${base}/runs/last`]: last}, at: '/project/p1/app:1/test/2/runs/last'});
        expect(await screen.findByRole('heading', {level: 1})).toHaveTextContent('Run 3');
        expect(screen.queryByRole('navigation', {name: 'Runs of this test'})).not.toBeInTheDocument();
    });

    it('moves to the run before', async () => {
        openRun(fullRun(), {answers: {[`GET ${base}/run/2`]: {testRun: fullRun({id: 2}), nextRunId: 3, prevRunId: null}}});
        await userEvent.click(await screen.findByText('Run 2'));
        expect(currentLocation()).toBe('/project/p1/app:1/test/2/run/2');
        expect(await screen.findByRole('heading', {level: 1})).toHaveTextContent('Run 2');
    });

    it('moves to the run after', async () => {
        openRun(fullRun(), {answers: {[`GET ${base}/run/4`]: {testRun: fullRun({id: 4}), nextRunId: null, prevRunId: 3}}});
        await userEvent.click(await screen.findByText('Run 4'));
        expect(await screen.findByRole('heading', {level: 1})).toHaveTextContent('Run 4');
        expect(screen.getByText('Latest run')).toBeInTheDocument();
    });

    it('opens the protocol of a test row', async () => {
        openRun(fullRun());
        await userEvent.click(await screen.findByText('Shop'));
        expect(currentLocation()).toBe(`${url}/2`);
    });
});

describe('the test run page: actions', () => {
    it('has no replay without recordings', async () => {
        openRun(fullRun());
        await screen.findByText('Shop');
        expect(screen.queryByText('Replay')).not.toBeInTheDocument();
        expect(screen.queryByRole('button', {name: /Replay/})).not.toBeInTheDocument();
    });

    it('offers the replay of a recorded run: of the whole run and from a test', async () => {
        openRun(fullRun({recordings: [{id: 1} as never]}));
        expect((await screen.findByText('Replay')).closest('a')).toHaveAttribute('href', `${url}/replay`);
        await userEvent.click(screen.getByRole('button', {name: 'Replay Shop'}));
        expect(currentLocation()).toBe(`${url}/replay?t=${new Date(protocol().startedAt).valueOf()}&device=7`);
    });

    it('starts the test again and opens the new run', async () => {
        const {api} = openRun(fullRun(), {answers: {[`POST ${base}/run`]: run({id: 9})}});
        await userEvent.click(await screen.findByText('Run again'));
        await waitFor(() => expect(currentLocation()).toBe('/project/p1/app:1/test/2/run/9'));
        expect(api.calls.find(r => r.key === `POST ${base}/run`)?.data).toEqual({appBinaryId: 5, params: 'A=1;B=', startUrl: 'game://start'});
    });
});

describe('the test run page: runs without results', () => {
    it('shows why a run did not start', async () => {
        openRun(fullRun({protocols: []}));
        expect(await screen.findByText('No test ran. The run stopped with:')).toBeInTheDocument();
        expect(screen.getAllByText('shop broke')).toHaveLength(2);
        expect(screen.getByText('Did not run')).toBeInTheDocument();
    });

    it.each([
        {state: 'running', text: 'Waiting for the first test to start…'},
        {state: 'finished', text: 'This run has no test results.'},
    ])('says so for a $state run without results or errors', async ({state, text}) => {
        const empty = fullRun({protocols: [], log: [], appBinary: null, parameter: '', startUrl: '', deviceStatus: []});
        openRun(empty, {response: {state, devices: undefined}});
        expect(await screen.findByText(text)).toBeInTheDocument();
        expect(screen.getByText('The run had no parameters.')).toBeInTheDocument();
        expect(screen.getByText(/No bundle was installed/)).toBeInTheDocument();
        expect(screen.getByText('No log yet.')).toBeInTheDocument();
    });

    it('explains an interrupted run', async () => {
        openRun(fullRun({protocols: [protocol({id: 1, testResult: S.TestResultOpen, endedAt: undefined})]}));
        expect(await screen.findByText(/Some tests never finished/)).toBeInTheDocument();
        expect(screen.getByText('Interrupted')).toBeInTheDocument();
    });
});

describe('the test run page: loading', () => {
    it('shows the error of a run that cannot be loaded and loads it again on retry', async () => {
        const api = fakeApi({[`GET ${base}/run/3`]: new Error('master is down')});
        renderAt(url, {api, sse: new FakeSse()});
        expect(await screen.findByText('master is down')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
        await waitFor(() => expect(api.calls).toHaveLength(2));
    });
});
