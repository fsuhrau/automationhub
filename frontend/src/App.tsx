import type {} from '@mui/material/themeCssVarsAugmentation';

import {BrowserRouter as Router, Route, Routes} from "react-router-dom";
import SignIn from "./pages/sign-in/SignIn";
import PrivateRoute from "./router/PrivateRoute";
import SignUp from "./pages/sign-up/SignUp";
import React from "react";
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
import {HubStateProvider} from "./hooks/HubStateProvider";
import {AuthProvider} from "./hooks/AuthProvider";

function App() {

    return (
        <div className="App">
            <HubStateProvider>
                <Router>
                    <AuthProvider>
                        <Routes>
                            <Route path="/login" element={<SignIn/>}/>
                            <Route path="/signup" element={<SignUp/>}/>

                            <Route element={<PrivateRoute/>}>
                                <Route path={"/"}>
                                    <Route index={true}
                                           element={<ProjectsIndexPage/>}/>
                                    <Route path={"project"}
                                           element={<HubMainLayout/>}>
                                        <Route path={":project_identifier"}
                                               element={<ProjectDashboard/>}/>
                                        <Route path={":project_identifier/tests"}
                                               element={<TestContent/>}/>
                                        <Route path={":project_identifier/bundles"}
                                               element={<AppBundlesPage/>}/>
                                        <Route path={":project_identifier/app"}
                                               element={<ApplicationPage/>}>
                                            <Route path={'bundles'}
                                                   element={<AppBundlesPage/>}/>
                                            <Route path={"tests"}
                                                   element={<TestContent/>}/>
                                            <Route path={'test/:testId'}
                                                   element={<TestPageLoader edit={false}/>}/>
                                            <Route path={'test/new'} element={<AddTestPage/>}/>
                                            <Route path={'test/:testId/edit'}
                                                   element={<TestPageLoader edit={true}/>}/>
                                            <Route path={'test/:testId/run/:runId/:protocolId'}
                                                   element={<TestProtocolLoader/>}/>
                                            <Route path={"test/:testId/run/:runId"} element={<TestRunPageLoader/>}/>
                                            <Route path={"test/:testId/runs/last"} element={<TestRunPageLoader/>}/>
                                            <Route path={"results"}/>
                                            <Route path={"performance"}/>
                                        </Route>
                                        <Route path={":project_identifier/settings"}
                                               element={<SettingsPage/>}/>
                                        <Route path={":project_identifier/users"}/>
                                        <Route path={":project_identifier/device/:deviceId/edit"}
                                               element={<DevicePageLoader edit={true}/>}/>
                                        <Route path={":project_identifier/device/:deviceId"}
                                               element={<DevicePageLoader edit={false}/>}/>
                                        <Route path={":project_identifier/devices/manager"}
                                               element={<DevicesManagerContent/>}/>
                                        <Route path={":project_identifier/devices"} element={<DevicesContent/>}/>
                                    </Route>
                                </Route>
                            </Route>
                        </Routes>
                    </AuthProvider>
                </Router>
            </HubStateProvider>
        </div>
    );
}

export default App;
