import React from 'react';
import AppBar from '@mui/material/AppBar';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import Link from '@mui/material/Link';
import MenuIcon from '@mui/icons-material/Menu';
import NotificationsIcon from '@mui/icons-material/Notifications';
import Toolbar from '@mui/material/Toolbar';
import Tooltip from '@mui/material/Tooltip';

const lightColor = 'rgba(255, 255, 255, 0.7)';

interface DefaultHeaderProps {
    onDrawerToggle: () => void;
}

const DefaultHeader: React.FC<DefaultHeaderProps> = (props) => {

    const { onDrawerToggle } = props;

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
                        <Grid item={ true } xs={ true }/>
                        <Grid item={ true }>
                            <Link
                                href="/doc"
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
