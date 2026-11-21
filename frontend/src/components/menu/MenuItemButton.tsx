import * as React from 'react';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Tooltip from '@mui/material/Tooltip';

// railItem centers an item's icon when the menu is collapsed to icons
const railItem = {justifyContent: 'center', px: 1.5} as const;

interface MenuItemButtonProps {
    text: string,
    hint?: string,
    icon: React.ReactNode,
    collapsed?: boolean,
    selected?: boolean,
    onClick?: () => void,
    href?: string,
}

// MenuItemButton is one menu entry; collapsed it shows only its icon and
// names itself in a tooltip and to screen readers.
const MenuItemButton: React.FC<MenuItemButtonProps> = ({text, hint, icon, collapsed, selected, onClick, href}) => {
    const link = href ? {component: 'a' as const, href, target: '_blank', rel: 'noreferrer'} : {};
    return <ListItem disablePadding={true} sx={{display: 'block'}}>
        <Tooltip title={collapsed ? text : ''} placement="right">
            <ListItemButton selected={selected} aria-current={selected ? 'page' : undefined}
                aria-label={collapsed ? text : undefined} onClick={onClick}
                sx={collapsed ? railItem : undefined} {...link}>
                <ListItemIcon sx={collapsed ? {minWidth: 0} : undefined}>{icon}</ListItemIcon>
                {!collapsed && <ListItemText primary={text} secondary={hint}/>}
            </ListItemButton>
        </Tooltip>
    </ListItem>;
};

export default MenuItemButton;
