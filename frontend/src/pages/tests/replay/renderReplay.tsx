// Renders the replay page like the app routes it: ApiClient fake, hub state
// with project p1, app 2 and two devices, and the real route path.
import React, {useEffect} from 'react';
import {act, fireEvent, render, screen} from '@testing-library/react';
import {MemoryRouter, Outlet, Route, Routes, useLocation} from 'react-router-dom';
import {vi} from 'vitest';
import {ApiClient} from '../../../api/ApiClient';
import {ApiClientProvider} from '../../../api/ApiClientProvider';
import {HubStateActions} from '../../../application/HubStateActions';
import ApplicationPage from '../../../hooks/ApplicationProvider';
import {HubStateProvider, useHubState} from '../../../hooks/HubStateProvider';
import {ProjectProvider} from '../../../hooks/ProjectProvider';
import IDeviceData from '../../../types/device';
import ITestRunData from '../../../types/test.run';
import {TestReplayPageLoader} from '../TestReplayPage';
import {run} from './testFixtures';

// the hub knows project p1 with app 2 and two devices
const HubSeed: React.FC<{ children: React.ReactNode }> = ({children}) => {
    const {dispatch} = useHubState();
    useEffect(() => {
        dispatch({type: HubStateActions.ProjectsUpdate, payload: [{id: 1, identifier: 'p1', name: 'P', companyId: 1, accessTokens: [], apps: []}]});
        dispatch({type: HubStateActions.AppsUpdate, payload: [{id: 2, name: 'App', identifier: 'a'} as never]});
        dispatch({type: HubStateActions.UpdateDevices, payload: [
            {id: 5, name: 'pixel', alias: 'Lab phone'} as IDeviceData, {id: 6, name: 'iphone', alias: ''} as IDeviceData,
        ]});
    }, [dispatch]);
    return <>{children}</>;
};

const Where: React.FC = () => <output data-testid="where">{useLocation().pathname + useLocation().search}</output>;

const fakeApi = (get: ApiClient['get']): ApiClient => ({get, post: vi.fn(), put: vi.fn(), delete: vi.fn()});

export const respondWith = (testRun: ITestRunData): ApiClient['get'] =>
    vi.fn().mockResolvedValue({testRun, nextRunId: 0, prevRunId: 0}) as ApiClient['get'];

export const renderReplay = (get: ApiClient['get'], search = ''): void => {
    render(<ApiClientProvider client={fakeApi(get)}>
        <HubStateProvider>
            <HubSeed>
                <MemoryRouter initialEntries={[`/project/p1/app:2/test/3/run/7/replay${search}`]}>
                    <Where/>
                    <Routes>
                        <Route path="/project/:project_identifier" element={<ProjectProvider><Outlet/></ProjectProvider>}>
                            <Route path=":appId" element={<ApplicationPage/>}>
                                <Route path="test/:testId/run/:runId/replay" element={<TestReplayPageLoader/>}/>
                                <Route path="test/:testId/run/:runId" element={<p>the run page</p>}/>
                            </Route>
                        </Route>
                    </Routes>
                </MemoryRouter>
            </HubSeed>
        </HubStateProvider>
    </ApiClientProvider>);
};

export const loaded = async (testRun: ITestRunData = run(), search = ''): Promise<ApiClient['get']> => {
    const get = respondWith(testRun);
    renderReplay(get, search);
    await screen.findByRole('heading', {level: 1});
    return get;
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
