import React, { Dispatch } from 'react';
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
import { FormControl, InputLabel, MenuItem, Select, SelectChangeEvent } from "@mui/material";
import { ApplicationState, STATE_ACTIONS } from "../../application/application.state";

const lightColor = 'rgba(255, 255, 255, 0.7)';

interface DefaultHeaderProps {
    onDrawerToggle: () => void;
    appState: ApplicationState;
    dispatch: any;
}

const DefaultHeader: React.FC<DefaultHeaderProps> = (props) => {

    const { onDrawerToggle, appState, dispatch } = props;

    const handleChange = (event: SelectChangeEvent) => {
        dispatch({type: STATE_ACTIONS.CHANGE_PROJECT, payload: event.target.value})
    };

    return (
        <React.Fragment>
            <AppBar color="primary" position="sticky" elevation={ 0 }>
                <Toolbar>
                    <Grid container={ true } spacing={ 1 } alignItems="center">
                        <Grid sx={ { display: { sm: 'none', xs: 'block' } } } item={true}>
                            <IconButton
                                color="inherit"
                                aria-label="open drawer"
                                onClick={ onDrawerToggle }
                                edge="start"
                            >
                                <MenuIcon/>
                            </IconButton>
                        </Grid>
                        <Grid item={ true } xs={2}>
                            <ListItem>
                                <ListItemText>
                                    <FormControl variant="standard" >
                                        <Select
                                            id="team-select"
                                            value={appState.project?.Name}
                                            label="Project"
                                            autoWidth={true}
                                            onChange={handleChange}
                                            disableUnderline={true}
                                        >
                                            <MenuItem value={10}>Rise of Cultures</MenuItem>
                                            <MenuItem value={20}>Lost Survivors</MenuItem>
                                        </Select>
                                    </FormControl>
                                </ListItemText>
                            </ListItem>
                        </Grid>
                        <Grid item={true} xs={ true }/>
                        <Grid item={ true }>
                            <Link
                                href="https://www.github.com/fsuhrau/automationhub"
                                variant="body2"
                                sx={ {
                                    textDecoration: 'none',
                                    color: lightColor,
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
                                <IconButton color="inherit">
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
