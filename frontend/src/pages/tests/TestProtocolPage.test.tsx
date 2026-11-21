import {beforeEach, describe, expect, it, vi} from 'vitest';
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ITestProtocolData from '../../types/test.protocol';
import {TestResultState as S} from '../../types/test.result.state.enum';
import {entry, protocol} from './run/testkit/fixtures';
import {location, login, openProtocol, shop, tab, url} from './run/testkit/protocolPage';

beforeEach(() => {
    Element.prototype.scrollIntoView = vi.fn();
});

describe('the protocol page: overview', () => {
    it('shows the test with its result, device and run', async () => {
        openProtocol(login());
        expect(await screen.findByRole('heading', {level: 1})).toHaveTextContent('Login');
        expect(screen.getByText('Game.Tests')).toBeInTheDocument();
        expect(screen.getByText('Passed', {selector: '.MuiChip-label'})).toBeInTheDocument();
        expect(screen.getByText('iPhone 12', {selector: 'a'})).toHaveAttribute('href', '/project/p1/device/7');
        expect(screen.getByText(/\(iOS 17.1\)/)).toBeInTheDocument();
        expect(screen.getByText('Run 3', {selector: 'a'})).toHaveAttribute('href', '/project/p1/app:1/test/2/run/3');
        expect(screen.getByText(/took 10s/)).toBeInTheDocument();
    });

    it('shows the steps and how the test ended', async () => {
        openProtocol(login());
        expect(await screen.findByText('open menu')).toBeInTheDocument();
        expect(screen.getByText('tap play')).toBeInTheDocument();
        expect(screen.getByText('Passed', {selector: 'p'})).toBeInTheDocument();
    });

    it('folds the detail of a message behind more and less', async () => {
        openProtocol(login());
        await userEvent.click(await screen.findByRole('button', {name: 'more (2 lines)'}));
        expect(screen.getByText(/second line/)).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'less'}));
        expect(screen.queryByText(/second line/)).not.toBeInTheDocument();
    });

    it('switches the tab and goes from the performance tiles to the charts', async () => {
        openProtocol(login());
        await userEvent.click(await screen.findByRole('button', {name: 'Charts and checkpoints'}));
        expect(tab(/Performance/)).toHaveAttribute('aria-selected', 'true');
        await userEvent.click(tab(/Log/));
        expect(location()).toBe(`${url}?t=log`);
        await userEvent.click(tab(/Overview/));
        expect(location()).toBe(url);
    });

});

describe('the protocol page: steps and averages', () => {
    it('says when the test reported no steps', async () => {
        openProtocol(login({entries: []}));
        expect(await screen.findByText('The test reported no steps.')).toBeInTheDocument();
    });

    it('compares the averages with the usual ones', async () => {
        openProtocol(login());
        expect(await screen.findByText('Frame rate (average)')).toBeInTheDocument();
        expect(screen.getByText('55.0 fps')).toBeInTheDocument();
        expect(screen.getByText('worse · usually 60.0 fps')).toBeInTheDocument();
    });
});

describe('the protocol page: the other tests of the run', () => {
    it('goes to the next test, with no previous one on the first', async () => {
        openProtocol(login());
        await screen.findByText('Test 1 of 2');
        expect(screen.getByRole('button', {name: 'Previous'})).toBeDisabled();
        await userEvent.click(screen.getByRole('button', {name: 'Next'}));
        expect(location()).toBe('/project/p1/app:1/test/2/run/3/2');
    });

    it('goes to the previous test, keeping the tab, with no next one on the last', async () => {
        openProtocol(shop, [login()], {at: '/project/p1/app:1/test/2/run/3/2?t=performance'});
        await screen.findByText('Test 2 of 2');
        expect(screen.getByRole('button', {name: 'Next'})).toBeDisabled();
        await userEvent.click(screen.getByRole('button', {name: 'Previous'}));
        expect(location()).toBe('/project/p1/app:1/test/2/run/3/1?t=performance');
    });

    it('skips to the next failure', async () => {
        openProtocol(login(), [protocol({id: 2, testName: 'Game.Tests/Other'}), {...shop, id: 3}], {at: `${url}?t=log`});
        expect(await screen.findByText('Test 1 of 3')).toBeInTheDocument();
        await userEvent.click(screen.getByText('Next failure'));
        expect(location()).toBe('/project/p1/app:1/test/2/run/3/3?t=log');
    });
});

describe('a failed test', () => {
    const errors = [1, 2, 3, 4].map(i => entry({id: 10 + i, source: 'unity', level: 'error', message: `error ${i}`, runtime: 3 + i}));
    const failed = (): ITestProtocolData => login({testResult: S.TestResultFailed, entries: [...login().entries, ...errors]});

    it('shows why it failed: the last step and the last errors', async () => {
        openProtocol(failed());
        expect(await screen.findByText('Why it failed')).toBeInTheDocument();
        expect(screen.getByText(/Last step, at 0:02.500/)).toBeInTheDocument();
        expect(screen.queryByText('error 1')).not.toBeInTheDocument();
        expect(screen.getByText('error 4')).toBeInTheDocument();
    });

    it('says an unstable test is unstable and a failure without message has none', async () => {
        openProtocol(login({testResult: S.TestResultUnstable}));
        expect(await screen.findByText('Why it is unstable')).toBeInTheDocument();
        expect(screen.getByText(/reported no error message/)).toBeInTheDocument();
    });

    it('shows an error in the log', async () => {
        openProtocol(failed());
        await userEvent.click(await screen.findAllByRole('button', {name: 'Show in log'}).then(b => b[2]));
        expect(tab(/Log/)).toHaveAttribute('aria-selected', 'true');
        expect(location()).toContain('t=log');
        await waitFor(() => expect(Element.prototype.scrollIntoView).toHaveBeenCalled());
        expect(screen.getByText('error 3')).toBeInTheDocument();
    });

    it('shows all errors in the log', async () => {
        openProtocol(failed());
        await userEvent.click(await screen.findByText('All 4 errors'));
        expect(screen.getByRole('button', {name: 'Errors only 4'})).toHaveClass('MuiChip-filled');
        expect(screen.queryByText('open menu')).not.toBeInTheDocument();
    });
});
