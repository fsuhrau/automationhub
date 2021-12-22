import React, { useState } from 'react';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import Link from '@mui/material/Link';
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom';
import CssBaseline from '@mui/material/CssBaseline';
import Navigator from './components/navigator';
import TestContent from './pages/tests/tests.content';
import DevicesContent from './pages/devices/device.content';
import Content from './Content';
import AddTestPage from './pages/tests/add.test.content';
import TestRunPageLoader from './pages/tests/test.run.page.loader';
import TestProtocolLoader from './pages/tests/test.protocol.loader';
import Moment from 'react-moment';
import AppsPage from './pages/apps/apps.content';
import { SSEProvider } from 'react-hooks-sse';
import { AppContext } from './context/app.context';
import DefaultHeader from './pages/shared/header';
import TestPageLoader from './pages/tests/test.page.loader';
import DevicesManagerContent from './pages/devices/devices.manager.content';
import DevicePageLoader from './pages/devices/device.page.loader';
import { useMediaQuery } from '@mui/material';
import { Box } from '@mui/system';

Moment.globalLocale = 'de';

const Copyright: React.FC = () => {
    return (<Typography variant="body2" color="text.secondary" align="center">
        { 'Copyright © ' }
        <Link color="inherit"
            href="https://www.github.com/fsuhrau/automationhub"
            target="_blank"
        >
            AutomationHUB
        </Link>{ ' ' }
        { new Date().getFullYear() }
        { '.' }
    </Typography>);
};

let theme = createTheme({
    palette: {
        primary: {
            light: '#63ccff',
            main: '#009be5',
            dark: '#006db3',
        },
        secondary: {
            light: '#CC3333',
            main: '#bb1c2a',
            dark: '#951621',
        },
    },
    typography: {
        h1: {
            fontWeight: 500,
            fontSize: 30,
            letterSpacing: 0.5,
        },
        h2: {
            fontWeight: 500,
            fontSize: 29,
            letterSpacing: 0.5,
        },
        h3: {
            fontWeight: 500,
            fontSize: 27,
            letterSpacing: 0.5,
        },
        h4: {
            fontWeight: 500,
            fontSize: 25,
            letterSpacing: 0.5,
        },
        h5: {
            fontWeight: 500,
            fontSize: 20,
            letterSpacing: 0.5,
        },
        h6: {
            fontWeight: 500,
            fontSize: 18,
            letterSpacing: 0.5,
        },
    },
    shape: {
        borderRadius: 8,
    },
    components: {
        MuiTab: {
            defaultProps: {
                disableRipple: true,
            },
        },
        MuiPaper: {
            variants: [
                {
                    props: { variant: 'paper_content' },
                    style: {
                        maxWidth: 1200,
                        margin: 'auto',
                        overflow: 'hidden',
                    },
                },
            ],
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
    components: {
        MuiDrawer: {
            styleOverrides: {
                paper: {
                    backgroundColor: '#081627',
                },
            },
        },
        MuiButton: {
            styleOverrides: {
                root: {
                    textTransform: 'none',
                },
                contained: {
                    boxShadow: 'none',
                    '&:active': {
                        boxShadow: 'none',
                    },
                },
            },
        },
        MuiTabs: {
            styleOverrides: {
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
        },
        MuiTab: {
            styleOverrides: {
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
        },
        MuiIconButton: {
            styleOverrides: {
                root: {
                    padding: theme.spacing(1),
                },
            },
        },
        MuiTooltip: {
            styleOverrides: {
                tooltip: {
                    borderRadius: 4,
                },
            },
        },
        MuiDivider: {
            styleOverrides: {
                root: {
                    backgroundColor: 'rgb(255,255,255,0.15)',
                },
            },
        },
        MuiListItemButton: {
            styleOverrides: {
                root: {
                    '&.Mui-selected': {
                        color: '#4fc3f7',
                    },
                },
            },
        },
        MuiListItemText: {
            styleOverrides: {
                primary: {
                    fontSize: 14,
                    fontWeight: theme.typography.fontWeightMedium,
                },
            },
        },
        MuiListItemIcon: {
            styleOverrides: {
                root: {
                    color: 'inherit',
                    minWidth: 'auto',
                    marginRight: theme.spacing(2),
                    '& svg': {
                        fontSize: 20,
                    },
                },
            },
        },
        MuiAvatar: {
            styleOverrides: {
                root: {
                    width: 32,
                    height: 32,
                },
            },
        },
    },
};

declare module '@mui/material/Paper' {
    interface PaperPropsVariantOverrides {
        paper_content: true;
    }
}

const drawerWidth = 256;

const App: React.FC = () => {
    const [mobileOpen, setMobileOpen] = useState(false);
    const isSmUp = useMediaQuery(theme.breakpoints.up('sm'));

    const handleDrawerToggle = (): void => {
        setMobileOpen(!mobileOpen);
    };

    return <SSEProvider endpoint="/api/sse">
        <AppContext.Provider value={ { title: '' } }>
            <Router>
                <ThemeProvider theme={ theme }>
                    <Box sx={ { display: 'flex', minHeight: '100vh' } }>
                        <CssBaseline/>
                        <Box
                            component="nav"
                            sx={ { width: { sm: drawerWidth }, flexShrink: { sm: 0 } } }
                        >
                            { isSmUp ? null : (
                                <Navigator
                                    PaperProps={ { style: { width: drawerWidth } } }
                                    variant="temporary"
                                    open={ mobileOpen }
                                    onClose={ handleDrawerToggle }
                                />
                            ) }
                            <Navigator
                                PaperProps={ { style: { width: drawerWidth } } }
                                sx={ { display: { sm: 'block', xs: 'none' } } }
                            />
                        </Box>
                        <Box sx={ { flex: 1, display: 'flex', flexDirection: 'column' } }>
                            <DefaultHeader onDrawerToggle={ handleDrawerToggle }/>
                            <Box component="main" sx={ { flex: 1, py: 6, px: 4, bgcolor: '#eaeff1' } }>
                                <Switch>
                                    <Route path="/web/tests">
                                        <TestContent/>
                                    </Route>
                                    <Route path={ '/web/test/new' }>
                                        <AddTestPage/>
                                    </Route>
                                    <Route path={ '/web/test/:testId/run/:runId/:protocolId' }>
                                        <TestProtocolLoader/>
                                    </Route>
                                    <Route path="/web/test/:testId/runs/last">
                                        <TestRunPageLoader/>
                                    </Route>
                                    <Route path="/web/test/:testId/run/:runId">
                                        <TestRunPageLoader/>
                                    </Route>
                                    <Route path={ '/web/test/:testId/edit' }>
                                        <TestPageLoader edit={ true }/>
                                    </Route>
                                    <Route path={ '/web/test/:testId' }>
                                        <TestPageLoader edit={ false }/>
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
                                    <Route path="/web/device/:deviceId/edit">
                                        <DevicePageLoader edit={ true }/>
                                    </Route>
                                    <Route path="/web/device/:deviceId">
                                        <DevicePageLoader edit={ false }/>
                                    </Route>
                                    <Route path="/web/devices/manager">
                                        <DevicesManagerContent/>
                                    </Route>
                                    <Route path="/web/devices">
                                        <DevicesContent/>
                                    </Route>
                                </Switch>
                            </Box>
                            <Box component="footer" sx={ { p: 2, bgcolor: '#eaeff1' } }>
                                <Copyright/>
                            </Box>
                        </Box>
                    </Box>
                </ThemeProvider>
            </Router>
        </AppContext.Provider>
    </SSEProvider>;
};

export default App;