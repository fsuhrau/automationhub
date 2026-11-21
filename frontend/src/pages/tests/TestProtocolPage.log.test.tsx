import {beforeEach, describe, expect, it, vi} from 'vitest';
import {screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ITestProtocolData from '../../types/test.protocol';
import {at, entry, protocol} from './run/testkit/fixtures';
import {login, openProtocol, shop, steps, tab, url} from './run/testkit/protocolPage';

beforeEach(() => {
    Element.prototype.scrollIntoView = vi.fn();
});

const openLog = async (main: ITestProtocolData = login(), others: ITestProtocolData[] = [shop]): Promise<void> => {
    openProtocol(main, others, {at: `${url}?t=log`});
    await screen.findByRole('log', {name: 'Protocol log'});
};

const boom = entry({id: 3, source: 'unity', level: 'error', message: 'boom', runtime: 4});

describe('the log tab: filters', () => {
    it('filters the entries by text', async () => {
        await openLog(login({entries: [...steps, boom]}));
        expect(screen.getByText(/3 entries/)).toBeInTheDocument();
        await userEvent.type(screen.getByRole('textbox', {name: 'Find in log'}), 'boom');
        expect(screen.getByText(/1 of 3 entries/)).toBeInTheDocument();
    });

    it('filters the errors and a source', async () => {
        await openLog(login({entries: [...steps, boom]}));
        await userEvent.click(screen.getByRole('button', {name: 'Errors only 1'}));
        expect(screen.queryByText('open menu')).not.toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Errors only 1'}));
        await userEvent.click(screen.getByText('step 2'));
        expect(screen.queryByText('open menu')).not.toBeInTheDocument();
        expect(screen.getByText('boom')).toBeInTheDocument();
    });

    it('offers to show everything when nothing matches', async () => {
        await openLog();
        await userEvent.type(screen.getByRole('textbox', {name: 'Find in log'}), 'zzz');
        expect(screen.getByText(/Nothing matches/)).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Show everything'}));
        expect(screen.getByText('open menu')).toBeInTheDocument();
    });

    it('is disabled without entries', async () => {
        openProtocol(login({entries: []}));
        await screen.findByRole('tab', {name: 'Log 0'});
        expect(tab(/Log/)).toBeDisabled();
    });
});

describe('the log tab: the log', () => {
    const runner = protocol({
        id: 5, testName: 'Game.Tests/Login', startedAt: at(-2), endedAt: at(20),
        entries: [entry({id: 50, source: '', message: 'app started', runtime: 0}), entry({id: 51, source: '', message: 'app closed', runtime: 15})],
    });

    it('merges the log of the runner as setup and teardown, which can be switched off', async () => {
        await openLog(login({parentTestProtocolId: 5, startedAt: at(0)}), [runner]);
        expect(await screen.findByText('app started')).toBeInTheDocument();
        expect(screen.getByText('app closed')).toBeInTheDocument();
        expect(screen.getByText(/setup before 0:00/)).toBeInTheDocument();
        await userEvent.click(screen.getByText('setup 1'));
        expect(screen.queryByText('app started')).not.toBeInTheDocument();
        expect(screen.getByText('app closed')).toBeInTheDocument();
    });

    it('shows a page of a long log and more on request', async () => {
        const long = Array.from({length: 520}, (_, i) => entry({id: 100 + i, message: `line ${i}`, runtime: i}));
        await openLog(login({entries: long}));
        expect(screen.queryByText('line 519')).not.toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Show 20 more of 20'}));
        expect(screen.getByText('line 519')).toBeInTheDocument();
    });

    it('downloads the log as a text file', async () => {
        URL.createObjectURL = vi.fn(() => 'blob:log');
        URL.revokeObjectURL = vi.fn();
        const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);
        await openLog();
        await userEvent.click(screen.getByRole('button', {name: 'Download'}));
        expect(click).toHaveBeenCalled();
        expect(URL.createObjectURL).toHaveBeenCalledTimes(1);
        click.mockRestore();
    });
});
