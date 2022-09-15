import React from 'react';
import AppBar from '@mui/material/AppBar';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import Link from '@mui/material/Link';
import MenuIcon from '@mui/icons-material/Menu';
import NotificationsIcon from '@mui/icons-material/Notifications';
import Toolbar from '@mui/material/Toolbar';
import Tooltip from '@mui/material/Tooltip';
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import { FormControl, MenuItem, Select, SelectChangeEvent } from "@mui/material";
import { ApplicationState, ApplicationStateActions } from "./application.state";
import { useLocation, useNavigate } from "react-router-dom";

const lightColor = 'rgba(255, 255, 255, 0.7)';

interface DefaultHeaderProps {
    onDrawerToggle: () => void;
    appstate: ApplicationState;
    dispatch: any;
    color: string;
}

const DefaultHeader: React.FC<DefaultHeaderProps> = (props) => {

    const {onDrawerToggle, appstate, dispatch, color} = props;

    const location = useLocation()
    const navigate = useNavigate()

    const handleChange = (event: SelectChangeEvent) => {
        if (appstate.projectId !== event.target.value) {
            if (appstate.projectId !== null) {
                navigate(location.pathname.replace(appstate.projectId, event.target.value));
            } else {
                navigate(`/project/${event.target.value}`)
            }
            // dispatch({type: ApplicationStateActions.ChangeProject, payload: event.target.value})
        }
    }

    return (
        <React.Fragment>
            <AppBar sx={{backgroundColor: color}} position="sticky" elevation={ 0 }>
                <Toolbar>
                    <Grid container={ true } spacing={ 1 } alignItems="center">
                        <Grid sx={ {display: {sm: 'none', xs: 'block'}} } item={ true }>
                            <IconButton
                                color="default"
                                aria-label="open drawer"
                                onClick={ onDrawerToggle }
                                edge="start"
                            >
                                <MenuIcon/>
                            </IconButton>
                        </Grid>
                        <Grid item={ true } xs={ 2 }>
                            { appstate.project !== null && <ListItem>
                                <ListItemText>
                                    <FormControl variant="standard">
                                        <Select
                                            id="project-select"
                                            defaultValue={appstate.projectId === null ? undefined : appstate.projectId}
                                            label="Project"
                                            autoWidth={ true }
                                            onChange={ handleChange }
                                            disableUnderline={ true }
                                            sx={ {color: "black"} }
                                        >
                                            {
                                                appstate.projects.map(project => (
                                                    <MenuItem key={ `project_item_${ project.Identifier }` }
                                                              value={ project.Identifier }>{ project.Name }</MenuItem>
                                                ))
                                            }
                                        </Select>
                                    </FormControl>
                                </ListItemText>
                            </ListItem>
                            }
                        </Grid>
                        <Grid item={ true } xs={ true }/>
                        <Grid item={ true }>
                            <Link
                                href="https://www.github.com/fsuhrau/automationhub"
                                variant="body2"
                                sx={ {
                                    textDecoration: 'none',
                                    color: 'black',
                                    '&:hover': {
                                        color: 'common.white',
                                    },
                                } }
                                rel="noopener noreferrer"
                                target="_blank"
                            >
                                Go to docs
                            </Link>
                        </Grid>
                        <Grid item={ true }>
                            <Tooltip title="Alerts • No alerts">
                                <IconButton color="default">
                                    <NotificationsIcon/>
                                </IconButton>
                            </Tooltip>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
        </React.Fragment>
    );
};

export default DefaultHeader;
