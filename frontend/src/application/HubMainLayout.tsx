import React, {useEffect} from 'react';
import {Outlet} from 'react-router-dom';
import CssBaseline from '@mui/material/CssBaseline';
import Moment from 'react-moment';
import {SSEProvider, useSSE} from 'react-hooks-sse';
import {Box} from '@mui/system';
import {HubStateActions} from "./HubState";
import MainMenu from "../components/MainMenu";
import AppNavbar from "../components/AppNavbar";
import Stack from "@mui/material/Stack";
import Header from "../components/Header";
import {ProjectProvider} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import {getProjects} from "../project/project.service";
import HubStateUpdates from './HubStateUpdates';

Moment.globalLocale = 'de';

const HubMainLayout: React.FC = () => {

    const {dispatch, state} = useHubState()

    const handleDrawerToggle = (): void => {
        dispatch({type: HubStateActions.ToggleMobileOpen, payload: {}});
    };

    useEffect(() => {
        if (state.projects == null) {
            getProjects().then(projects => {
                dispatch({type: HubStateActions.ProjectsUpdate, payload: projects})
            })
        }
    }, [state.projects])


    return <ProjectProvider>
        <SSEProvider endpoint="/api/sse">
            <HubStateUpdates>
                <CssBaseline enableColorScheme/>
                <Box sx={{display: 'flex'}}>
                    <MainMenu/>
                    <AppNavbar/>
                    {/* Main content */}
                    <Box
                        component="main"
                        sx={(theme) => ({
                            flexGrow: 1,
                            overflow: 'auto',
                        })}
                    >
                        <Stack
                            spacing={2}
                            sx={{
                                alignItems: 'center',
                                mx: 3,
                                pb: 5,
                                mt: {xs: 8, md: 0},
                            }}
                        >
                            <Header/>
                            <Outlet/>
                            {/*<MainGrid/>*/}
                        </Stack>
                    </Box>
                </Box>
            </HubStateUpdates>
        </SSEProvider>
    </ProjectProvider>
};

export default HubMainLayout;