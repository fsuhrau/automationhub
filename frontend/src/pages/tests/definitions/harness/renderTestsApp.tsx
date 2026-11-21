import React from 'react';
import {RenderResult} from '@testing-library/react';
import {Route, Routes} from 'react-router-dom';
import {ApiClient} from '../../../../api/ApiClient';
import {FakeSse} from '../../../../testing/fakeSse';
import {appRoute, projectRoute, renderWithProviders} from '../../../../testing/renderWithProviders';
import {IAppData} from '../../../../types/app';
import {UnityClientContext} from '../useUnityService';
import EditTestPage from '../../EditTestPage';
import TestPageLoader from '../../TestPageLoader';
import TestsIndexPage from '../../TestsIndexPage';
import {aProject, anApp} from './fixtures';

const Stub: React.FC<{ text: string }> = ({text}) => <p>{text}</p>;

export interface TestsAppOptions {
    api: ApiClient,
    // the unity editor's client; defaults to the same fake
    unity?: ApiClient,
    apps?: IAppData[],
    sse?: FakeSse,
}

// renderTestsApp renders the test pages at their real routes
// (/project/proj/app:1/...) with the real project and app providers
export const renderTestsApp = (path: string, {api, unity = api, apps = [anApp()], sse}: TestsAppOptions): RenderResult =>
    renderWithProviders(
        <UnityClientContext.Provider value={unity}>
            <Routes>
                {projectRoute(appRoute(<>
                    <Route path="tests" element={<TestsIndexPage/>}/>
                    <Route path="test/new" element={<EditTestPage/>}/>
                    <Route path="test/:testId" element={<TestPageLoader edit={false}/>}/>
                    <Route path="test/:testId/edit" element={<TestPageLoader edit={true}/>}/>
                    <Route path="test/:testId/run/:runId" element={<Stub text="Run page"/>}/>
                    <Route path="test/:testId/runs/last" element={<Stub text="Latest run page"/>}/>
                </>))}
            </Routes>
        </UnityClientContext.Provider>,
        {api, sse, url: path, probe: true, hub: {projects: [aProject()], apps}},
    );
