import React, {ReactNode} from "react";
import {Navigate, Outlet, useParams} from "react-router-dom";
import {useProjectContext} from "./ProjectProvider";
import {useHubState} from "./HubStateProvider";
import {parseAppParam, useSelectedApp} from "./useSelectedApp";
import PageLoader from "../components/PageLoader";
import {routes} from "../lib/routes";

interface ApplicationProviderProps {
    children: ReactNode;
}

type ApplicationContextType = {
    appId: number | null,
}

const ApplicationContext = React.createContext<ApplicationContextType>({appId: null});

export const useApplicationContext = (): ApplicationContextType => React.useContext(ApplicationContext);

const ApplicationProvider: React.FC<ApplicationProviderProps> = ({children}) => {
    const {projectIdentifier} = useProjectContext();
    const {state} = useHubState();
    const {appId: param} = useParams<{ appId: string }>();
    // the page's app, or the one picked last, or the first of the project
    const {app} = useSelectedApp();

    if (state.apps === null) {
        // apps of the project are still loading
        return <PageLoader/>;
    }
    if (app === null) {
        // there is no app to show yet
        return <Navigate to={routes.project(projectIdentifier).settings('apps')}/>;
    }
    // an app of the URL that isn't known stays the page's app, so its pages show the error
    return <ApplicationContext.Provider value={{appId: parseAppParam(param) ?? app.id}}>
        {children}
    </ApplicationContext.Provider>;
};

const ApplicationPage: React.FC = () => <ApplicationProvider>
    <Outlet/>
</ApplicationProvider>;

export default ApplicationPage;
