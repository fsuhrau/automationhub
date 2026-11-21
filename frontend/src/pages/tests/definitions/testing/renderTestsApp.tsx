import React, {useEffect} from 'react';
import {render, RenderResult} from '@testing-library/react';
import {MemoryRouter, Outlet, Route, Routes, useLocation} from 'react-router-dom';
import {SSEProvider} from 'react-hooks-sse';
import {ApiClient} from '../../../../api/ApiClient';
import {ApiClientProvider} from '../../../../api/ApiClientProvider';
import {HubStateActions} from '../../../../application/HubState';
import {ErrorProvider} from '../../../../ErrorProvider';
import ApplicationPage from '../../../../hooks/ApplicationProvider';
import {HubStateProvider, useHubState} from '../../../../hooks/HubStateProvider';
import {ProjectProvider} from '../../../../hooks/ProjectProvider';
import {IAppData} from '../../../../types/app';
import {UnityClientContext} from '../useUnityService';
import EditTestPage from '../../EditTestPage';
import TestPageLoader from '../../TestPageLoader';
import TestsIndexPage from '../../TestsIndexPage';
import {FakeEventSource} from './fakeEventSource';
import {aProject, anApp} from './fixtures';

const Seed: React.FC<{ apps: IAppData[] }> = ({apps}) => {
    const {dispatch} = useHubState();
    useEffect(() => {
        dispatch({type: HubStateActions.ProjectsUpdate, payload: [aProject()]});
        dispatch({type: HubStateActions.AppsUpdate, payload: apps});
    }, [dispatch, apps]);
    return null;
};

const Where: React.FC = () => <p data-testid="location">{useLocation().pathname}</p>;

const Stub: React.FC<{ text: string }> = ({text}) => <p>{text}</p>;

export interface TestsAppOptions {
    api: { client: ApiClient },
    // the unity editor's client; defaults to the same fake
    unity?: ApiClient,
    apps?: IAppData[],
    sse?: FakeEventSource,
}

// renderTestsApp renders the test pages at their real routes
// (/project/proj/app:1/...) with the real project and app providers
export const renderTestsApp = (path: string, options: TestsAppOptions): RenderResult => {
    const {api, unity = api.client, apps = [anApp()], sse = new FakeEventSource()} = options;
    return render(
        <ApiClientProvider client={api.client}>
            <UnityClientContext.Provider value={unity}>
                <SSEProvider source={() => sse as unknown as EventSource}>
                    <ErrorProvider>
                        <HubStateProvider>
                            <Seed apps={apps}/>
                            <MemoryRouter initialEntries={[path]}>
                                <Where/>
                                <Routes>
                                    <Route path="/project/:project_identifier" element={<ProjectProvider><Outlet/></ProjectProvider>}>
                                        <Route path=":appId" element={<ApplicationPage/>}>
                                            <Route path="tests" element={<TestsIndexPage/>}/>
                                            <Route path="test/new" element={<EditTestPage/>}/>
                                            <Route path="test/:testId" element={<TestPageLoader edit={false}/>}/>
                                            <Route path="test/:testId/edit" element={<TestPageLoader edit={true}/>}/>
                                            <Route path="test/:testId/run/:runId" element={<Stub text="Run page"/>}/>
                                            <Route path="test/:testId/runs/last" element={<Stub text="Latest run page"/>}/>
                                        </Route>
                                    </Route>
                                </Routes>
                            </MemoryRouter>
                        </HubStateProvider>
                    </ErrorProvider>
                </SSEProvider>
            </UnityClientContext.Provider>
        </ApiClientProvider>,
    );
};
