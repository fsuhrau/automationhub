import * as React from 'react';
import MuiAvatar from '@mui/material/Avatar';
import MuiListItemAvatar from '@mui/material/ListItemAvatar';
import MenuItem from '@mui/material/MenuItem';
import ListItemText from '@mui/material/ListItemText';
import Select, {SelectChangeEvent, selectClasses} from '@mui/material/Select';
import {styled} from '@mui/material/styles';
import {Divider} from "@mui/material";
import ListItemIcon from "@mui/material/ListItemIcon";
import {AddRounded} from "@mui/icons-material";
import {useLocation, useNavigate} from "react-router-dom";
import {getPlatformName} from "../types/platform.type.enum";
import {IAppData} from "../types/app";
import PlatformTypeIcon from "./PlatformTypeIcon";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import {rememberApp, useSelectedApp} from "../hooks/useSelectedApp";
import {routes} from "../lib/routes";

const Avatar = styled(MuiAvatar)(({theme}) => ({
    backgroundColor: theme.palette.background.paper,
    color: theme.palette.text.primary,
    border: `1px solid ${theme.palette.divider}`,
}));

const ListItemAvatar = styled(MuiListItemAvatar)({
    minWidth: 0,
    marginRight: 12,
});

const ADD_APP = 'add';

const selectSx = {
    maxHeight: 56,
    '&.MuiList-root': {p: '8px'},
    [`& .${selectClasses.select}`]: {display: 'flex', alignItems: 'center', gap: '2px', pl: 1},
};

// keep the remembered app in line with the page, so the menu links of
// project pages lead back to the app last worked on
const useRememberSelectedApp = (app: IAppData | null): void => {
    const appId = app?.id;
    React.useEffect(() => {
        if (appId !== undefined) rememberApp(appId);
    }, [appId]);
};

// the page the picked app opens: bundles stay bundles, everything else is the tests
const appPage = (pathname: string, appId: number, projectIdentifier: string): string => {
    const app = routes.project(projectIdentifier).app(appId);
    return /\/app:\d+\/bundles/.test(pathname) ? app.bundles() : app.tests();
};

// a function, not a component: Select passes its props to its MenuItem children
const appOption = (app: IAppData): React.ReactElement => (
    <MenuItem key={`app_item_${app.id}`} value={`${app.id}`}>
        <ListItemAvatar>
            <Avatar alt={app.name} variant={'rounded'}>
                <PlatformTypeIcon platformType={app.platform}/>
            </Avatar>
        </ListItemAvatar>
        <ListItemText primary={app.name} secondary={getPlatformName(app.platform)}/>
    </MenuItem>
);

// ApplicationSelection picks the app that the app pages (tests, bundles) show.
// It only navigates when the user picks an app: on an app page it goes to the
// same section of the new app, elsewhere to the new app's tests. Project pages
// (overview, devices, settings) stay where they are, since they don't depend
// on the app.
export default function ApplicationSelection(): React.ReactElement {
    const navigate = useNavigate();
    const {pathname} = useLocation();
    const {state} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const {app} = useSelectedApp();
    useRememberSelectedApp(app);

    const handleChange = (event: SelectChangeEvent): void => {
        const value = event.target.value;
        if (value === ADD_APP) {
            navigate(routes.project(projectIdentifier).settings('apps'));
            return;
        }
        rememberApp(+value);
        navigate(appPage(pathname, +value, projectIdentifier));
    };

    return (
        <Select labelId="Application-select" id="Application-simple-select" value={app ? `${app.id}` : ''}
            onChange={handleChange} displayEmpty={true} inputProps={{'aria-label': 'App'}} fullWidth={true} sx={selectSx}>
            {state.apps?.map(appOption)}
            <Divider sx={{mx: -1}}/>
            <MenuItem value={ADD_APP}>
                <ListItemIcon>
                    <AddRounded/>
                </ListItemIcon>
                <ListItemText primary="Add app" secondary="in project settings"/>
            </MenuItem>
        </Select>
    );
}
