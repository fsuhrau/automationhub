import React, { useEffect, useReducer } from 'react';
import Typography from '@mui/material/Typography';
import Link from '@mui/material/Link';
import { Outlet, useLocation, useParams } from 'react-router-dom';
import CssBaseline from '@mui/material/CssBaseline';
import Navigator from './navigator';
import Moment from 'react-moment';
import { SSEProvider } from 'react-hooks-sse';
import { useMediaQuery } from '@mui/material';
import { Box } from '@mui/system';
import { ApplicationStateActions, appReducer, InitialApplicationState } from "./application.state";
import { getProjects } from "../project/project.service";
import DefaultHeader from "./header";
import theme from "../style/theme";
import { ApplicationProps } from "./application.props";
import { ProjectProvider } from "../project/project.context";

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


declare module '@mui/material/Paper' {
    interface PaperPropsVariantOverrides {
        paper_content: true;
    }
}

const drawerWidth = 256;

const App: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const {appState, dispatch} = props;

    const isSmUp = useMediaQuery(theme.breakpoints.up('sm'));

    const handleDrawerToggle = (): void => {
        dispatch({type: ApplicationStateActions.ToggleMobileOpen});
    };

    let params = useParams();
    const location = useLocation();

    useEffect(() => {
        getProjects().then(response => {
            dispatch({type: ApplicationStateActions.UpdateProjects, payload: response.data})
        })

        if (params.project_id !== null && params.project_id !== undefined && params.project_id !== appState.projectId) {
            dispatch({type: ApplicationStateActions.ChangeProject, payload: params.project_id})
        }
    }, [appState.projectId, params.project_id])

    return <SSEProvider endpoint="/api/sse">
            <Box sx={ {display: 'flex', minHeight: '100vh'} }>
                <CssBaseline/>
                { appState.project !== null && <Box
                    component="nav"
                    sx={ {width: {sm: drawerWidth}, flexShrink: {sm: 0}} }
                >
                    { isSmUp ? null : (
                        <Navigator
                            appState={ appState }
                            PaperProps={ {style: {width: drawerWidth}} }
                            variant="temporary"
                            open={ appState.mobileOpen }
                            onClose={ handleDrawerToggle }
                        />
                    ) }
                    <Navigator
                        appState={ appState }
                        PaperProps={ {style: {width: drawerWidth}} }
                        sx={ {display: {sm: 'block', xs: 'none'}} }
                    />
                </Box> }
                <Box sx={ {flex: 1, display: 'flex', flexDirection: 'column'} }>
                    <DefaultHeader color={'#eaeff1'} appstate={ appState } dispatch={ dispatch } onDrawerToggle={ handleDrawerToggle }/>
                    <Box component="main" sx={ {flex: 1, padding: 2, bgcolor: '#eaeff1'} }>
                        <ProjectProvider>
                            <Outlet />
                        </ProjectProvider>
                    </Box>
                    <Box component="footer" sx={ {p: 2, bgcolor: '#eaeff1'} }>
                        <Copyright/>
                    </Box>
                </Box>
            </Box>
    </SSEProvider>
};

export default App;