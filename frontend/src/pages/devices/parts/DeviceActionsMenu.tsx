import React, {useState} from 'react';
import {IconButton, ListItemIcon, ListItemText, Menu, MenuItem} from '@mui/material';
import {DeleteOutlineRounded, LockOpenRounded, MoreVertRounded} from '@mui/icons-material';

interface DeviceActionsMenuProps {
    canUnlock: boolean,
    onUnlock: () => void,
    onDelete: () => void,
}

// DeviceActionsMenu is the "more actions" menu of the device page
export const DeviceActionsMenu: React.FC<DeviceActionsMenuProps> = ({canUnlock, onUnlock, onDelete}) => {
    const [anchor, setAnchor] = useState<HTMLElement | null>(null);
    const choose = (action: () => void) => (): void => {
        setAnchor(null);
        action();
    };
    return <>
        <IconButton aria-label="More actions" onClick={e => setAnchor(e.currentTarget)}><MoreVertRounded/></IconButton>
        <Menu anchorEl={anchor} open={!!anchor} onClose={() => setAnchor(null)}>
            <MenuItem disabled={!canUnlock} onClick={choose(onUnlock)}>
                <ListItemIcon><LockOpenRounded fontSize="small"/></ListItemIcon>
                <ListItemText primary="Unlock" secondary="End the test run or control session holding it"/>
            </MenuItem>
            <MenuItem onClick={choose(onDelete)} sx={{color: 'error.main'}}>
                <ListItemIcon><DeleteOutlineRounded fontSize="small" color="error"/></ListItemIcon>
                <ListItemText primary="Delete device"/>
            </MenuItem>
        </Menu>
    </>;
};
