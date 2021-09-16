import { FC, useState } from 'react';
import { createStyles, createTheme, ThemeProvider, withStyles, WithStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import Link from '@material-ui/core/Link';
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom';
import CssBaseline from '@material-ui/core/CssBaseline';
import Hidden from '@material-ui/core/Hidden';
import Navigator from './components/navigator';
import TestHeader from './pages/tests/tests.header';
import TestContent from './pages/tests/tests.content';
import DeviceHeader from './pages/devices/device.header';
import DevicesContent from './pages/devices/device.content';
import Header from './Header';
import Content from './Content';
import AddTestPage from './pages/tests/add.test.content';
import TestRunsPage from './pages/tests/test.runs.content';
import TestRunPage from './pages/tests/test.run.content';
import TestProtocolPage from './pages/tests/test.protocol.content';
import Moment from 'react-moment';
import AppsPage from './pages/apps/apps.content';
import AppsHeader from './pages/apps/apps.header';

Moment.globalLocale = 'de';

const Copyright: FC = () => (
    <Typography variant="body2" color="textSecondary" align="center">
        {'Copyright © '}
        <Link color="inherit" href="https://www.automation-hub.com/">
            AutomationHUB
        </Link>{' '}
        {new Date().getFullYear()}
        {'.'}
    </Typography>
);

let theme = createTheme({
    palette: {
        primary: {
            light: '#63ccff',
            main: '#009be5',
            dark: '#006db3',
        },
    },
    typography: {
        h5: {
            fontWeight: 500,
            fontSize: 26,
            letterSpacing: 0.5,
        },
    },
    shape: {
        borderRadius: 8,
    },
    props: {
        MuiTab: {
            disableRipple: true,
        },
    },
    mixins: {
        toolbar: {
            minHeight: 48,
        },
    },
});

theme = {
    ...theme,
    overrides: {
        MuiDrawer: {
            paper: {
                backgroundColor: '#18202c',
            },
        },
        MuiButton: {
            label: {
                textTransform: 'none',
            },
            contained: {
                boxShadow: 'none',
                '&:active': {
                    boxShadow: 'none',
                },
            },
        },
        MuiTabs: {
            root: {
                marginLeft: theme.spacing(1),
            },
            indicator: {
                height: 3,
                borderTopLeftRadius: 3,
                borderTopRightRadius: 3,
                backgroundColor: theme.palette.common.white,
            },
        },
        MuiTab: {
            root: {
                textTransform: 'none',
                margin: '0 16px',
                minWidth: 0,
                padding: 0,
                [theme.breakpoints.up('md')]: {
                    padding: 0,
                    minWidth: 0,
                },
            },
        },
        MuiIconButton: {
            root: {
                padding: theme.spacing(1),
            },
        },
        MuiTooltip: {
            tooltip: {
                borderRadius: 4,
            },
        },
        MuiDivider: {
            root: {
                backgroundColor: '#404854',
            },
        },
        MuiListItemText: {
            primary: {
                fontWeight: theme.typography.fontWeightMedium,
            },
        },
        MuiListItemIcon: {
            root: {
                color: 'inherit',
                marginRight: 0,
                '& svg': {
                    fontSize: 20,
                },
            },
        },
        MuiAvatar: {
            root: {
                width: 32,
                height: 32,
            },
        },
    },
};

const drawerWidth = 256;

const styles = createStyles({
    root: {
        display: 'flex',
        minHeight: '100vh',
    },
    drawer: {
        [theme.breakpoints.up('sm')]: {
            width: drawerWidth,
            flexShrink: 0,
        },
    },
    app: {
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
    },
    main: {
        flex: 1,
        padding: theme.spacing(6, 4),
        background: '#eaeff1',
    },
    footer: {
        padding: theme.spacing(2),
        background: '#eaeff1',
    },
});

export type AppProps = WithStyles<typeof styles>;

const App: FC<AppProps> = (props) => {
    const { classes } = props;
    const [mobileOpen, setMobileOpen] = useState(false);

    const handleDrawerToggle = (): void => {
        setMobileOpen(!mobileOpen);
    };

    return <Router>
        <ThemeProvider theme={theme}>
            <div className={classes.root}>
                <CssBaseline/>
                <nav className={classes.drawer}>
                    <Hidden smUp={true} implementation="js">
                        <Navigator
                            PaperProps={{ style: { width: drawerWidth } }}
                            variant="temporary"
                            open={mobileOpen}
                            onClose={handleDrawerToggle}
                        />
                    </Hidden>
                    <Hidden xsDown={true} implementation="css">
                        <Navigator PaperProps={{ style: { width: drawerWidth } }}/>
                    </Hidden>
                </nav>
                <div className={classes.app}>
                    <Switch>
                        <Route path="/tests">
                            <TestHeader onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path={'/test/new'}>
                            <TestHeader onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path={'/test/:testId/run/:runId/:protocolId'}>
                            <TestHeader onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path="/test/:testId/runs">
                            <TestHeader onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path="/test/:testId/runs/last">
                            <TestHeader onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path="/results">
                            <Header onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path="/performance">
                            <Header onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path="/settings">
                            <Header onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path="/apps">
                            <AppsHeader onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path="/users">
                            <Header onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                        <Route path="/devices">
                            <DeviceHeader onDrawerToggle={handleDrawerToggle}/>
                        </Route>
                    </Switch>
                    <main className={classes.main}>
                        <Switch>
                            <Route path="/tests">
                                <TestContent/>
                            </Route>
                            <Route path={'/test/new'}>
                                <AddTestPage/>
                            </Route>
                            <Route path={'/test/:testId/run/:runId/:protocolId'}>
                                <TestProtocolPage/>
                            </Route>
                            <Route path={'/test/:testId/runs/last'}>
                                <TestRunPage/>
                            </Route>
                            <Route path={'/test/:testId/runs'}>
                                <TestRunsPage/>
                            </Route>
                            <Route path="/test/new">
                                <AddTestPage/>
                            </Route>
                            <Route path="/results">
                                <Content/>
                            </Route>
                            <Route path="/performance">
                                <Content/>
                            </Route>
                            <Route path="/settings">
                                <Content/>
                            </Route>
                            <Route path="/apps">
                                <AppsPage/>
                            </Route>
                            <Route path="/users">
                                <Content/>
                            </Route>
                            <Route path="/devices">
                                <DevicesContent/>
                            </Route>
                        </Switch>
                    </main>
                    <footer className={classes.footer}>
                        <Copyright/>
                    </footer>
                </div>
            </div>
        </ThemeProvider>
    </Router>;
};

export default withStyles(styles)(App);
