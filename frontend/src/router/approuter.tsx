import React, { useReducer } from "react";
import { appReducer, InitialApplicationState } from "../application/application.state";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import App from "../application/app";
import ProjectMainPage from "../pages/main/project.main.page";
import TestContent from "../pages/tests/tests.content";
import AddTestPage from "../pages/tests/add.test.content";
import TestProtocolLoader from "../pages/tests/test.protocol.loader";
import TestRunPageLoader from "../pages/tests/test.run.page.loader";
import TestPageLoader from "../pages/tests/test.page.loader";
import SettingsPage from "../pages/settings/settings.page";
import DevicePageLoader from "../pages/devices/device.page.loader";
import DevicesManagerContent from "../pages/devices/devices.manager.content";
import DevicesContent from "../pages/devices/device.content";
import ProjectsIndexPage from "../project/projects.index.page";
import { ProjectAppProvider } from "../project/app.context";
import ProjectAppPage from "../pages/tests/project.app.page";
import AppBundlesPage from "../pages/apps/appbundles.page";
import AppSelectionContext from "../pages/apps/app.selection.context";

export const AppRouter: React.FC = () => {
    const [state, dispatch] = useReducer(appReducer, InitialApplicationState);
    return (
        <BrowserRouter>
            <Routes>
                <Route path={ "/" } element={ <App appState={ state } dispatch={ dispatch } /> }>
                    <Route index={true} element={ <ProjectsIndexPage appState={ state } dispatch={ dispatch }/> }/>
                    <Route path={ "project/:project_id" } element={ <ProjectMainPage/> }/>
                    <Route path={ "project/:project_id/tests" } element={ <ProjectAppProvider><TestContent appState={ state } dispatch={ dispatch }/></ProjectAppProvider> }/>
                    <Route path={ "project/:project_id/bundles" } element={ <ProjectAppProvider ><AppBundlesPage appState={ state } dispatch={ dispatch }/></ProjectAppProvider> }/>
                    <Route path={ "project/:project_id/app/:app_id" } element={<ProjectAppPage />}>
                        <Route path={ 'bundles' } element={ <AppSelectionContext appState={ state } dispatch={ dispatch } path={'bundles'}><AppBundlesPage appState={ state } dispatch={ dispatch }/></AppSelectionContext> }/>
                        <Route path={ "tests" } element={ <TestContent appState={ state } dispatch={ dispatch }/> }/>
                        <Route path={ 'test/:testId' } element={ <TestPageLoader appState={ state } edit={ false }/> }/>
                        <Route path={ 'test/new' } element={ <AddTestPage appState={ state }/> }/>
                        <Route path={ 'test/:testId/edit' } element={ <TestPageLoader appState={ state } edit={ true }/> }/>
                        <Route path={ 'test/:testId/run/:runId/:protocolId' } element={ <TestProtocolLoader/> }/>
                        <Route path={ "test/:testId/run/:runId" } element={ <TestRunPageLoader/> }/>
                        <Route path={ "test/:testId/runs/last" } element={ <TestRunPageLoader/> }/>
                        <Route path={ "results" }/>
                        <Route path={ "performance" }/>
                    </Route>
                    <Route path={ "project/:project_id/settings" } element={ <SettingsPage appState={ state } dispatch={ dispatch }/> }/>
                    <Route path={ "project/:project_id/users" }/>
                    <Route path={ "project/:project_id/device/:deviceId/edit" } element={ <DevicePageLoader edit={ true }/> }/>
                    <Route path={ "project/:project_id/device/:deviceId" } element={ <DevicePageLoader edit={ false }/> }/>
                    <Route path={ "project/:project_id/devices/manager" } element={ <DevicesManagerContent/> }/>
                    <Route path={ "project/:project_id/devices" } element={ <DevicesContent/> }/>
                </Route>
            </Routes>
        </BrowserRouter>
    )
};
