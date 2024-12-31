import React from "react";
import {Navigate, Outlet, useParams} from "react-router-dom";
import {useProjectContext} from "./ProjectProvider";
import {ApplicationState, ApplicationStateActions} from "../application/ApplicationState";

type ApplicationContextType = {
    appId: number | null,
}

const initialState: ApplicationContextType = {
    appId: null,
}

const ApplicationContext = React.createContext(initialState);

type ApplicationProps = {
    children: React.ReactNode;
    appState: ApplicationState;
    dispatch?: any;
};

export const useApplicationContext = (): ApplicationContextType => {
    const contextState = React.useContext(ApplicationContext);
    if (contextState === null) {
        throw new Error('useApplicationContext must be used within a ApplicationProvider tag');
    }
    return contextState;
}

const ApplicationProvider: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const {project, projectId} = useProjectContext();

    const {appState, dispatch} = props;

    // no app in context
    if (appState.appId === null || appState.appId === 0) {
        // redirect to project settings if there are no apps
        if (project.Apps === undefined || project.Apps.length === 0) {
            return (<Navigate to={`/project/${projectId}/settings`}/>);
        }

        dispatch({type: ApplicationStateActions.ChangeActiveApp, payload: project.Apps[0].ID})
        return null
    }

    return (
        <ApplicationContext.Provider value={{
            appId: appState.appId,
        }}>
            {props.children}
        </ApplicationContext.Provider>
    );
};

type ApplicationPageProps = {
    appState: ApplicationState;
    dispatch?: any;
};

const ApplicationPage: React.FC<ApplicationPageProps> = (props: ApplicationPageProps) => {
    const {appState, dispatch} = props;
    return <ApplicationProvider appState={appState} dispatch={dispatch}>
        <Outlet />
    </ApplicationProvider>;
};

export default ApplicationPage;