import IProject from "../project/project";
import { useParams } from "react-router-dom";
import { IAppData } from "../types/app";

export enum ApplicationStateActions {
    ChangeProject = 'ChangeProject',
    UpdateProjects = 'UpdateProjects',
    ToggleMobileOpen = 'ToggleMobileOpen',
    UpdateProjectAttribute = 'UpdateProjectAttribute',
    UpdateAppAttribute = 'UpdateAppAttribute',
    AddNewApp = 'AddNewApp',
    ChangeActiveApp = 'ChangeActiveApp',
}

export interface ApplicationStateAction {
    type: ApplicationStateActions,
    payload?: any,
}

export interface ApplicationState {
    mobileOpen: boolean,
    projects: IProject[],
    projectId: string | null,
    appId: number | null,
    project: IProject | null,
}

export const InitialApplicationState: ApplicationState = {
    mobileOpen: false,
    projects: [],
    projectId: null,
    appId: null,
    project: null,
}

export function appReducer(state: ApplicationState, action: ApplicationStateAction): ApplicationState {

    const {type, payload} = action;

    switch (type) {
        case ApplicationStateActions.ChangeActiveApp: {
            // const project = state.projects.find(p => p.Identifier === payload as string);
            return {
                ...state,
                appId: payload as number,
            }
        }
        case ApplicationStateActions.AddNewApp: {
            if (state.project !== null) {
                const apps = state.project?.Apps as IAppData[];
                apps.push(payload)
                return {
                    ...state,
                    project: {...state.project, Apps: apps },
                }
            }
            return state
        }
        case ApplicationStateActions.UpdateAppAttribute: {
            if (state.project !== null) {
                const apps = state.project?.Apps.map(a => (a.ID as number) === (payload.app_id as number) ? {...a, [payload.attribute]: payload.value} as IAppData : a as IAppData) as IAppData[];
                return {
                    ...state,
                    project: {...state.project, Apps: apps },
                }
            }
            return state
        }
        case ApplicationStateActions.UpdateProjectAttribute: {
            return {
                ...state,
                project: {...state.project, [payload.attribute]: payload.value} as IProject
            }
        }
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
