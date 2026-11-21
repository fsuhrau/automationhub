import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IBaselineCandidate} from '../../types/protocol.baseline';
import BaselinePicker from './BaselinePicker';
import {METRICS} from './run/performanceModel';
import {fakeApi, FakeApi} from '../../testing/fakeApi';
import {at} from './run/harness/fixtures';
import {wrapperOf} from '../../testing/renderWithProviders';

const refs = {projectId: 'p1', appId: 1, testId: 2, runId: 5, protocolId: 9};
const key = 'GET /p1/app/1/test/2/run/5/9/baselines';

const candidate = (over: Partial<IBaselineCandidate> = {}): IBaselineCandidate => ({
    protocolId: 40, testRunId: 3, testId: 2, startedAt: at(-500), testResult: 3, deviceId: 7, device: 'iPhone 12',
    bundle: {id: 1, name: 'Game', version: '1.0', platform: 'ios', tags: 'release,beta'}, parameter: 'A=1',
    avgFps: 50, avgMem: 0, avgCpu: 10, avgVertexCount: 10, avgTriangles: 10, ...over,
});

interface Props {
    open?: boolean,
    selectedRunId?: number | null,
}

const show = (api: FakeApi, {open = true, selectedRunId = null}: Props = {}): ReturnType<typeof vi.fn> => {
    const onPick = vi.fn();
    const Wrapper = wrapperOf({api});
    render(<Wrapper>
        <BaselinePicker open={open} onClose={() => undefined} onPick={onPick} refs={refs} deviceId={7} selectedRunId={selectedRunId} metrics={METRICS}/>
    </Wrapper>);
    return onPick;
};

describe('the baseline picker', () => {
    it('lists the runs to compare with and tells the picked one', async () => {
        const later = candidate({protocolId: 41, testRunId: 8, deviceId: 2, bundle: null, parameter: ''});
        const api = fakeApi({[key]: {candidates: [candidate(), later], tags: []}});
        const onPick = show(api);
        await userEvent.click(await screen.findByText('Run 3'));
        expect(onPick).toHaveBeenCalledWith(expect.objectContaining({testRunId: 3}));
        expect(screen.getByText('Run 8 (later)')).toBeInTheDocument();
        expect(screen.getByText('unknown bundle')).toBeInTheDocument();
        expect(screen.getByText(/\(other device\)/)).toBeInTheDocument();
    });

    it('shows the bundle, its tags and the averages of each run', async () => {
        show(fakeApi({[key]: {candidates: [candidate()], tags: []}}));
        expect(await screen.findByText('Game 1.0')).toBeInTheDocument();
        expect(screen.getByText('beta')).toBeInTheDocument();
        expect(screen.getByText('50.0 fps')).toBeInTheDocument();
        expect(screen.getAllByText('–')).toHaveLength(1);
        expect(screen.getByText(/A=1/)).toBeInTheDocument();
    });

    it('marks the run that is compared now', async () => {
        show(fakeApi({[key]: {candidates: [candidate()], tags: []}}), {selectedRunId: 3});
        expect((await screen.findByText('Run 3')).closest('[role="button"]')).toHaveClass('Mui-selected');
    });

    it('does not ask the master while closed', async () => {
        const api = fakeApi({[key]: {candidates: [], tags: []}});
        show(api, {open: false});
        await act(async () => undefined);
        expect(api.calls).toHaveLength(0);
    });
});

describe('the baseline picker: empty and failing', () => {
    it('says there is no other run', async () => {
        show(fakeApi({[key]: {candidates: [], tags: []}}));
        expect(await screen.findByText(/No other run of this test measured performance on this device. Try any device./)).toBeInTheDocument();
    });

    it('shows the error and loads again on retry', async () => {
        const api = fakeApi({[key]: new Error('down')});
        show(api);
        await userEvent.click(await screen.findByRole('button', {name: 'Retry'}));
        await waitFor(() => expect(api.calls).toHaveLength(2));
        expect(screen.getByText('The runs could not be loaded.')).toBeInTheDocument();
    });

    it('says how many are listed at most', async () => {
        const many = Array.from({length: 50}, (_, i) => candidate({protocolId: i + 100, testRunId: i + 100}));
        show(fakeApi({[key]: {candidates: many, tags: []}}));
        expect(await screen.findByText('The newest 50; search to find older runs.')).toBeInTheDocument();
    });
});

describe('the baseline picker: searching', () => {
    beforeEach(() => {
        vi.useFakeTimers({shouldAdvanceTime: true});
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    const params = (api: FakeApi): unknown[] => api.calls.map(r => r.config?.params);

    it('searches by text once typing pauses, by tag and on any device', async () => {
        const api = fakeApi({[key]: {candidates: [], tags: ['release']}});
        show(api);
        await userEvent.type(await screen.findByPlaceholderText(/Bundle name, version/), ' rel ', {advanceTimers: vi.advanceTimersByTime});
        await act(async () => {
            await vi.advanceTimersByTimeAsync(300);
        });
        await userEvent.click(screen.getByText('release'));
        await userEvent.click(screen.getByRole('switch'));
        await waitFor(() => expect(params(api).at(-1)).toEqual({q: 'rel', tag: 'release', anyDevice: 'true'}));
        expect(screen.getByText(/matching "rel"/)).toBeInTheDocument();
        await userEvent.click(screen.getByText('release'));
        await waitFor(() => expect(params(api).at(-1)).toEqual({q: 'rel', anyDevice: 'true'}));
    });
});
