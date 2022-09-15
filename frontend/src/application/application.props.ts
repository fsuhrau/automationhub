import { ApplicationState } from "./application.state";

export interface ApplicationProps {
    appState: ApplicationState;
    dispatch: any;
}