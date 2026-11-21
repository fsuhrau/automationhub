import React from "react";
import {Navigate, Route, Routes} from "react-router-dom";
import SignIn from "../pages/sign-in/SignIn";
import SignUp from "../pages/sign-up/SignUp";
import PrivateRoute from "../router/PrivateRoute";
import HubMainLayout from "../application/HubMainLayout";
import ProjectsIndexPage from "../project/ProjectsIndexPage";
import ProjectDashboard from "../pages/main/ProjectDashboard";
import TestsIndexPage from "../pages/tests/TestsIndexPage";
import AppBundlesPage from "../pages/apps/AppBundlesPage";
import AppBundlePage from "../pages/apps/AppBundlePage";
import AppBundleComparePage from "../pages/apps/AppBundleComparePage";
import TestPageLoader from "../pages/tests/TestPageLoader";
import EditTestPage from "../pages/tests/EditTestPage";
import TestProtocolPageLoader from "../pages/tests/TestProtocolPageLoader";
import TestRunPageLoader from "../pages/tests/TestRunPageLoader";
import {TestReplayPageLoader} from "../pages/tests/TestReplayPage";
import SettingsPage from "../pages/settings/SettingsPage";
import DevicePageLoader from "../pages/devices/DevicePageLoader";
import DeviceControlPage from "../pages/devices/DeviceControlPage";
import DevicesContent from "../pages/devices/DevicesPage";
import ProfilePage from "../pages/profile/ProfilePage";
import ApplicationPage from "../hooks/ApplicationProvider";

// The path patterns below are the shapes that lib/routes.ts builds URLs for
// (lib/routes.test.ts checks that every pattern has a builder). Keep both in step.

// the pages of one app of a project: tests, bundles, runs
const appPages = (
    <Route path={":project_identifier/:appId"} element={<ApplicationPage/>}>
        <Route path={'bundles'} element={<AppBundlesPage/>}/>
        <Route path={'bundles/compare'} element={<AppBundleComparePage/>}/>
        <Route path={'bundle/:bundleId'} element={<AppBundlePage/>}/>
        <Route path={"tests"} element={<TestsIndexPage/>}/>
        <Route path={'test/:testId'} element={<TestPageLoader edit={false}/>}/>
        <Route path={'test/new'} element={<EditTestPage/>}/>
        <Route path={'test/:testId/edit'} element={<TestPageLoader edit={true}/>}/>
        <Route path={'test/:testId/run/:runId/replay'} element={<TestReplayPageLoader/>}/>
        <Route path={'test/:testId/run/:runId/:protocolId'} element={<TestProtocolPageLoader/>}/>
        <Route path={"test/:testId/run/:runId"} element={<TestRunPageLoader/>}/>
        <Route path={"test/:testId/runs/last"} element={<TestRunPageLoader/>}/>
        <Route path={"results"}/>
        <Route path={"performance"}/>
    </Route>
);

// the pages of the devices of a project
const devicePages = (
    <>
        <Route path={":project_identifier/device/:deviceId/control"} element={<DeviceControlPage/>}/>
        <Route path={":project_identifier/device/:deviceId/edit"} element={<DevicePageLoader edit={true}/>}/>
        <Route path={":project_identifier/device/:deviceId"} element={<DevicePageLoader edit={false}/>}/>
        <Route path={":project_identifier/devices"} element={<DevicesContent/>}/>
    </>
);

// the pages inside the project frame (menu, header)
const projectPages = (
    <Route path={"project"} element={<HubMainLayout/>}>
        <Route path={":project_identifier"} element={<ProjectDashboard/>}/>
        {/* the overview is a project page; old app links lead there */}
        <Route path={":project_identifier/:appId/home"} element={<Navigate to={"../.."} relative={"path"} replace={true}/>}/>
        {appPages}
        <Route path={":project_identifier/settings"} element={<SettingsPage/>}/>
        <Route path={":project_identifier/users"}/>
        {devicePages}
    </Route>
);

// every page of the app; all but sign in and sign up need a session
export const AppRoutes: React.FC = () => (
    <Routes>
        <Route path="/login" element={<SignIn/>}/>
        <Route path="/signup" element={<SignUp/>}/>

        <Route element={<PrivateRoute/>}>
            <Route path={"/"}>
                <Route index={true} element={<ProjectsIndexPage/>}/>
                <Route path={"profile"} element={<ProfilePage/>}/>
                {projectPages}
            </Route>
        </Route>
    </Routes>
);
