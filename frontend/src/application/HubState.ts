import IProject from "../project/project";
import { useParams } from "react-router-dom";
import { IAppData } from "../types/app";
import IHubStatsData from "../types/hub.stats";

export enum HubStateActions {
    UpdateProjects = 'UpdateProjects',
    ToggleMobileOpen = 'ToggleMobileOpen',
    UpdateProjectAttribute = 'UpdateProjectAttribute',
    UpdateAppAttribute = 'UpdateAppAttribute',
    AddNewApp = 'AddNewApp',
    ChangeActiveApp = 'ChangeActiveApp',
}

export interface HubStateAction {
    type: HubStateActions,
    payload?: any,
}

export interface HubState {
    mobileOpen: boolean,
    projects: IProject[],
    //appId: number | null,
    stats: IHubStatsData | null,
}

export const InitialHubState: HubState = {
    mobileOpen: false,
    projects: [],
    //appId: null,
    stats: null,
}

export function hubReducer(state: HubState, action: HubStateAction): HubState {

    const {type, payload} = action;

    switch (type) {
        /*
        case HubStateActions.ChangeActiveApp: {
            return {
                ...state,
                appId: payload as number,
            }
        }
         */
        case HubStateActions.AddNewApp: {
            return {
                ...state,
                projects: state.projects.map(p => {
                    if (p.Identifier == payload.projectIdentifier) {
                        p.Apps.push(payload)
                    }
                    return p
                })
            }
        }
        case HubStateActions.UpdateAppAttribute: {
            return {
                ...state,
                projects: state.projects.map(p => {
                    if (p.Identifier === payload.projectIdentifier) {
                        p.Apps = p.Apps.map(a => (a.ID as number) === (payload.appId as number) ? {...a, [payload.attribute]: payload.value} as IAppData : a as IAppData)
                    }
                    return p
                })
            }
        }
        case HubStateActions.UpdateProjectAttribute: {
            return {
                ...state,
                projects: state.projects.map(p => {
                    if (p.Identifier == payload.projectIdentifier) {
                        return {
                            ...p,
                            [payload.attribute]: payload.value
                        }
                    }
                    return p
                })
            }
        }

        case HubStateActions.UpdateProjects: {
            return {
                ...state,
                projects: payload,
            }
        }
        case HubStateActions.ToggleMobileOpen:
            return {
                ...state,
                mobileOpen: !state.mobileOpen,
            }
        default:
            return state
    }
}
