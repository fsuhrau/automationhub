import React from 'react';
import {
  BrowserRouter as Router,
  Switch,
  Route,
} from "react-router-dom";
import {
  createTheme,
  createStyles,
  ThemeProvider,
  withStyles,
  WithStyles,
} from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Hidden from '@material-ui/core/Hidden';
import Typography from '@material-ui/core/Typography';
import Link from '@material-ui/core/Link';
import Navigator from './components/navigator';

import Content from './Content';
import Header from './Header';

import TestHeader from './pages/tests/Header'
import TestContent from './pages/tests/Content'
import DevicesContent from './pages/devices/device.content'
import DeviceHeader from './pages/devices/device.header'

function Copyright() {
  return (
    <Typography variant="body2" color="textSecondary" align="center">
      {'Copyright © '}
      <Link color="inherit" href="https://www.automation-hub.com/">
        AutomationHUB
      </Link>{' '}
      {new Date().getFullYear()}
      {'.'}
    </Typography>
  );
}

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

export interface AppProps extends WithStyles<typeof styles> { }

function App(props: AppProps) {
  const { classes } = props;
  const [mobileOpen, setMobileOpen] = React.useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  return (
    <Router>
      <ThemeProvider theme={theme}>
        <div className={classes.root}>
          <CssBaseline />
          <nav className={classes.drawer}>
            <Hidden smUp implementation="js">
              <Navigator
                PaperProps={{ style: { width: drawerWidth } }}
                variant="temporary"
                open={mobileOpen}
                onClose={handleDrawerToggle}
              />
            </Hidden>
            <Hidden xsDown implementation="css">
              <Navigator PaperProps={{ style: { width: drawerWidth } }} />
            </Hidden>
          </nav>
          <div className={classes.app}>
            <Switch>
              <Route path="/tests">
                <TestHeader onDrawerToggle={handleDrawerToggle} />
              </Route>
              <Route path="/results">
                <Header onDrawerToggle={handleDrawerToggle} />
              </Route>
              <Route path="/performance">
                <Header onDrawerToggle={handleDrawerToggle} />
              </Route>
              <Route path="/settings">
                <Header onDrawerToggle={handleDrawerToggle} />
              </Route>
              <Route path="/apps">
                <Header onDrawerToggle={handleDrawerToggle} />
              </Route>
              <Route path="/users">
                <Header onDrawerToggle={handleDrawerToggle} />
              </Route>
              <Route path="/devices">
                <DeviceHeader onDrawerToggle={handleDrawerToggle} />
              </Route>
            </Switch>
            <main className={classes.main}>
              <Switch>
                <Route path="/tests">
                  <TestContent />
                </Route>
                <Route path="/results">
                  <Content />
                </Route>
                <Route path="/performance">
                  <Content />
                </Route>
                <Route path="/settings">
                  <Content />
                </Route>
                <Route path="/apps">
                  <Content />
                </Route>
                <Route path="/users">
                  <Content />
                </Route>
                <Route path="/devices">
                  <DevicesContent />
                </Route>
              </Switch>
            </main>
            <footer className={classes.footer}>
              <Copyright />
            </footer>
          </div>
        </div>
      </ThemeProvider>
    </Router>
  );
}

export default withStyles(styles)(App);