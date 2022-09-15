import IProject from "../project/project";
import { useParams } from "react-router-dom";

export enum ApplicationStateActions {
    ChangeProject = 'ChangeProject',
    UpdateProjects = 'UpdateProjects',
    ToggleMobileOpen = 'ToggleMobileOpen',
}

export interface ApplicationStateAction {
    type: ApplicationStateActions,
    payload?: any,
}

export interface ApplicationState {
    mobileOpen: boolean,
    projects: IProject[],
    projectId: string | null,
    project: IProject | null,
}

export const InitialApplicationState: ApplicationState = {
    mobileOpen: false,
    projects: [],
    projectId: null,
    project: null,
}

export function appReducer(state: ApplicationState, action: ApplicationStateAction): ApplicationState {

    const {type, payload} = action;

    switch (type) {
        case ApplicationStateActions.ChangeProject: {
            const project = state.projects.find(p => p.Identifier === payload as string);
            return {
                ...state,
                projectId: payload as string,
                project: project === undefined ? null : project,
            }
        }
        case ApplicationStateActions.UpdateProjects: {
            const project = state.projects.find(p => p.Identifier === state.projectId);
            return {
                ...state,
                project: project === undefined ? null : project,
                projects: payload,
            }
        }
        case ApplicationStateActions.ToggleMobileOpen:
            return {
                ...state,
                mobileOpen: !state.mobileOpen,
            }
        default:
            return state
    }
}
