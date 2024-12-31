import { ApplicationState } from "./ApplicationState";

export interface ApplicationProps {
    appState: ApplicationState;
    dispatch?: any;
}