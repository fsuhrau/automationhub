import {DistributiveOmit} from "@mui/types";
import {DrawerProps} from "@mui/material/Drawer";
import {ApplicationState} from "../application/ApplicationState";

export interface NavigatorProps extends DistributiveOmit<DrawerProps, 'classes'> {
    appState: ApplicationState
    dispatch?: any
}