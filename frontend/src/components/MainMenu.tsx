import * as React from 'react';
import type {} from '@mui/material/themeCssVarsAugmentation';

import {styled} from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import MuiDrawer, {drawerClasses} from '@mui/material/Drawer';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import ApplicationSelection from './ApplicationSelection';
import MenuContent from './MenuContent';
import OptionsMenu from './OptionsMenu';
import {NavigatorProps} from "./NavigatorProps";
import {useAuth} from "../hooks/AuthProvider";

const drawerWidth = 240;

const Drawer = styled(MuiDrawer)({
    width: drawerWidth,
    flexShrink: 0,
    boxSizing: 'border-box',
    mt: 10,
    [`& .${drawerClasses.paper}`]: {
        width: drawerWidth,
        boxSizing: 'border-box',
    },
});


export default function MainMenu(props: NavigatorProps) {
    const {appState, dispatch} = props;
    const {user} = useAuth()

    return (
        <Drawer
            variant="permanent"
            sx={{
                display: {xs: 'none', md: 'block'},
                [`& .${drawerClasses.paper}`]: {
                    backgroundColor: 'background.paper',
                },
            }}
        >
            <Box
                sx={{
                    display: 'flex',
                    mt: 'calc(var(--template-frame-height, 0px) + 4px)',
                    p: 1.5,
                }}
            >
                <ApplicationSelection appState={appState} dispatch={dispatch}/>
            </Box>
            <Divider/>
            <MenuContent appState={appState} dispatch={dispatch}/>
            {/*<CardAlert/>*/}
            <Stack
                direction="row"
                sx={{
                    p: 2,
                    gap: 1,
                    alignItems: 'center',
                    borderTop: '1px solid',
                    borderColor: 'divider',
                }}
            >
                <Avatar
                    sizes="small"
                    alt={user.name}
                    src={user.avatar}
                    sx={{width: 36, height: 36}}
                />
                <Box sx={{mr: 'auto'}}>
                    <Typography variant="body2" sx={{fontWeight: 500, lineHeight: '16px'}}>
                        {user.name}
                    </Typography>
                    <Typography variant="caption" sx={{color: 'text.secondary'}}>
                        {user.email}
                    </Typography>
                </Box>
                <OptionsMenu/>
            </Stack>
        </Drawer>
    );
}
