import React from "react";
import {Navigate, useParams} from "react-router-dom";
import IProject from "../project/project";
import {useHubState} from "./HubStateProvider";
import PageLoader from "../components/PageLoader";
import {rememberProject} from "../project/projectsModel";
import {routes} from "../lib/routes";

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

export const useProjectContext = (): ProjectContextType => React.useContext(ProjectContext);

// links built from a missing value end up as "undefined" or "null"
const isMissing = (identifier: string | undefined): identifier is undefined =>
    identifier === undefined || identifier === "undefined" || identifier === "null";

export const ProjectProvider: React.FC<ProjectProps> = ({children}) => {
    const {state} = useHubState()
    const {project_identifier} = useParams();

    if (isMissing(project_identifier)) {
        return <Navigate to={routes.projects()}/>;
    }
    if (state.projects === null) {
        // projects are still loading
        return <PageLoader/>;
    }
    const project = state.projects.find(p => p.identifier === project_identifier)
    if (project === undefined) {
        return <Navigate to={routes.projects()}/>;
    }

    // the projects page lists the last opened project first
    rememberProject(project.identifier);

    return (
        <ProjectContext.Provider value={{project, projectIdentifier: project_identifier}}>
            {children}
        </ProjectContext.Provider>
    );
};
