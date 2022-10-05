import React from "react";
import { useParams } from "react-router-dom";

type ProjectAppContextType = {
    projectId: string,
    appId: number,
}

const initialState: ProjectAppContextType = {
    projectId: "",
    appId: 0,
}

const ProjectAppContext = React.createContext(initialState);

type ProjectAppProps = {
    children: React.ReactNode;
};

export const useProjectAppContext = (): ProjectAppContextType => {
    const contextState = React.useContext(ProjectAppContext);
    if (contextState === null) {
        throw new Error('useProjectAppContext must be used within a ProjectAppProvider tag');
    }
    return contextState;
}

export const ProjectAppProvider: React.FC<ProjectAppProps> = (props: ProjectAppProps) => {

    let { project_id, app_id } = useParams();

    return (
        <ProjectAppContext.Provider value={{
            projectId: project_id === undefined ? "" : project_id,
            appId: app_id === undefined ? 0 : +app_id,
        }}>
            { props.children }
            </ProjectAppContext.Provider>
    );
};