import React, { FC, useState } from 'react';
import { createStyles, createTheme, ThemeProvider, withStyles, WithStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import Link from '@material-ui/core/Link';
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom';
import CssBaseline from '@material-ui/core/CssBaseline';
import Hidden from '@material-ui/core/Hidden';
import Navigator from './components/navigator';
import TestContent from './pages/tests/tests.content';
import DevicesContent from './pages/devices/device.content';
import Content from './Content';
import AddTestPage from './pages/tests/add.test.content';
import TestRunPage from './pages/tests/test.run.content';
import TestProtocolPage from './pages/tests/test.protocol.content';
import Moment from 'react-moment';
import AppsPage from './pages/apps/apps.content';
import { SSEProvider } from 'react-hooks-sse';
import { AppContext } from './context/app.context';
import { TestContextProvider } from './context/test.context';
import DefaultHeader from './pages/shared/header';
import EditTestPage from './pages/tests/edit.test.content';
import TestPage from './pages/tests/test.content';

Moment.globalLocale = 'de';

const Copyright: FC = () => (
    <Typography variant="body2" color="textSecondary" align="center">
        { 'Copyright © ' }
        <Link color="inherit" href="https://www.automation-hub.com/">
            AutomationHUB
        </Link>{ ' ' }
        { new Date().getFullYear() }
        { '.' }
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
                [ theme.breakpoints.up('md') ]: {
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
        [ theme.breakpoints.up('sm') ]: {
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

    return <SSEProvider endpoint="/api/sse">
        <AppContext.Provider value={ { title: '' } }>
            <Router>
                <ThemeProvider theme={ theme }>
                    <div className={ classes.root }>
                        <CssBaseline/>
                        <nav className={ classes.drawer }>
                            <Hidden smUp={ true } implementation="js">
                                <Navigator
                                    PaperProps={ { style: { width: drawerWidth } } }
                                    variant="temporary"
                                    open={ mobileOpen }
                                    onClose={ handleDrawerToggle }
                                />
                            </Hidden>
                            <Hidden xsDown={ true } implementation="css">
                                <Navigator PaperProps={ { style: { width: drawerWidth } } }/>
                            </Hidden>
                        </nav>
                        <div className={ classes.app }>
                            <DefaultHeader onDrawerToggle={ handleDrawerToggle }/>
                            <main className={ classes.main }>
                                <Switch>
                                    <Route path="/web/tests">
                                        <TestContent/>
                                    </Route>
                                    <Route path={ '/web/test/new' }>
                                        <AddTestPage/>
                                    </Route>
                                    <Route path={ '/web/test/:testId/run/:runId/:protocolId' }>
                                        <TestProtocolPage/>
                                    </Route>
                                    <Route path="/web/test/:testId/runs/last">
                                        <TestContextProvider>
                                            <TestRunPage/>
                                        </TestContextProvider>
                                    </Route>
                                    <Route path="/web/test/:testId/run/:runId">
                                        <TestContextProvider>
                                            <TestRunPage/>
                                        </TestContextProvider>
                                    </Route>
                                    <Route path={ '/web/test/:testId/edit' }>
                                        <TestContextProvider>
                                            <TestPage edit={true}/>
                                        </TestContextProvider>
                                    </Route>
                                    <Route path={ '/web/test/:testId' }>
                                        <TestContextProvider>
                                            <TestPage edit={false}/>
                                        </TestContextProvider>
                                    </Route>
                                    <Route path="/web/results">
                                        <Content/>
                                    </Route>
                                    <Route path="/web/performance">
                                        <Content/>
                                    </Route>
                                    <Route path="/web/settings">
                                        <Content/>
                                    </Route>
                                    <Route path="/web/apps">
                                        <AppsPage/>
                                    </Route>
                                    <Route path="/web/users">
                                        <Content/>
                                    </Route>
                                    <Route path="/web/devices">
                                        <DevicesContent/>
                                    </Route>
                                </Switch>
                            </main>
                            <footer className={ classes.footer }>
                                <Copyright/>
                            </footer>
                        </div>
                    </div>
                </ThemeProvider>
            </Router>
        </AppContext.Provider>
    </SSEProvider>;
};

export default withStyles(styles)(App);
