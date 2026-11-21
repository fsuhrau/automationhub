import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {PlatformType} from '../../types/platform.type.enum';
import {TestResultState as S} from '../../types/test.result.state.enum';
import {at, device, logEntry, protocol} from './run/testkit/fixtures';
import {base, fullRun, openRun, runningRun, status} from './run/testkit/runPage';

// the live screen and device log open streams of their own
vi.mock('../../components/DeviceLiveScreen', async () => {
    const react = await import('react');
    return {default: (p: { label: string }) => react.createElement('p', null, `live screen of ${p.label}`)};
});
vi.mock('../../components/DeviceLogStream', async () => {
    const react = await import('react');
    return {default: (p: { deviceId: number }) => react.createElement('p', null, `device log of ${p.deviceId}`)};
});

const running = {state: 'running'};
const cancel = `POST ${base}/run/3/cancel`;

beforeEach(() => {
    vi.useFakeTimers({shouldAdvanceTime: true});
    vi.setSystemTime(at(5));
});

afterEach(() => {
    vi.useRealTimers();
});

describe('a run that is executing: live data', () => {
    it('counts the time it has been running', async () => {
        openRun(runningRun(), {response: running});
        expect(await screen.findByText(/running for 5s/)).toBeInTheDocument();
        await act(async () => {
            await vi.advanceTimersByTimeAsync(3000);
        });
        expect(screen.getByText(/running for 8s/)).toBeInTheDocument();
        expect(screen.getByText('0 of 1 done')).toBeInTheDocument();
    });

    it('takes new tests and results from the event stream', async () => {
        const {sse} = openRun(runningRun(), {response: running});
        await screen.findByText('Login');
        sse.emit('test_run_3_protocol', protocol({id: 2, testName: 'Game/Shop', testResult: S.TestResultOpen, endedAt: undefined}));
        sse.emit('test_run_3_protocol', protocol({id: 1, testName: 'Game/Login', testResult: S.TestResultSuccess, endedAt: at(7)}));
        expect(screen.getByText('Shop')).toBeInTheDocument();
        expect(screen.getByText('1 of 2 done')).toBeInTheDocument();
    });

    it('takes new log lines once and the end of the run from the event stream', async () => {
        const {sse} = openRun(runningRun(), {response: running});
        await screen.findByText('Login');
        const line = {testRunId: 3, entry: logEntry({id: 3, log: 'a new line'})};
        sse.emit('test_run_3_log', line);
        sse.emit('test_run_3_log', line);
        expect(screen.getAllByText('a new line')).toHaveLength(1);
        sse.emit('test_run_3_finished', {});
        await waitFor(() => expect(screen.queryByRole('button', {name: 'Cancel run'})).not.toBeInTheDocument());
        expect(screen.getByText(/took/)).toBeInTheDocument();
    });

    it('offers its recording once it finished while open', async () => {
        const recorded = runningRun();
        recorded.test.testConfig.recordSession = true;
        const {sse} = openRun(recorded, {response: {...running, devices: undefined}});
        await screen.findByText('Login');
        expect(screen.queryByText('Replay')).not.toBeInTheDocument();
        sse.emit('test_run_3_finished', {});
        expect(await screen.findByText('Replay')).toBeInTheDocument();
    });
});

describe('a run that is executing: cancelling', () => {
    it('cancels the run once', async () => {
        const {api} = openRun(runningRun(), {response: running, answers: {[cancel]: null}});
        await userEvent.click(await screen.findByRole('button', {name: 'Cancel run'}));
        expect(await screen.findByRole('button', {name: 'Cancelling…'})).toBeDisabled();
        expect(api.requests.filter(r => r.key === cancel)).toHaveLength(1);
    });

    it('shows when the run cannot be cancelled', async () => {
        openRun(runningRun(), {response: running, answers: {[cancel]: new Error('too late')}});
        await userEvent.click(await screen.findByRole('button', {name: 'Cancel run'}));
        expect(await screen.findByRole('alert')).toHaveTextContent('too late');
        expect(screen.getByRole('button', {name: 'Cancel run'})).toBeEnabled();
    });
});

describe('a run that is executing: devices', () => {
    it('shows the screen and the log of the device the run holds, and a recording marker', async () => {
        const recorded = runningRun();
        recorded.test.testConfig.recordSession = true;
        openRun(recorded, {response: {...running, devices: [7]}});
        expect(await screen.findByText('live screen of iPhone 12')).toBeInTheDocument();
        expect(screen.getByText('device log of 7')).toBeInTheDocument();
        expect(screen.getByText('Recording')).toBeInTheDocument();
        expect(screen.getByText('Open').closest('a')).toHaveAttribute('href', '/project/p1/device/7/control');
        const card = screen.getByText('Live screen').closest('.MuiCard-root') as HTMLElement;
        expect(within(card).getByText(/iPhone 12 · iOS 17.1/)).toBeInTheDocument();
    });

    it('lets the user pick the device to watch when the run holds several', async () => {
        const pixel = {...status(8, 1, 0), device: device({id: 8, name: 'Pixel 8', platformType: PlatformType.Android})};
        openRun(runningRun({deviceStatus: [status(7, 1, 0), pixel]}), {response: {...running, devices: [7, 8]}});
        const group = await screen.findByRole('group', {name: 'Device to watch'});
        await userEvent.click(within(group).getAllByRole('button')[1]);
        expect(screen.getByText('live screen of Pixel 8')).toBeInTheDocument();
        expect(screen.getByText('device log of 8')).toBeInTheDocument();
    });

    it('waits for a device, asking the master again', async () => {
        const {api} = openRun(runningRun(), {response: {...running, devices: []}});
        expect(await screen.findByText('Waiting for the run to take a device…')).toBeInTheDocument();
        api.requests.length = 0;
        await act(async () => {
            await vi.advanceTimersByTimeAsync(3000);
        });
        expect(api.requests.map(r => r.key)).toEqual([`GET ${base}/run/3`]);
    });

    it('waits for the first device when the master does not say which are held', async () => {
        const unplaced = protocol({id: 1, deviceId: null, device: null, testResult: S.TestResultOpen, endedAt: undefined});
        openRun(fullRun({deviceStatus: [], protocols: [unplaced]}), {response: {...running, devices: undefined}});
        expect(await screen.findByText('The screen shows once a device has started the app.')).toBeInTheDocument();
        expect(screen.getByText('No device has started yet.')).toBeInTheDocument();
    });
});
