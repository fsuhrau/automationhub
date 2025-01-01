import React, {ReactNode} from "react";
import {Navigate, Outlet, useParams} from "react-router-dom";
import {useProjectContext} from "./ProjectProvider";
import {HubState, HubStateActions} from "../application/HubState";
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

    const {state, dispatch} = useHubState()
    const {project, projectIdentifier} = useProjectContext();

    // no app in context
    if (state.appId === null || state.appId === 0) {
        // redirect to project settings if there are no apps
        if (project.Apps === undefined || project.Apps.length === 0) {
            return (<Navigate to={`/project/${projectIdentifier}/settings`}/>);
        }

        dispatch({type: HubStateActions.ChangeActiveApp, payload: project.Apps[0].ID})
        return null
    }

    return (
        <ApplicationContext.Provider value={{
            appId: state.appId,
        }}>
            {children}
        </ApplicationContext.Provider>
    );
};

const ApplicationPage: React.FC = () => {
    debugger;
    return <ApplicationProvider>
        <Outlet />
    </ApplicationProvider>;
};

export default ApplicationPage;