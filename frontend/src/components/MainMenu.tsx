import * as React from 'react';
import type {} from '@mui/material/themeCssVarsAugmentation';
import Drawer, {drawerClasses} from '@mui/material/Drawer';
import type {Theme} from '@mui/material/styles';
import MenuContent from './MenuContent';
import UserFooter from './menu/UserFooter';
import {useMenuCollapsed} from './menu/useMenuCollapsed';
import {useAuth} from "../hooks/AuthProvider";

const drawerWidth = 240;
// collapsed, the menu is a rail of icons
const railWidth = 64;

const drawerSx = (width: number) => (theme: Theme): object => {
    const grow = theme.transitions.create('width', {duration: theme.transitions.duration.shorter});
    return {
        display: {xs: 'none', md: 'block'},
        width,
        flexShrink: 0,
        boxSizing: 'border-box',
        transition: grow,
        [`& .${drawerClasses.paper}`]: {
            width,
            boxSizing: 'border-box',
            overflowX: 'hidden',
            backgroundColor: 'background.paper',
            transition: grow,
        },
    };
};

// MainMenu is the project menu on the left. It collapses to a rail of icons
// to give the page more room; the choice is remembered in the browser.
export default function MainMenu(): React.ReactElement {
    const {user} = useAuth();
    const {collapsed, toggle, expand} = useMenuCollapsed();

    return (
        <Drawer variant="permanent" sx={drawerSx(collapsed ? railWidth : drawerWidth)}>
            <MenuContent collapsed={collapsed} onToggleCollapsed={toggle} onExpand={expand}/>
            {user && <UserFooter user={user} collapsed={collapsed}/>}
        </Drawer>
    );
}
