import * as React from 'react';
import {useEffect} from 'react';
import MuiAvatar from '@mui/material/Avatar';
import MuiListItemAvatar from '@mui/material/ListItemAvatar';
import MenuItem from '@mui/material/MenuItem';
import ListItemText from '@mui/material/ListItemText';
import ListSubheader from '@mui/material/ListSubheader';
import Select, {SelectChangeEvent, selectClasses} from '@mui/material/Select';
import {styled} from '@mui/material/styles';
import {ApplicationProps} from "../application/ApplicationProps";
import {getPlatformName} from "../types/platform.type.enum";
import PlatformTypeIcon from "./PlatformTypeIcon";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useNavigate} from "react-router-dom";
import {ApplicationStateActions} from "../application/ApplicationState";
import {Divider} from "@mui/material";
import ListItemIcon from "@mui/material/ListItemIcon";
import {AddRounded} from "@mui/icons-material";

const Avatar = styled(MuiAvatar)(({theme}) => ({
    width: 28,
    height: 28,
    backgroundColor: (theme.vars || theme).palette.background.paper,
    color: (theme.vars || theme).palette.text.secondary,
    border: `1px solid ${(theme.vars || theme).palette.divider}`,
}));

const ListItemAvatar = styled(MuiListItemAvatar)({
    minWidth: 0,
    marginRight: 12,
});

export default function ApplicationSelection(props: ApplicationProps) {

    const {appState, dispatch} = props;

    const {project, projectId} = useProjectContext();

    const navigate = useNavigate();

    const handleChange = (event: SelectChangeEvent) => {
        if (appState.appId != +event.target.value) {
            dispatch({type: ApplicationStateActions.ChangeActiveApp, payload: +event.target.value})
        }
    };

    const appId = appState.appId !== 0 && appState.appId !== null ? appState.appId.toString() : project.Apps === undefined || project.Apps.length === 0 ? "" : project.Apps[0].ID.toString();
    if (appId !== appState.appId?.toString() && appId !== "") {
        dispatch({type: ApplicationStateActions.ChangeActiveApp, payload: appId})
    }

    return (
        <Select
            labelId="Application-select"
            id="Application-simple-select"
            value={appId}
            onChange={handleChange}
            displayEmpty
            inputProps={{'aria-label': 'Select Application'}}
            fullWidth
            sx={{
                maxHeight: 56,
                width: 215,
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
            <ListSubheader sx={{pt: 0}}>Production</ListSubheader>
            {
                project.Apps.map(app => (
                    <MenuItem key={`app_item_${app.ID}`} value={`${app.ID}`}>
                        <ListItemAvatar>
                            <Avatar alt={app.Name}>
                                <PlatformTypeIcon platformType={app.Platform}/>
                                {/*<DevicesRoundedIcon sx={{fontSize: '1rem'}}/>*/}
                            </Avatar>
                        </ListItemAvatar>
                        <ListItemText primary={app.Name} secondary={getPlatformName(app.Platform)}/>
                    </MenuItem>
                ))
            }
            {/*
      <MenuItem value={10}>
        <ListItemAvatar>
          <Avatar alt="Sitemark App">
            <SmartphoneRoundedIcon sx={{ fontSize: '1rem' }} />
          </Avatar>
        </ListItemAvatar>
        <ListItemText primary="Sitemark-app" secondary="Mobile application" />
      </MenuItem>
      <MenuItem value={20}>
        <ListItemAvatar>
          <Avatar alt="Sitemark Store">
            <DevicesRoundedIcon sx={{ fontSize: '1rem' }} />
          </Avatar>
        </ListItemAvatar>
        <ListItemText primary="Sitemark-Store" secondary="Web app" />
      </MenuItem>
      <ListSubheader>Development</ListSubheader>
      <MenuItem value={30}>
        <ListItemAvatar>
          <Avatar alt="Sitemark Store">
            <ConstructionRoundedIcon sx={{ fontSize: '1rem' }} />
          </Avatar>
        </ListItemAvatar>
        <ListItemText primary="Sitemark-Admin" secondary="Web app" />
      </MenuItem>
      */}
            <Divider sx={{mx: -1}}/>
            <MenuItem value={40}>
                <ListItemIcon>
                    <AddRounded/>
                </ListItemIcon>
                <ListItemText primary="Add product" secondary="Web app"/>
            </MenuItem>
        </Select>
    );
}
