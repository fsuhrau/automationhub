import {beforeEach, describe, expect, it, vi} from 'vitest';
import {screen, waitFor} from '@testing-library/react';
import {TestResultState as S} from '../../types/test.result.state.enum';
import {at, entry, protocol} from './run/testkit/fixtures';
import {base, login, openProtocol, steps} from './run/testkit/protocolPage';

beforeEach(() => {
    Element.prototype.scrollIntoView = vi.fn();
});

const running = (): ReturnType<typeof login> => login({testResult: S.TestResultOpen, endedAt: undefined, entries: steps});

describe('a protocol that is executing', () => {
    it('says it runs, and new steps arrive through the event stream once', async () => {
        const {sse} = openProtocol(running());
        expect(await screen.findByText('The test is running; the log and result update live.')).toBeInTheDocument();
        const payload = {testProtocolId: 1, entry: entry({id: 3, message: 'tap shop', runtime: 4})};
        sse.emit('test_protocol_1_log', payload);
        sse.emit('test_protocol_1_log', payload);
        expect(screen.getAllByText('tap shop')).toHaveLength(1);
        expect(screen.getByRole('tab', {name: 'Log 3'})).toBeEnabled();
    });

    it('takes its result from the event stream', async () => {
        const {sse} = openProtocol(running());
        await screen.findByText('Running', {selector: '.MuiChip-label'});
        sse.emit('test_run_3_protocol', protocol({id: 1, testName: 'Game.Tests/Login', testResult: S.TestResultFailed, endedAt: at(9)}));
        expect(await screen.findByText('Failed', {selector: '.MuiChip-label'})).toBeInTheDocument();
        expect(screen.queryByText('The test is running; the log and result update live.')).not.toBeInTheDocument();
    });

    it('shows the keys the app reports and changes them', async () => {
        const {sse} = openProtocol(running());
        await screen.findByText('Running', {selector: '.MuiChip-label'});
        const send = (value: string): void => sse.emit('test_protocol_1_keys', {testRunId: 3, testProtocolId: 1, key: {testProtocolId: 1, key: 'player', value}});
        send('1234');
        expect(await screen.findByText('1234')).toBeInTheDocument();
        send('5678');
        expect(screen.getByText('5678')).toBeInTheDocument();
        expect(screen.queryByText('1234')).not.toBeInTheDocument();
    });

    it('loads the run again when the runner that wrapped the test ends', async () => {
        const runner = protocol({id: 5, testName: 'Game.Tests/Login', endedAt: undefined});
        const {sse, api} = openProtocol(login({parentTestProtocolId: 5}), [runner]);
        await screen.findByText('open menu');
        const before = api.requests.filter(r => r.key === `GET ${base}`).length;
        sse.emit('test_run_3_protocol', protocol({id: 5, endedAt: at(30)}));
        await waitFor(() => expect(api.requests.filter(r => r.key === `GET ${base}`)).toHaveLength(before + 1));
    });
});
