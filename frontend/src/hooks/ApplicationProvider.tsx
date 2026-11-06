import React, {ReactNode} from "react";
import {Navigate, Outlet, useParams} from "react-router-dom";
import {useProjectContext} from "./ProjectProvider";
import {useHubState} from "./HubStateProvider";
import PageLoader from "../components/PageLoader";

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
    const {state} = useHubState();
    const {appId: urlAppId} = useParams<{ appId: string }>();

    if (state.apps === null || state.apps === undefined) {
        // apps of the project are still loading
        return <PageLoader/>;
    }
    // redirect to project settings if there are no apps
    if (state.apps.length === 0) {
        return (<Navigate to={`/project/${projectIdentifier}/settings?tab=apps`}/>);
    }

    const known = (id: number | null) => id !== null && !isNaN(id) && state.apps!.some(a => a.id === id);
    let appId: number | null = urlAppId ? +urlAppId.replace("app:", "") : null;
    if (!urlAppId) {
        // the stored app may belong to another project
        let stored: number | null = null;
        try {
            const value = localStorage.getItem('appId');
            stored = value ? +value : null;
        } catch {
            stored = null;
        }
        appId = known(stored) ? stored : state.apps[0].id as number;
    }


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