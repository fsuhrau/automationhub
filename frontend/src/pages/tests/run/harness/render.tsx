import {RenderResult} from '@testing-library/react';
import {Route, Routes} from 'react-router-dom';
import {FakeApi} from '../../../../testing/fakeApi';
import {FakeSse} from '../../../../testing/fakeSse';
import {aProject} from '../../../../testing/fixtures';
import {appRoute, projectRoute, renderWithProviders} from '../../../../testing/renderWithProviders';
import {IAppData} from '../../../../types/app';
import IDeviceData from '../../../../types/device';
import TestProtocolPageLoader from '../../TestProtocolPageLoader';
import TestRunPageLoader from '../../TestRunPageLoader';

export interface Env {
    api: FakeApi,
    sse: FakeSse,
    // the devices of the hub
    devices?: IDeviceData[],
}

const NO_DEVICES: IDeviceData[] = [];

// renderAt shows the page the url leads to, in the app's routes and providers;
// the hub knows the project "p1" with the app 1, as after its first load
export const renderAt = (url: string, {api, sse, devices = NO_DEVICES}: Env): RenderResult => renderWithProviders(
    <Routes>
        {projectRoute(appRoute(<>
            <Route path="test/:testId/runs/last" element={<TestRunPageLoader/>}/>
            <Route path="test/:testId/run/:runId" element={<TestRunPageLoader/>}/>
            <Route path="test/:testId/run/:runId/replay" element={<p>replay page</p>}/>
            <Route path="test/:testId/run/:runId/:protocolId" element={<TestProtocolPageLoader/>}/>
        </>))}
    </Routes>,
    {api, sse, url, probe: true, hub: {projects: [aProject('p1', 'P1')], apps: [{id: 1, name: 'App'} as IAppData], devices}},
);
