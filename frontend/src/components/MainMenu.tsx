import * as React from 'react';
import type {} from '@mui/material/themeCssVarsAugmentation';

import Avatar from '@mui/material/Avatar';
import Drawer, {drawerClasses} from '@mui/material/Drawer';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import MenuContent from './MenuContent';
import OptionsMenu from './OptionsMenu';
import {useAuth} from "../hooks/AuthProvider";

const drawerWidth = 240;
// collapsed, the menu is a rail of icons
const railWidth = 64;

const STORAGE_KEY = 'mainMenuCollapsed';

const loadCollapsed = (): boolean => {
    try {
        return localStorage.getItem(STORAGE_KEY) === 'true';
    } catch {
        return false;
    }
};

const saveCollapsed = (collapsed: boolean) => {
    try {
        localStorage.setItem(STORAGE_KEY, `${collapsed}`);
    } catch {
        // the menu still collapses, it just isn't remembered
    }
};

// MainMenu is the project menu on the left. It collapses to a rail of icons
// to give the page more room; the choice is remembered in the browser.
export default function MainMenu() {

    const {user} = useAuth()
    const [collapsed, setCollapsed] = React.useState(loadCollapsed);
    const toggle = React.useCallback(() => setCollapsed(c => {
        saveCollapsed(!c);
        return !c;
    }), []);
    const expand = React.useCallback(() => {
        saveCollapsed(false);
        setCollapsed(false);
    }, []);

    const width = collapsed ? railWidth : drawerWidth;

    return (
        <Drawer
            variant="permanent"
            sx={theme => ({
                display: {xs: 'none', md: 'block'},
                width,
                flexShrink: 0,
                boxSizing: 'border-box',
                transition: theme.transitions.create('width', {duration: theme.transitions.duration.shorter}),
                [`& .${drawerClasses.paper}`]: {
                    width,
                    boxSizing: 'border-box',
                    overflowX: 'hidden',
                    backgroundColor: 'background.paper',
                    transition: theme.transitions.create('width', {duration: theme.transitions.duration.shorter}),
                },
            })}
        >
            <MenuContent collapsed={collapsed} onToggleCollapsed={toggle} onExpand={expand}/>
            {user && <Stack
                direction={collapsed ? 'column' : 'row'}
                sx={{
                    p: collapsed ? 1 : 2,
                    gap: 1,
                    alignItems: 'center',
                    borderTop: '1px solid',
                    borderColor: 'divider',
                }}
            >
                <Tooltip title={collapsed ? `${user.name} · ${user.email}` : ''} placement="right">
                    <Avatar
                        sizes="small"
                        alt={user.name}
                        src={user.avatar}
                        sx={{width: 36, height: 36}}
                    />
                </Tooltip>
                {!collapsed && <Box sx={{mr: 'auto', minWidth: 0}}>
                    <Typography variant="body2" noWrap sx={{fontWeight: 500, lineHeight: '16px'}}>
                        {user.name}
                    </Typography>
                    <Typography variant="caption" noWrap component="div" sx={{color: 'text.secondary'}}>
                        {user.email}
                    </Typography>
                </Box>}
                <OptionsMenu/>
            </Stack>}
        </Drawer>
    );
}
