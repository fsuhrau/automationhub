import React, {ReactNode} from "react";
import {Navigate, Outlet, useParams} from "react-router-dom";
import {useProjectContext} from "./ProjectProvider";
import {useHubState} from "./HubStateProvider";

interface ApplicationProviderProps {
    children: ReactNode;
}

type ApplicationContextType = {
    appId: number | null,
}

const initialState: ApplicationContextType = {
    appId: null,
}

const ApplicationContext = React.createContext(initialState);

export const useApplicationContext = (): ApplicationContextType => {
    const contextState = React.useContext(ApplicationContext);
    if (contextState === null) {
        throw new Error('useApplicationContext must be used within a ApplicationProvider tag');
    }
    return contextState;
}

const ApplicationProvider: React.FC<ApplicationProviderProps> = ({children}) => {

    const {project, projectIdentifier} = useProjectContext();
    const {appId: urlAppId} = useParams<{ appId: string }>();

    // redirect to project settings if there are no apps
    if (project.apps === undefined || project.apps.length === 0) {
        return (<Navigate to={`/project/${projectIdentifier}/settings`}/>);
    }

    let appIdString
    if (urlAppId) {
        appIdString = urlAppId.replace("app:", "")
    } else {
        appIdString = localStorage.getItem('appId') || (project.apps.length > 0 ? project.apps[0].id : "");
    }
    const appId = appIdString !== "" ? +appIdString : null;


    return (
        <ApplicationContext.Provider value={{
            appId: appId,
        }}>
            {children}
        </ApplicationContext.Provider>
    );
};

const ApplicationPage: React.FC = () => {
    return <ApplicationProvider>
        <Outlet/>
    </ApplicationProvider>;
};

export default ApplicationPage;