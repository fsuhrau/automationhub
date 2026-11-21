import {describe, expect, it} from 'vitest';
import {fireEvent, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {TestResultState as S} from '../../types/test.result.state.enum';
import {protocol} from './run/harness/fixtures';
import {fullRun, openRun, url} from './run/harness/runPage';
import {currentLocation} from '../../testing/renderWithProviders';

const many = Array.from({length: 7}, (_, i) => protocol({
    id: i + 1, testName: `Suite.Part/test_${i}`, testResult: [S.TestResultFailed, S.TestResultUnstable][i] ?? S.TestResultSuccess,
}));

describe('the tests of a run: filters', () => {
    it('narrows the list by result', async () => {
        openRun(fullRun({protocols: many}));
        await screen.findByText('test_0');
        expect(screen.getByRole('img', {name: '1 Failed, 1 Unstable, 0 Not finished, 5 Passed'})).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Failed 1'}));
        expect(screen.queryByText('test_3')).not.toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Failed 1'}));
        expect(screen.getByText('test_3')).toBeInTheDocument();
    });

    it('narrows the list by name and clears the filter', async () => {
        openRun(fullRun({protocols: many}));
        await screen.findByText('test_0');
        await userEvent.type(screen.getByRole('textbox', {name: 'Find a test'}), 'nothing');
        expect(screen.getByText(/No tests match/)).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Show all'}));
        expect(screen.getByText('test_6')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'All 7'}));
    });

    it('opens a test with the keyboard', async () => {
        openRun(fullRun({protocols: many}));
        const row = (await screen.findByText('test_2')).closest('tr') as HTMLElement;
        row.focus();
        await userEvent.keyboard('{Enter}');
        expect(currentLocation()).toBe(`${url}/3`);
    });
});

describe('the tests of a run: rows', () => {
    it('expands a runner with several tests', async () => {
        const runner = protocol({id: 1, testName: 'Runner'});
        const second = protocol({id: 3, parentTestProtocolId: 1, testName: 'Second', device: null});
        openRun(fullRun({protocols: [runner, protocol({id: 2, parentTestProtocolId: 1, testName: 'First'}), second]}));
        const group = (await screen.findByText('Runner')).closest('tr') as HTMLElement;
        expect(group).toHaveAttribute('aria-expanded', 'false');
        expect(within(group).getByText('2 tests')).toBeInTheDocument();
        await userEvent.click(group);
        expect(screen.getByText('First')).toBeInTheDocument();
        expect(screen.getByText('Second')).toBeInTheDocument();
        await userEvent.click(group);
        expect(screen.queryByText('First')).not.toBeInTheDocument();
    });

    it('wraps long suite names after dots', async () => {
        openRun(fullRun({protocols: [protocol({id: 1, testName: 'Very.Long.Suite/test'})]}));
        expect(await screen.findByText(/Very/)).toHaveTextContent('Very.Long.Suite');
    });

    it('keeps following the log while the user scrolls', async () => {
        openRun(fullRun());
        const log = await screen.findByRole('log', {name: 'Execution log'});
        fireEvent.scroll(log);
        expect(log).toBeInTheDocument();
    });

    it('filters the errors of the execution log', async () => {
        openRun(fullRun());
        await screen.findByText('run started');
        await userEvent.click(screen.getByRole('button', {name: 'Errors only 1'}));
        expect(screen.queryByText('run started')).not.toBeInTheDocument();
        expect(screen.getAllByText('shop broke')).toHaveLength(1);
    });
});
