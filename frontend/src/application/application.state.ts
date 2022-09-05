import IProject from "../project/project";

export const STATE_ACTIONS = {
    CHANGE_PROJECT: 'change_project',
    UPDATE_PROJECTS: 'update_projects',
    TOGGLE_MOBILE_OPEN: 'toggle_mobile_open',
}
export interface ApplicationState {
    mobileOpen: boolean,
    project: IProject | null,
    projects: IProject[],
}

export const INITIAL_STATE : ApplicationState = {
    mobileOpen: false,
    project: null,
    projects: [],
}

const appReducer = (state: ApplicationState, action: any) => {
    switch (action.type) {
        case STATE_ACTIONS.CHANGE_PROJECT:
            return {
                ...state,
                project: state.projects.find(p => p.ID === action.payload as string),
            }
        case STATE_ACTIONS.UPDATE_PROJECTS:
            return {
                ...state,
                projects: action.payload,
            }
        case STATE_ACTIONS.TOGGLE_MOBILE_OPEN:
            return {
                ...state,
                mobileOpen: !state.mobileOpen,
            }
        default:
            return state
    }
};

export default appReducer;