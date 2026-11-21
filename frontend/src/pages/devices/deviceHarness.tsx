import {ReactElement} from 'react';
import {RenderResult} from '@testing-library/react';
import {Route, Routes} from 'react-router-dom';
import {ApiClient} from '../../api/ApiClient';
import {aProject} from '../../testing/fixtures';
import {LocationProbe, projectRoute, renderWithProviders} from '../../testing/renderWithProviders';
import IDeviceData from '../../types/device';
import {INodeData} from '../../types/node';

export interface PageOptions {
    api?: ApiClient,
    // the route of the page below /project/:project_identifier, and the url opened
    route: string,
    url: string,
    devices?: IDeviceData[],
    nodes?: INodeData[],
}

// renders the page in its route inside the providers, the project p1 known to
// the hub; a navigation away from it shows the new path in the element "location"
export const renderPage = (page: ReactElement, {api, route, url, devices, nodes}: PageOptions): RenderResult =>
    renderWithProviders(
        <Routes>
            {projectRoute(<>
                <Route path={route} element={page}/>
                <Route path="*" element={<LocationProbe/>}/>
            </>)}
            <Route path="*" element={<LocationProbe/>}/>
        </Routes>,
        {api, url, hub: {projects: [aProject('p1', 'P1')], devices, nodes}},
    );
