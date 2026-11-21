import React, {useEffect} from 'react';
import {render, RenderResult} from '@testing-library/react';
import {MemoryRouter, Outlet, Route, Routes, useLocation} from 'react-router-dom';
import {SSEProvider} from 'react-hooks-sse';
import {ApiClientProvider} from '../../../../api/ApiClientProvider';
import {HubStateActions} from '../../../../application/HubStateActions';
import {ErrorProvider} from '../../../../ErrorProvider';
import {useHubState, HubStateProvider} from '../../../../hooks/HubStateProvider';
import ApplicationPage from '../../../../hooks/ApplicationProvider';
import {ProjectProvider} from '../../../../hooks/ProjectProvider';
import IProject from '../../../../project/project';
import {IAppData} from '../../../../types/app';
import IDeviceData from '../../../../types/device';
import TestProtocolPageLoader from '../../TestProtocolPageLoader';
import TestRunPageLoader from '../../TestRunPageLoader';
import {FakeApi, FakeSse} from './fakes';

// the hub knows the project "p1" with the app 1, as after its first load
const HubSeed: React.FC<{ devices: IDeviceData[], children: React.ReactNode }> = ({devices, children}) => {
    const {state, dispatch} = useHubState();
    useEffect(() => {
        dispatch({type: HubStateActions.ProjectsUpdate, payload: [{id: 1, identifier: 'p1', name: 'P1'} as IProject]});
        dispatch({type: HubStateActions.AppsUpdate, payload: [{id: 1, name: 'App'} as IAppData]});
        dispatch({type: HubStateActions.UpdateDevices, payload: devices});
    }, [dispatch, devices]);
    return state.devices ? <>{children}</> : null;
};

const Where: React.FC = () => {
    const {pathname, search} = useLocation();
    return <output aria-label="location">{pathname}{search}</output>;
};

export interface Env {
    api: FakeApi,
    sse: FakeSse,
    // the devices of the hub
    devices?: IDeviceData[],
}

const NO_DEVICES: IDeviceData[] = [];

// renderAt shows the page the url leads to, in the app's routes and providers
export const renderAt = (url: string, {api, sse, devices = NO_DEVICES}: Env): RenderResult => render(
    <ErrorProvider><ApiClientProvider client={api}><SSEProvider source={sse.source}><HubStateProvider>
        <HubSeed devices={devices}><MemoryRouter initialEntries={[url]}>
            <Routes>
                <Route path="/project/:project_identifier" element={<ProjectProvider><Outlet/></ProjectProvider>}>
                    <Route path=":appId" element={<ApplicationPage/>}>
                        <Route path="test/:testId/runs/last" element={<TestRunPageLoader/>}/>
                        <Route path="test/:testId/run/:runId" element={<TestRunPageLoader/>}/>
                        <Route path="test/:testId/run/:runId/replay" element={<p>replay page</p>}/>
                        <Route path="test/:testId/run/:runId/:protocolId" element={<TestProtocolPageLoader/>}/>
                    </Route>
                </Route>
            </Routes>
            <Where/>
        </MemoryRouter></HubSeed>
    </HubStateProvider></SSEProvider></ApiClientProvider></ErrorProvider>,
);
