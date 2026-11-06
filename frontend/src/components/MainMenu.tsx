import * as React from 'react';
import type {} from '@mui/material/themeCssVarsAugmentation';

import {styled} from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import MuiDrawer, {drawerClasses} from '@mui/material/Drawer';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import MenuContent from './MenuContent';
import OptionsMenu from './OptionsMenu';
import {useAuth} from "../hooks/AuthProvider";

const drawerWidth = 240;

const Drawer = styled(MuiDrawer)({
    width: drawerWidth,
    flexShrink: 0,
    boxSizing: 'border-box',
    [`& .${drawerClasses.paper}`]: {
        width: drawerWidth,
        boxSizing: 'border-box',
    },
});

export default function MainMenu() {

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
            <MenuContent/>
            {user && <Stack
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
                <Box sx={{mr: 'auto', minWidth: 0}}>
                    <Typography variant="body2" noWrap sx={{fontWeight: 500, lineHeight: '16px'}}>
                        {user.name}
                    </Typography>
                    <Typography variant="caption" noWrap component="div" sx={{color: 'text.secondary'}}>
                        {user.email}
                    </Typography>
                </Box>
                <OptionsMenu/>
            </Stack>}
        </Drawer>
    );
}
