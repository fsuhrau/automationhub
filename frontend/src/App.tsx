
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import AuthProvider from "./hooks/AuthProvider";
import SignIn from "./pages/sign-in/SignIn";
import PrivateRoute from "./router/PrivateRoute";
import SignUp from "./pages/sign-up/SignUp";
import React, {useReducer} from "react";
import {appReducer, InitialApplicationState} from "./application/ApplicationState";
import HubMainLayout from "./application/HubMainLayout";
import ProjectsIndexPage from "./project/ProjectsIndexPage";
import ProjectDashboard from "./pages/main/ProjectDashboard";
import TestContent from "./pages/tests/tests.content";
import AppBundlesPage from "./pages/apps/appbundles.page";
import TestPageLoader from "./pages/tests/test.page.loader";
import AddTestPage from "./pages/tests/add.test.content";
import TestProtocolLoader from "./pages/tests/test.protocol.loader";
import TestRunPageLoader from "./pages/tests/test.run.page.loader";
import SettingsPage from "./pages/settings/settings.page";
import DevicePageLoader from "./pages/devices/device.page.loader";
import DevicesManagerContent from "./pages/devices/devices.manager.content";
import DevicesContent from "./pages/devices/device.content";
import ApplicationPage from "./hooks/ApplicationProvider";

function App() {

    const [state, dispatch] = useReducer(appReducer, InitialApplicationState);

    return (
        <div className="App">
            <Router>
                <AuthProvider>
                    <Routes>
                        <Route path="/login" element={<SignIn />} />
                        <Route path="/signup" element={<SignUp />} />

                        <Route element={<PrivateRoute />}>
                            <Route path={ "/" } >
                                <Route index={true} element={ <ProjectsIndexPage appState={ state } dispatch={ dispatch }/> }/>
                                <Route path={ "project" } element={<HubMainLayout appState={ state } dispatch={ dispatch } />} >
                                    <Route path={ ":project_id" } element={ <ProjectDashboard/> }/>
                                    <Route path={ ":project_id/tests" } element={ <TestContent appState={ state } dispatch={ dispatch }/> }/>
                                    <Route path={ ":project_id/bundles" } element={ <AppBundlesPage appState={ state } dispatch={ dispatch }/> }/>
                                    <Route path={ ":project_id/app" } element={<ApplicationPage appState={ state } dispatch={ dispatch } />}>
                                        <Route path={ 'bundles' } element={<AppBundlesPage appState={ state } dispatch={ dispatch }/>}/>
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
                                    <Route path={ ":project_id/settings" } element={ <SettingsPage appState={ state } dispatch={ dispatch }/> }/>
                                    <Route path={ ":project_id/users" }/>
                                    <Route path={ ":project_id/device/:deviceId/edit" } element={ <DevicePageLoader edit={ true }/> }/>
                                    <Route path={ ":project_id/device/:deviceId" } element={ <DevicePageLoader edit={ false }/> }/>
                                    <Route path={ ":project_id/devices/manager" } element={ <DevicesManagerContent/> }/>
                                    <Route path={ ":project_id/devices" } element={ <DevicesContent/> }/>
                                </Route>
                            </Route>
                        </Route>
                    </Routes>
                </AuthProvider>
            </Router>
        </div>
    );
}

export default App;
