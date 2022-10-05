import React, { useEffect, useReducer } from "react";
import { useParams } from "react-router-dom";
import { appReducer, InitialApplicationState } from "../application/application.state";

type ProjectContextType = {
    projectId: string | null,
}

const initialState: ProjectContextType = {
    projectId: null,
}

const ProjectContext = React.createContext(initialState);

type ProjectProps = {
    projectId?: string | null;
    children: React.ReactNode;
};

export const useProjectContext = (): ProjectContextType => {
    const contextState = React.useContext(ProjectContext);
    if (contextState === null) {
        throw new Error('useProjectContext must be used within a ProjectProvider tag');
    }
    return contextState;
}

export const ProjectProvider: React.FC<ProjectProps> = (props: ProjectProps) => {

    const [state, dispatch] = useReducer(appReducer, InitialApplicationState);

    let { project_id } = useParams();

    return (
        <ProjectContext.Provider value={{
            projectId: project_id === undefined ? null : project_id,
        }}>
            { state.projects !== null && props.children }
            </ProjectContext.Provider>
    );
};