import {fireEvent, screen, within} from '@testing-library/react';
import {afterEach, beforeEach, expect, it} from 'vitest';
import {fakeApi} from '../../testing/fakeApi';
import ITestRunData from '../../types/test.run';
import {clockText, loaded, renderReplay, restoreAll, RUN, stubMedia} from './replay/renderReplay';
import {logEntry, protocol, recording, run, T0} from './replay/testFixtures';

beforeEach(stubMedia);
afterEach(restoreAll);

it('starts at the time of the url', async () => {
    await loaded(run(), `?t=${T0 + 4000}`);
    expect(clockText()).toBe('0:04.0 / 0:10.0');
});

it('filters the log to tests and errors', async () => {
    await loaded();
    fireEvent.click(screen.getByRole('button', {name: 'Errors'}));
    expect(screen.queryByText('boot')).not.toBeInTheDocument();
    expect(screen.getByText('crash')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', {name: 'Tests and steps'}));
    expect(screen.queryByText('boot')).not.toBeInTheDocument();
    expect(screen.getByText('tap')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', {name: 'Everything'}));
    expect(screen.getByText('boot')).toBeInTheDocument();
});

it('says when a filter shows nothing', async () => {
    await loaded(run({log: [], protocols: [protocol({testResult: 3, entries: []})]}));
    fireEvent.click(screen.getByRole('button', {name: 'Errors'}));
    expect(screen.getByText('Nothing to show.')).toBeInTheDocument();
});

it('stops following the video when the list is scrolled by hand', async () => {
    await loaded();
    const follow = screen.getByRole('button', {name: 'Follow the video'});
    expect(follow).toHaveClass('MuiChip-filled');
    fireEvent.wheel(screen.getByRole('list'));
    expect(follow).toHaveClass('MuiChip-outlined');
    fireEvent.click(follow);
    expect(follow).toHaveClass('MuiChip-filled');
    fireEvent.click(follow);
    expect(follow).toHaveClass('MuiChip-outlined');
});

const two = (): ITestRunData => run({protocols: [protocol(), protocol({id: 12, deviceId: 6, testName: 'Other'})]});

it('devices: switches the device and keeps the time in the url', async () => {
    await loaded(two());
    const group = screen.getByRole('group', {name: 'Device'});
    expect(within(group).getByRole('button', {name: 'Lab phone'})).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', {name: 'Ahead 5 seconds'}));
    fireEvent.click(within(group).getByRole('button', {name: 'iphone'}));
    expect(screen.getByTestId('location')).toHaveTextContent(`device=6&t=${T0 + 5000}`);
    expect(screen.getByText('This device has no recording.', {selector: 'p'})).toBeInTheDocument();
    expect(screen.getByText(/Turn on "Record the session"/)).toBeInTheDocument();
});

it('devices: picks the device of the url', async () => {
    await loaded(two(), '?device=6');
    expect(screen.getByText('Test started: Other')).toBeInTheDocument();
    expect(screen.queryByText(/Login/)).not.toBeInTheDocument();
});

it('devices: shows no device chips for a single device', async () => {
    await loaded();
    expect(screen.queryByRole('group', {name: 'Device'})).not.toBeInTheDocument();
});

it('devices: names a device the hub does not know by its id', async () => {
    await loaded(run({recordings: [recording({deviceId: 99})], protocols: []}));
    expect(screen.getByLabelText('Screen of Device 99')).toBeInTheDocument();
});

it('empty: says so', async () => {
    await loaded(run({recordings: [], protocols: [], log: []}));
    expect(screen.getByText('This run has nothing to replay yet.')).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: 'Play'})).not.toBeInTheDocument();
});

it('empty: replays the log alone when no device exists', async () => {
    await loaded(run({recordings: [], protocols: [], log: [logEntry({id: 1, log: 'only', time: 0})]}));
    expect(screen.getByText('only')).toBeInTheDocument();
    expect(screen.getByText('This device has no recording.', {selector: 'p'})).toBeInTheDocument();
});

it('error: shows the error and retries', async () => {
    let asked = 0;
    const api = fakeApi({[RUN]: () => {
        if (asked++ === 0) throw new Error('boom');
        return {testRun: run()};
    }});
    renderReplay(api);
    expect(await screen.findByText('boom')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', {name: 'Retry'}));
    expect(await screen.findByRole('heading', {level: 1})).toBeInTheDocument();
    expect(api.called(RUN)).toHaveLength(2);
});

