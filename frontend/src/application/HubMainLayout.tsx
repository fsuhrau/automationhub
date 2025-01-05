import React from 'react';
import {Outlet} from 'react-router-dom';
import CssBaseline from '@mui/material/CssBaseline';
import Moment from 'react-moment';
import {SSEProvider} from 'react-hooks-sse';
import {Box} from '@mui/system';
import {HubStateActions} from "./HubState";
import MainMenu from "../components/MainMenu";
import AppNavbar from "../components/AppNavbar";
import {alpha} from "@mui/material/styles";
import Stack from "@mui/material/Stack";
import Header from "../components/Header";
import {ProjectProvider} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";

Moment.globalLocale = 'de';

const HubMainLayout: React.FC = () => {

    const {dispatch} = useHubState()

    const handleDrawerToggle = (): void => {
        dispatch({type: HubStateActions.ToggleMobileOpen, payload: {}});
    };

    return <ProjectProvider>
        <SSEProvider endpoint="/api/sse">
            <CssBaseline enableColorScheme/>
            <Box sx={{display: 'flex'}}>
                <MainMenu/>
                <AppNavbar/>
                {/* Main content */}
                <Box
                    component="main"
                    sx={(theme) => ({
                        flexGrow: 1,
                        backgroundColor: alpha(theme.palette.background.default, 1),
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
        </SSEProvider>
    </ProjectProvider>
};

export default HubMainLayout;