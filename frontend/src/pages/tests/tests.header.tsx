import React, { FC, useContext } from 'react';
import AppBar from '@material-ui/core/AppBar';
import Grid from '@material-ui/core/Grid';
import Hidden from '@material-ui/core/Hidden';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import NotificationsIcon from '@material-ui/icons/Notifications';
import Toolbar from '@material-ui/core/Toolbar';
import Tooltip from '@material-ui/core/Tooltip';
import Typography from '@material-ui/core/Typography';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import { TestContext } from '../../context/test.context';

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

interface HeaderProps extends WithStyles<typeof styles> {
    onDrawerToggle: () => void;
}

const TestsHeader: FC<HeaderProps> = (props) => {
    const { classes, onDrawerToggle } = props;
    const testContext = useContext(TestContext);
    const { test } = testContext;

    return (
        <React.Fragment>
            <AppBar color="primary" position="sticky" elevation={ 0 }>
                <Toolbar>
                    <Grid container={ true } spacing={ 1 } alignItems="center">
                        <Hidden smUp={ true }>
                            <Grid item={ true }>
                                <IconButton
                                    color="inherit"
                                    aria-label="open drawer"
                                    onClick={ onDrawerToggle }
                                    className={ classes.menuButton }
                                >
                                    <MenuIcon/>
                                </IconButton>
                            </Grid>
                        </Hidden>
                        <Grid item={ true } xs={ true }/>
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
            <AppBar
                component="div"
                className={ classes.secondaryBar }
                color="primary"
                position="static"
                elevation={ 0 }
            >
                <Toolbar>
                    <Grid container={ true } alignItems="center" spacing={ 1 }>
                        <Grid item={ true } xs={ true }>
                            <Typography color="inherit" variant="h5" component="h1">
                                {test?.Name}
                            </Typography>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
        </React.Fragment>
    );
};

export default withStyles(styles)(TestsHeader);
