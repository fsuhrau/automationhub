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
    const {project} = useProjectContext();
    const regex = /app:\d+/g;

    const {appId: urlAppId} = useParams<{ appId: string }>();
    const navigate = useNavigate();
    const [appId, setAppId] = React.useState<string>(() => {
        if (urlAppId) {
            return urlAppId.replace("app:", "")
        }
        return localStorage.getItem('appId') || (project.Apps.length > 0 ? project.Apps[0].ID.toString() : "");
    });
    const location = useLocation();

    const handleChange = (event: SelectChangeEvent) => {
        const newAppId = event.target.value;
        setAppId(newAppId);
        localStorage.setItem('appId', newAppId);
        if (location.pathname.indexOf("app:") >= 0) {
            const newPath = location.pathname.replace(regex, `app:${newAppId}`);
            navigate(newPath);
        } else {
            navigate(`/project/${project.Identifier}/app:${newAppId}/home`);
        }
    };

    React.useEffect(() => {
        if (urlAppId) {
            const cleanId = urlAppId.replace("app:", "")
            if (cleanId !== appId) {
                setAppId(urlAppId);
                localStorage.setItem('appId', urlAppId);
            }
        } else if (!urlAppId && appId) {

            if (location.pathname.indexOf("app:") >= 0) {
                const newPath = location.pathname.replace(regex, `app:${appId}`);
                navigate(newPath);
            } else if (location.pathname.indexOf("/settings") == -1
                && location.pathname.indexOf("/users") == -1
                && location.pathname.indexOf("/devices") == -1
                && location.pathname.indexOf("/device") == -1
            ) {
                navigate(`/project/${project.Identifier}/app:${appId}/home`);
            }
        }
    }, [urlAppId, appId, project.Identifier, navigate]);

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
            <MenuItem value={40}>
                <ListItemIcon>
                    <AddRounded/>
                </ListItemIcon>
                <ListItemText primary="Add product" secondary="Web app"/>
            </MenuItem>
        </Select>
    );
}
