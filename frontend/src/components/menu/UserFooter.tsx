import * as React from 'react';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import OptionsMenu from '../OptionsMenu';
import {IUser} from '../../types/user';

const footerSx = (collapsed: boolean): object => ({
    p: collapsed ? 1 : 2,
    gap: 1,
    alignItems: 'center',
    borderTop: '1px solid',
    borderColor: 'divider',
});

const UserText: React.FC<{ user: IUser }> = ({user}) => (
    <Box sx={{mr: 'auto', minWidth: 0}}>
        <Typography variant="body2" noWrap={true} sx={{fontWeight: 500, lineHeight: '16px'}}>{user.name}</Typography>
        <Typography variant="caption" noWrap={true} component="div" sx={{color: 'text.secondary'}}>{user.email}</Typography>
    </Box>
);

// who is signed in, at the bottom of the menu, with the account menu
const UserFooter: React.FC<{ user: IUser, collapsed: boolean }> = ({user, collapsed}) => (
    <Stack direction={collapsed ? 'column' : 'row'} sx={footerSx(collapsed)}>
        <Tooltip title={collapsed ? `${user.name} · ${user.email}` : ''} placement="right">
            <Avatar sizes="small" alt={user.name} src={user.avatar} sx={{width: 36, height: 36}}/>
        </Tooltip>
        {!collapsed && <UserText user={user}/>}
        <OptionsMenu/>
    </Stack>
);

export default UserFooter;
