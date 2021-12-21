import React from 'react';

import AppBar from '@mui/material/AppBar';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import MenuIcon from '@mui/icons-material/Menu';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Toolbar from '@mui/material/Toolbar';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import NotificationsNoneIcon from '@mui/icons-material/NotificationsNone';

const lightColor = 'rgba(255, 255, 255, 0.7)';

interface HeaderProps {
    onDrawerToggle: () => void;
}

const Header: React.FC<HeaderProps> = (props: HeaderProps) => {
    const { onDrawerToggle } = props;
    return (
        <>
            <AppBar color="primary" position="sticky" elevation={ 0 }>
                <Toolbar>
                    <Grid container={ true } spacing={ 1 } alignItems="center">
                        <Grid item={ true }>
                            <IconButton
                                color="inherit"
                                aria-label="open drawer"
                                onClick={onDrawerToggle}
                                edge="start"
                            >
                                <MenuIcon/>
                            </IconButton>
                        </Grid>
                        <Grid item={ true } xs={ true }/>
                        <Grid item={ true }>
                            <Tooltip title="Alerts • No alerts">
                                <IconButton color="inherit">
                                    <NotificationsNoneIcon/>
                                </IconButton>
                            </Tooltip>
                        </Grid>
                        {/* <Grid item>
              <IconButton color="inherit" className={classes.iconButtonAvatar}>
                <Avatar src="/static/images/avatar/1.jpg" alt="My Avatar" />
              </IconButton>
            </Grid> */ }
                    </Grid>
                </Toolbar>
            </AppBar>
            <AppBar
                component="div"
                color="primary"
                position="static"
                elevation={0}
                sx={{ zIndex: 0 }}
            >
                <Toolbar>
                    <Grid container={ true } alignItems="center" spacing={ 1 }>
                        <Grid item={ true } xs={ true }>
                            <Typography color="inherit" variant="h5" component="h1">
                                Tests
                            </Typography>
                        </Grid>
                        <Grid item={ true }>
                            <Button
                                sx={{ borderColor: lightColor }}
                                variant="outlined"
                                color="inherit"
                                size="small"
                            >
                                Web setup
                            </Button>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <AppBar component="div" position="static" elevation={0} sx={{ zIndex: 0 }}>
                <Tabs value={ 0 } textColor="inherit">
                    <Tab label="Users"/>
                    <Tab label="Sign-in method"/>
                    <Tab label="Templates"/>
                    <Tab label="Usage"/>
                </Tabs>
            </AppBar>
        </>
    );
};

export default Header;
