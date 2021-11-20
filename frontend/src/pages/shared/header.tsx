import React, { FC } from 'react';
import AppBar from '@material-ui/core/AppBar';
import Grid from '@material-ui/core/Grid';
import IconButton from '@material-ui/core/IconButton';
import NotificationsIcon from '@material-ui/icons/Notifications';
import Toolbar from '@material-ui/core/Toolbar';
import Tooltip from '@material-ui/core/Tooltip';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';

const lightColor = 'rgba(255, 255, 255, 0.7)';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        secondaryBar: {
            zIndex: 0,
        },
        menuButton: {
            marginLeft: -theme.spacing(1),
        },
        iconButtonAvatar: {
            padding: 4,
        },
        link: {
            textDecoration: 'none',
            color: lightColor,
            '&:hover': {
                color: theme.palette.common.white,
            },
        },
        button: {
            borderColor: lightColor,
        },
    });

interface DefaultHeaderProps extends WithStyles<typeof styles> {
    onDrawerToggle: () => void;
}

const DefaultHeader: FC<DefaultHeaderProps> = (props) => {
    const {classes, onDrawerToggle} = props;

    return (
        <React.Fragment>
            <AppBar color="primary" position="sticky" elevation={ 0 }>
                <Toolbar>
                    <Grid container={ true } spacing={ 1 } alignItems="center">
                        <Grid item={ true } xs={ true }/>
                        <Grid item={ true }>
                            <Tooltip title="Alerts • No alerts">
                                <IconButton color="inherit">
                                    <NotificationsIcon/>
                                </IconButton>
                            </Tooltip>
                        </Grid>
                        { /*<Grid item>
                            <IconButton color="inherit" className={ classes.iconButtonAvatar }>
                                <Avatar src="/static/images/avatar/1.jpg" alt="My Avatar"/>
                            </IconButton>
                        </Grid> */
                        }
                    </Grid>
                </Toolbar>
            </AppBar>
        </React.Fragment>
    );
};

export default withStyles(styles)(DefaultHeader);
