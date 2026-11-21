import {beforeEach, describe, expect, it, vi} from 'vitest';
import {fireEvent, screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {entry, protocol} from './run/testkit/fixtures';
import {login, openProtocol, shop, steps, url} from './run/testkit/protocolPage';

beforeEach(() => {
    Element.prototype.scrollIntoView = vi.fn();
});

const shots = [
    entry({id: 30, source: 'screen', data: 'one.png', message: 'menu', runtime: 1}),
    entry({id: 31, source: 'screen', data: 'two.png', message: '', runtime: 2}),
];

const openAt = (tabName: string): void => {
    openProtocol(login({entries: shots}), [shop], {at: `${url}?t=${tabName}`});
};

describe('screenshots', () => {
    it('shows the last screenshot on the overview, and all of them on their tab', async () => {
        openProtocol(login({entries: [...steps, ...shots]}));
        expect(await screen.findByText('Last screenshot')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'All 2'}));
        expect(screen.getByAltText('menu')).toHaveAttribute('src', '/api/data/one.png');
        expect(screen.getByAltText('Screenshot at 0:02.000')).toBeInTheDocument();
        expect(screen.getByText('menu', {selector: 'p'})).toBeInTheDocument();
    });

    it('says when the file of a screenshot is gone', async () => {
        openAt('screenshots');
        fireEvent.error(await screen.findByAltText('menu'));
        expect(screen.getByText('The file is no longer on the master')).toBeInTheDocument();
        expect(screen.queryByRole('button', {name: 'Open screenshot at 0:01.000'})).not.toBeInTheDocument();
    });

    it('opens the last screenshot from the overview', async () => {
        openAt('');
        await userEvent.click(await screen.findByRole('button', {name: 'Open screenshot at 0:02.000'}));
        expect(within(await screen.findByRole('dialog')).getByText(/2 of 2/)).toBeInTheDocument();
    });

    it('opens a screenshot from the log', async () => {
        openAt('log');
        const log = await screen.findByRole('log', {name: 'Protocol log'});
        await userEvent.click(within(log).getAllByRole('button', {name: /Open screenshot/})[1]);
        expect(within(await screen.findByRole('dialog')).getByText(/2 of 2/)).toBeInTheDocument();
    });
});

describe('the screenshot viewer', () => {
    it('steps through the screenshots and closes', async () => {
        openAt('screenshots');
        await userEvent.click(await screen.findByRole('button', {name: 'Open screenshot at 0:01.000'}));
        const viewer = await screen.findByRole('dialog');
        expect(within(viewer).getByText(/1 of 2 · at 0:01.000/)).toBeInTheDocument();
        expect(within(viewer).getByRole('button', {name: 'Previous screenshot'})).toBeDisabled();
        await userEvent.click(within(viewer).getByRole('button', {name: 'Next screenshot'}));
        expect(within(viewer).getByText(/2 of 2/)).toBeInTheDocument();
        expect(within(viewer).getByRole('button', {name: 'Next screenshot'})).toBeDisabled();
        await userEvent.click(within(viewer).getByRole('button', {name: 'Previous screenshot'}));
        await userEvent.click(within(viewer).getByRole('button', {name: 'Close'}));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('closes with escape', async () => {
        openAt('screenshots');
        await userEvent.click(await screen.findByRole('button', {name: 'Open screenshot at 0:01.000'}));
        await screen.findByRole('dialog');
        await userEvent.keyboard('{Escape}');
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });
});

describe('keys', () => {
    const key = (k: string, value: string, id: number): { testProtocolId: number, key: string, value: string } => ({testProtocolId: id, key: k, value});

    it('shows the keys of the test and of its runner, the test winning', async () => {
        const runner = protocol({id: 5, testName: 'Game.Tests/Login', keys: [key('player', 'runner', 5), key('level', '3', 5)]});
        openProtocol(login({parentTestProtocolId: 5, keys: [key('player', 'mine', 1), key('empty', '', 1)]}), [runner]);
        const keys = (await screen.findByText('Keys')).closest('.MuiCard-root') as HTMLElement;
        expect(within(keys).getByText('mine')).toBeInTheDocument();
        expect(within(keys).queryByText('runner')).not.toBeInTheDocument();
        expect(within(keys).getByText('3')).toBeInTheDocument();
        expect(within(keys).getByText('(empty)')).toBeInTheDocument();
    });
});
