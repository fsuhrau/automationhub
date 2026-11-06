import * as React from 'react';
import MuiAvatar from '@mui/material/Avatar';
import MuiListItemAvatar from '@mui/material/ListItemAvatar';
import MenuItem from '@mui/material/MenuItem';
import ListItemText from '@mui/material/ListItemText';
import Select, {SelectChangeEvent, selectClasses} from '@mui/material/Select';
import {styled} from '@mui/material/styles';
import {getPlatformName} from "../types/platform.type.enum";
import PlatformTypeIcon from "./PlatformTypeIcon";
import {useProjectContext} from "../hooks/ProjectProvider";
import {Divider} from "@mui/material";
import ListItemIcon from "@mui/material/ListItemIcon";
import {AddRounded} from "@mui/icons-material";
import {useLocation, useNavigate} from "react-router-dom";
import {useHubState} from "../hooks/HubStateProvider";
import {getAllApps} from "../services/app.service";
import {HubStateActions} from "../application/HubState";
import {rememberApp, useSelectedApp} from "../hooks/useSelectedApp";

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

// ApplicationSelection picks the app that the app pages (tests, bundles) show.
// It only navigates when the user picks an app: on an app page it goes to the
// same section of the new app, elsewhere to the new app's tests. Project pages
// (overview, devices, settings) stay where they are, since they don't depend
// on the app.
export default function ApplicationSelection() {
    const navigate = useNavigate();
    const {pathname} = useLocation();
    const {state, dispatch} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const {app} = useSelectedApp();

    React.useEffect(() => {
        if (state.apps == null) {
            getAllApps(projectIdentifier).then(apps => {
                dispatch({type: HubStateActions.AppsUpdate, payload: apps})
            })
        }
    }, [state.apps, projectIdentifier]);

    // keep the remembered app in line with the page, so the menu links of
    // project pages lead back to the app last worked on
    React.useEffect(() => {
        if (app) rememberApp(app.id as number);
    }, [app?.id]);

    const handleChange = (event: SelectChangeEvent) => {
        const value = event.target.value;
        const base = `/project/${projectIdentifier}`;
        if (value === ADD_APP) {
            navigate(`${base}/settings`);
            return;
        }
        rememberApp(+value);
        const section = pathname.match(/\/app:\d+\/bundles/) ? 'bundles' : 'tests';
        navigate(`${base}/app:${value}/${section}`);
    };

    return (
        <Select
            labelId="Application-select"
            id="Application-simple-select"
            value={app ? `${app.id}` : ''}
            onChange={handleChange}
            displayEmpty
            inputProps={{'aria-label': 'App'}}
            fullWidth
            sx={{
                maxHeight: 56,
                '&.MuiList-root': {
                    p: '8px',
                },
                [`& .${selectClasses.select}`]: {
                    display: 'flex',
                    alignItems: 'center',
                    gap: '2px',
                    pl: 1,
                },
            }}
        >
            {
                state.apps?.map(a => (
                    <MenuItem key={`app_item_${a.id}`} value={`${a.id}`}>
                        <ListItemAvatar>
                            <Avatar alt={a.name} variant={'rounded'}>
                                <PlatformTypeIcon platformType={a.platform}/>
                            </Avatar>
                        </ListItemAvatar>
                        <ListItemText primary={a.name} secondary={getPlatformName(a.platform)}/>
                    </MenuItem>
                ))
            }
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
