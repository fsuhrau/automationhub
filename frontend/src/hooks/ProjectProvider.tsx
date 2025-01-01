import React from "react";
import {Navigate, useParams} from "react-router-dom";
import IProject from "../project/project";
import {useHubState} from "./HubStateProvider";

type ProjectContextType = {
    project: IProject
    projectIdentifier: string,
}

const initialState: ProjectContextType = {
    project: {} as IProject,
    projectIdentifier: "",
}

const ProjectContext = React.createContext(initialState);

type ProjectProps = {
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

    const {state} = useHubState()

    let {project_identifier} = useParams();


    if (project_identifier === undefined || project_identifier === "undefined" || project_identifier === null || project_identifier === "null") {
        debugger;
        return <Navigate to="/"/>;
    }

    const project = state.projects.find(p => p.Identifier === project_identifier)
    if (project === undefined) {
        debugger;
        return <Navigate to="/"/>
    }

    return (
        <ProjectContext.Provider value={{
            project: project,
            projectIdentifier: project_identifier,
        }}>
            {props.children}
        </ProjectContext.Provider>
    );
};