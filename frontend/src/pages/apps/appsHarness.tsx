// Renders the bundle pages the way the app does: the real providers and route
// shapes around them, with the API replaced by a fake that records its calls.
import {RenderResult} from '@testing-library/react';
import {Route, Routes} from 'react-router-dom';
import {ApiClient} from '../../api/ApiClient';
import {Call} from '../../testing/fakeApi';
import {appRoute, projectRoute, renderWithProviders} from '../../testing/renderWithProviders';
import {IAppData} from '../../types/app';
import AppBundleComparePage from './AppBundleComparePage';
import AppBundlePage from './AppBundlePage';
import AppBundlesPage from './AppBundlesPage';
import {androidApp, PROJECT} from './appsFixtures';
import {aProject} from '../../testing/fixtures';

// the signal a request was sent with
export const signalOf = (call: Call): AbortSignal | undefined => call.config?.signal as AbortSignal | undefined;

// renderBundles shows the page the url leads to, as /project/demo/app:3/...
export const renderBundles = (api: ApiClient, path: string, app: IAppData = androidApp): RenderResult =>
    renderWithProviders(
        <Routes>
            {projectRoute(appRoute(<>
                <Route path="bundles" element={<AppBundlesPage/>}/>
                <Route path="bundles/compare" element={<AppBundleComparePage/>}/>
                <Route path="bundle/:bundleId" element={<AppBundlePage/>}/>
            </>))}
        </Routes>,
        {api, url: `/project/${PROJECT}/app:3/${path}`, hub: {projects: [{...aProject(PROJECT, 'Demo')}], apps: [app]}},
    );
