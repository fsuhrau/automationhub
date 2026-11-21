import {describe, expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IBaseline, IBaselineCandidate} from '../../types/protocol.baseline';
import {at, performance} from './run/harness/fixtures';
import {base, login, openProtocol, url} from './run/harness/protocolPage';

const candidate: IBaselineCandidate = {
    protocolId: 40, testRunId: 1, testId: 2, startedAt: at(-500), testResult: 3, deviceId: 7, device: 'iPhone 12',
    bundle: {id: 1, name: 'Game', version: '1.0', platform: 'ios', tags: 'release'}, parameter: '',
    avgFps: 50, avgMem: 400, avgCpu: 10, avgVertexCount: 10, avgTriangles: 10,
};

const picked: IBaseline = {...candidate, performance: [performance({checkpoint: 'menu', runtime: 1, fps: 30})]};

const answers = {
    [`GET ${base}/1/baselines`]: {candidates: [candidate], tags: ['release']},
    [`GET ${base}/1/baseline/1`]: picked,
};

const openPerformance = (query = ''): void => {
    openProtocol(login(), [], {answers, at: `${url}?t=performance${query}`});
};

describe('the performance tab', () => {
    it('plots a chart per metric', async () => {
        openPerformance();
        expect(await screen.findByText('Frame rate (fps)')).toBeInTheDocument();
        expect(screen.getByText('Memory (MB)')).toBeInTheDocument();
        expect(screen.getAllByText('Checkpoints in order; hover for their names.')).toHaveLength(5);
        expect(document.querySelectorAll('svg[height="220"]')).toHaveLength(5);
    });

    it('draws the checkpoints of the test and the earlier run as reference line', async () => {
        openPerformance();
        const card = (await screen.findByText('Frame rate (fps)')).closest('.MuiCard-root') as HTMLElement;
        expect(card.querySelector('[id$="current-line-clip"]')).not.toBeNull();
        expect(card.querySelector('[id$="earlier-line-clip"]')).not.toBeNull();
        expect(within(card).getByText('40.0')).toBeInTheDocument();
    });

    it('draws no reference line without an earlier run', async () => {
        openProtocol(login({testProtocolHistory: null}), [], {at: `${url}?t=performance`});
        const card = (await screen.findByText('Frame rate (fps)')).closest('.MuiCard-root') as HTMLElement;
        expect(card.querySelector('[id$="current-line-clip"]')).not.toBeNull();
        expect(card.querySelector('[id$="earlier-line-clip"]')).toBeNull();
    });

    it('lists the checkpoints with their times against the earlier run', async () => {
        openPerformance();
        const table = await screen.findByRole('table', {name: 'Checkpoints'});
        const rows = within(table).getAllByRole('row').map(r => r.textContent);
        expect(rows[1]).toBe('menu0:02.0002.00 s1.50 s');
        expect(rows[2]).toBe('game0:05.0003.00 s4.50 s');
        expect(within(table).getByText('Run 2')).toBeInTheDocument();
    });

    it('says which run it compares with', async () => {
        openPerformance();
        expect(await screen.findByText(/the latest earlier run on this device/)).toBeInTheDocument();
        expect(screen.getByText('compared with run 2')).toBeInTheDocument();
    });
});

describe('comparing with a picked run', () => {
    it('picks a run from the list of candidates', async () => {
        openPerformance();
        await userEvent.click(await screen.findByRole('button', {name: 'Compare with…'}));
        await userEvent.click(await screen.findByText('Run 1'));
        expect(screen.getByLabelText('location')).toHaveTextContent(`${url}?t=performance&vs=1`);
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        expect(await screen.findByText(/Compared with/)).toHaveTextContent('Compared with run 1');
        expect(screen.getByText('Game 1.0')).toBeInTheDocument();
        expect(screen.getByText('release')).toBeInTheDocument();
    });

    it('closes the list of candidates without picking', async () => {
        openPerformance();
        await userEvent.click(await screen.findByRole('button', {name: 'Compare with…'}));
        await userEvent.click(await screen.findByRole('button', {name: 'Cancel'}));
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        expect(screen.getByLabelText('location')).toHaveTextContent(`${url}?t=performance`);
    });

    it('goes back to the latest earlier run', async () => {
        openPerformance('&vs=1');
        await userEvent.click(await screen.findByRole('button', {name: 'Latest earlier run'}));
        expect(screen.getByLabelText('location')).toHaveTextContent(`${url}?t=performance`);
        expect(await screen.findByText(/the latest earlier run on this device/)).toBeInTheDocument();
    });

    it('tells when the picked run could not be loaded', async () => {
        openProtocol(login(), [], {answers: {[`GET ${base}/1/baseline/1`]: new Error('gone')}, at: `${url}?t=performance&vs=1`});
        expect(await screen.findByText('The picked run could not be loaded.')).toBeInTheDocument();
    });

    it('says it loads the picked run', async () => {
        openProtocol(login(), [], {answers: {[`GET ${base}/1/baseline/1`]: () => new Promise(() => undefined)}, at: `${url}?t=performance&vs=1`});
        expect(await screen.findByText('Loading the picked run…')).toBeInTheDocument();
    });

    it('says when there is no earlier run to compare with', async () => {
        openProtocol(login({testProtocolHistory: null}), [], {at: `${url}?t=performance`});
        expect(await screen.findByText('No earlier run on this device to compare with')).toBeInTheDocument();
    });
});
