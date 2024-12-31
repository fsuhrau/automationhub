import React, {useEffect} from 'react';
import type {} from '@mui/material/themeCssVarsAugmentation';
import {Outlet, useLocation, useParams} from 'react-router-dom';
import CssBaseline from '@mui/material/CssBaseline';
import Moment from 'react-moment';
import {SSEProvider} from 'react-hooks-sse';
import {Box} from '@mui/system';
import {ApplicationStateActions} from "./ApplicationState";
import {ApplicationProps} from "./ApplicationProps";
import AppTheme from "../shared-theme/AppTheme";
import MainMenu from "../components/MainMenu";
import AppNavbar from "../components/AppNavbar";
import {alpha} from "@mui/material/styles";
import Stack from "@mui/material/Stack";
import Header from "../components/Header";
import {
    chartsCustomizations,
    dataGridCustomizations,
    datePickersCustomizations,
    treeViewCustomizations,
} from '../theme/customizations';
import {ProjectProvider} from "../hooks/ProjectProvider";

Moment.globalLocale = 'de';

const xThemeComponents = {
    ...chartsCustomizations,
    ...dataGridCustomizations,
    ...datePickersCustomizations,
    ...treeViewCustomizations,
};

declare module '@mui/material/Paper' {
    interface PaperPropsVariantOverrides {
        paper_content: true;
    }
}

const HubMainLayout: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const {appState, dispatch} = props;

    const handleDrawerToggle = (): void => {
        dispatch({type: ApplicationStateActions.ToggleMobileOpen});
    };

    return <ProjectProvider appState={appState} dispatch={dispatch}>
        <SSEProvider endpoint="/api/sse">
            <AppTheme {...props} themeComponents={xThemeComponents}>
                <CssBaseline enableColorScheme/>
                <Box sx={{display: 'flex'}}>
                    <MainMenu appState={appState} dispatch={dispatch}/>
                    <AppNavbar appState={appState} dispatch={dispatch}/>
                    {/* Main content */}
                    <Box
                        component="main"
                        sx={(theme) => ({
                            flexGrow: 1,
                            backgroundColor: theme.vars
                                ? `rgba(${theme.vars.palette.background.defaultChannel} / 1)`
                                : alpha(theme.palette.background.default, 1),
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
            </AppTheme>
        </SSEProvider>
    </ProjectProvider>
};

export default HubMainLayout;