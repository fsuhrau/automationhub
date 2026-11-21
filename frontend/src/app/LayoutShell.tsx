import React from 'react';
import {Outlet} from 'react-router-dom';
import CssBaseline from '@mui/material/CssBaseline';
import {Box} from '@mui/system';
import Stack from "@mui/material/Stack";
import MainMenu from "../components/MainMenu";
import AppNavbar from "../components/AppNavbar";
import Header from "../components/Header";

// the page of the project: the menu on the left (a bar on small screens),
// the header with the breadcrumbs, and the routed page
const LayoutShell: React.FC = () => <>
    <CssBaseline enableColorScheme={true}/>
    <Box sx={{display: 'flex'}}>
        <MainMenu/>
        <AppNavbar/>
        <Box component="main" sx={{flexGrow: 1, overflow: 'auto'}}>
            <Stack spacing={2} sx={{alignItems: 'center', mx: 3, pb: 5, mt: {xs: 8, md: 0}}}>
                <Header/>
                <Outlet/>
            </Stack>
        </Box>
    </Box>
</>;

export default LayoutShell;
