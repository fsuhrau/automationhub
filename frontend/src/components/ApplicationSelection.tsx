import * as React from 'react';
import MuiAvatar from '@mui/material/Avatar';
import MuiListItemAvatar from '@mui/material/ListItemAvatar';
import MenuItem from '@mui/material/MenuItem';
import ListItemText from '@mui/material/ListItemText';
import ListSubheader from '@mui/material/ListSubheader';
import Select, {SelectChangeEvent, selectClasses} from '@mui/material/Select';
import {styled} from '@mui/material/styles';
import {getPlatformName} from "../types/platform.type.enum";
import PlatformTypeIcon from "./PlatformTypeIcon";
import {useProjectContext} from "../hooks/ProjectProvider";
import {Divider} from "@mui/material";
import ListItemIcon from "@mui/material/ListItemIcon";
import {AddRounded} from "@mui/icons-material";
import {useLocation, useNavigate, useParams} from "react-router-dom";

const Avatar = styled(MuiAvatar)(({theme}) => ({
    backgroundColor: theme.palette.background.paper,
    color: theme.palette.text.primary,
    border: `1px solid ${theme.palette.divider}`,
}));

const ListItemAvatar = styled(MuiListItemAvatar)({
    minWidth: 0,
    marginRight: 12,
});

export default function ApplicationSelection() {

    const navigate = useNavigate();
    const location = useLocation();

    const {project} = useProjectContext();
    const regex = /app:\d+/g;

    const {appId} = useParams<{ appId: string }>();

    const extractAppId = (): string => {
        if (appId) {
            return appId.replace("app:", "")
        }
        let id = localStorage.getItem('appId');
        if (id !== null) {
            return id.replace("app:", "")
        }
        if (project.Apps.length > 0) {
            return project.Apps[0].ID.toString();
        }
        return "";
    }

    const [state, setState] = React.useState<{ appId: string }>({
        appId: extractAppId(),
    })


    const handleChange = (event: SelectChangeEvent) => {
        const newAppId = event.target.value;
        if (+newAppId < 0) {
            navigate(`/project/${project.Identifier}/settings`);
            return;
        }
        setState(prevState => ({...prevState, appId: newAppId}));
        localStorage.setItem('appId', newAppId);
        if (location.pathname.indexOf("app:") >= 0) {
            const newPath = location.pathname.replace(regex, `app:${newAppId}`);
            navigate(newPath);
        } else {
            navigate(`/project/${project.Identifier}/app:${newAppId}/home`);
        }
    };

    React.useEffect(() => {
        if (appId) {
            const cleanId = appId.replace("app:", "")
            if (cleanId !== state.appId) {
                setState(prevState => ({...prevState, appId: cleanId}));
                localStorage.setItem('appId', cleanId);
            }
        } else if (!appId && state.appId) {

            if (location.pathname.indexOf("app:") >= 0) {
                const newPath = location.pathname.replace(regex, `app:${state.appId}`);
                navigate(newPath);
            } else if (location.pathname.indexOf("/settings") == -1
                && location.pathname.indexOf("/users") == -1
                && location.pathname.indexOf("/devices") == -1
                && location.pathname.indexOf("/device") == -1
            ) {
                navigate(`/project/${project.Identifier}/app:${state.appId}/home`);
            }
        }
    }, [appId, state.appId, project.Identifier, navigate]);

    return (
        <Select
            labelId="Application-select"
            id="Application-simple-select"
            value={state.appId}
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
                            <Avatar alt={app.Name} variant={'rounded'}>
                                <PlatformTypeIcon platformType={app.Platform}/>
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
            <MenuItem value={-1}>
                <ListItemIcon>
                    <AddRounded/>
                </ListItemIcon>
                <ListItemText primary="Add product" secondary="Web app"/>
            </MenuItem>
        </Select>
    );
}
