import {BrowserRouter as Router, Route, Routes} from "react-router-dom";
import SignIn from "./pages/sign-in/SignIn";
import PrivateRoute from "./router/PrivateRoute";
import SignUp from "./pages/sign-up/SignUp";
import React from "react";
import HubMainLayout from "./application/HubMainLayout";
import ProjectsIndexPage from "./project/ProjectsIndexPage";
import ProjectDashboard from "./pages/main/ProjectDashboard";
import TestContent from "./pages/tests/TestsIndexPage";
import AppBundlesPage from "./pages/apps/AppBundlesPage";
import TestPageLoader from "./pages/tests/TestPageLoader";
import AddTestPage from "./pages/tests/add.test.content";
import TestProtocolPageLoader from "./pages/tests/TestProtocolPageLoader";
import TestRunPageLoader from "./pages/tests/TestRunPageLoader";
import SettingsPage from "./pages/settings/SettingsPage";
import DevicePageLoader from "./pages/devices/DevicePageLoader";
import DevicesManagerContent from "./pages/devices/devices.manager.content";
import DevicesContent from "./pages/devices/DevicesPage";
import ApplicationPage from "./hooks/ApplicationProvider";
import {HubStateProvider} from "./hooks/HubStateProvider";
import {AuthProvider} from "./hooks/AuthProvider";
import AppTheme from "./shared-theme/AppTheme";
import {
    chartsCustomizations,
    dataGridCustomizations,
    datePickersCustomizations,
    treeViewCustomizations
} from "./theme/customizations";
import TestsIndexPage from "./pages/tests/TestsIndexPage";

const xThemeComponents = {
    ...chartsCustomizations,
    ...dataGridCustomizations,
    ...datePickersCustomizations,
    ...treeViewCustomizations,
};

declare module '@mui/material/Paper' {
    interface PaperPropsVariantOverrides {
        paper_content: true;
    }
}

function App(props: any) {
    return (
        <div className="App">
            <AppTheme {...props} themeComponents={xThemeComponents}>
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
                                            <Route path={":project_identifier/:appId/home"}
                                                   element={<ProjectDashboard/>}/>
                                            <Route path={":project_identifier/:appId"}
                                                   element={<ApplicationPage/>}>
                                                <Route path={'bundles'}
                                                       element={<AppBundlesPage/>}/>
                                                <Route path={"tests"}
                                                       element={<TestsIndexPage/>}/>
                                                <Route path={'test/:testId'}
                                                       element={<TestPageLoader edit={false}/>}/>
                                                <Route path={'test/new'} element={<AddTestPage/>}/>
                                                <Route path={'test/:testId/edit'}
                                                       element={<TestPageLoader edit={true}/>}/>
                                                <Route path={'test/:testId/run/:runId/:protocolId'}
                                                       element={<TestProtocolPageLoader/>}/>
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
            </AppTheme>
        </div>
    );
}

export default App;
