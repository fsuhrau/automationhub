import React from "react";
import {Navigate, useParams} from "react-router-dom";
import IProject from "../project/project";
import {ApplicationState} from "../application/ApplicationState";

type ProjectContextType = {
    project: IProject
    projectId: string,
}

const initialState: ProjectContextType = {
    project: {} as IProject,
    projectId: "",
}

const ProjectContext = React.createContext(initialState);

type ProjectProps = {
    appState: ApplicationState;
    dispatch?: any;
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

    const { appState } = props;
    let {project_id} = useParams();

    if (project_id === undefined || project_id === "undefined" || project_id === null || project_id === "null") return <Navigate to="/"/>;

    const project = appState.projects.find(p => p.Identifier === project_id)
    if (project === undefined) return <Navigate to="/"/>

    return (
        <ProjectContext.Provider value={{
            project: project,
            projectId: project_id,
        }}>
            {props.children}
        </ProjectContext.Provider>
    );
};