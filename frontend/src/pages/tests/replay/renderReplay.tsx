// Renders the replay page like the app routes it: ApiClient fake, hub state
// with project p1, app 2 and two devices, and the real route path.
import {act, fireEvent, screen} from '@testing-library/react';
import {Route, Routes} from 'react-router-dom';
import {vi} from 'vitest';
import {fakeApi, FakeApi} from '../../../testing/fakeApi';
import {aProject} from '../../../testing/fixtures';
import {appRoute, projectRoute, renderWithProviders} from '../../../testing/renderWithProviders';
import {IAppData} from '../../../types/app';
import IDeviceData from '../../../types/device';
import ITestRunData from '../../../types/test.run';
import {TestReplayPageLoader} from '../TestReplayPage';
import {run} from './testFixtures';

export const RUN = 'GET /p1/app/2/test/3/run/7';

// the hub knows project p1 with app 2 and two devices
const hub = {
    projects: [aProject('p1', 'P')],
    apps: [{id: 2, name: 'App', identifier: 'a'} as IAppData],
    devices: [{id: 5, name: 'pixel', alias: 'Lab phone'} as IDeviceData, {id: 6, name: 'iphone', alias: ''} as IDeviceData],
};

export const answerWith = (testRun: ITestRunData): FakeApi => fakeApi({[RUN]: {testRun, nextRunId: 0, prevRunId: 0}});

export const renderReplay = (api: FakeApi, search = ''): void => {
    renderWithProviders(
        <Routes>
            {projectRoute(appRoute(<>
                <Route path="test/:testId/run/:runId/replay" element={<TestReplayPageLoader/>}/>
                <Route path="test/:testId/run/:runId" element={<p>the run page</p>}/>
            </>))}
        </Routes>,
        {api, url: `/project/p1/app:2/test/3/run/7/replay${search}`, hub, probe: true},
    );
};

export const loaded = async (testRun: ITestRunData = run(), search = ''): Promise<FakeApi> => {
    const api = answerWith(testRun);
    renderReplay(api, search);
    await screen.findByRole('heading', {level: 1});
    return api;
};

export const row = (text: string): HTMLElement => screen.getByText(text).closest('[role="button"]') as HTMLElement;
export const active = (): string | undefined => screen.queryAllByRole('button', {current: true})[0]?.textContent ?? undefined;
export const clockText = (): string => screen.getByText(/ \/ /).textContent as string;
export const advance = (ms: number): void => {
    act(() => {
        vi.advanceTimersByTime(ms);
    });
};
export const click = (name: string): void => {
    fireEvent.click(screen.getByRole('button', {name}));
};


export const stubMedia = (): void => {
    vi.spyOn(HTMLMediaElement.prototype, 'play').mockResolvedValue(undefined);
    vi.spyOn(HTMLMediaElement.prototype, 'pause').mockImplementation(() => undefined);
};

export const restoreAll = (): void => {
    vi.useRealTimers();
    vi.restoreAllMocks();
};
