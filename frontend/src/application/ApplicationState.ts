import IProject from "../project/project";
import { useParams } from "react-router-dom";
import { IAppData } from "../types/app";

export enum ApplicationStateActions {
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
    appId: number | null,
}

export const InitialApplicationState: ApplicationState = {
    mobileOpen: false,
    projects: [],
    appId: null,
}

export function appReducer(state: ApplicationState, action: ApplicationStateAction): ApplicationState {

    const {type, payload} = action;

    switch (type) {
        case ApplicationStateActions.ChangeActiveApp: {
            return {
                ...state,
                appId: payload as number,
            }
        }
        case ApplicationStateActions.AddNewApp: {
            return {
                ...state,
                projects: state.projects.map(p => {
                    if (p.ID == payload.projectId) {
                        p.Apps.push(payload)
                    }
                    return p
                })
            }
        }
        case ApplicationStateActions.UpdateAppAttribute: {
            return {
                ...state,
                projects: state.projects.map(p => {
                    if (payload.appId as number == payload.projectId) {
                        p.Apps = p.Apps.map(a => (a.ID as number) === (payload.appId as number) ? {...a, [payload.attribute]: payload.value} as IAppData : a as IAppData)
                    }
                    return p
                })
            }
        }
        case ApplicationStateActions.UpdateProjectAttribute: {
            return {
                ...state,
                projects: state.projects.map(p => {
                    if (p.ID == payload.projectId) {
                        return {
                            ...p,
                            [payload.attribute]: payload.value
                        }
                    }
                    return p
                })
            }
        }

        case ApplicationStateActions.UpdateProjects: {
            return {
                ...state,
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
