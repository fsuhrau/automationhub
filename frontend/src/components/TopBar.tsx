import * as React from 'react';
import {Link as RouterLink} from 'react-router-dom';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import ButtonBase from '@mui/material/ButtonBase';
import NavbarBreadcrumbs from './NavbarBreadcrumbs';
import OptionsMenu from './OptionsMenu';
import ColorModeIconDropdown from '../shared-theme/ColorModeIconDropdown';
import {CustomIcon} from './AppNavbar';
import {useAuth} from '../hooks/AuthProvider';

// TopBar is the header of the pages outside a project (the project list and
// the profile): the product name leads home, the breadcrumbs say where you
// are, and the user menu reaches the profile and logout from anywhere.
export default function TopBar() {
    const {user} = useAuth();
    return (
        <Box component="header" sx={{
            position: 'sticky', top: 0, zIndex: 2,
            borderBottom: 1, borderColor: 'divider', bgcolor: 'background.paper',
        }}>
            <Stack direction="row" alignItems="center" spacing={2}
                   sx={{maxWidth: 1700, mx: 'auto', px: {xs: 2, sm: 3}, py: 1}}>
                <ButtonBase component={RouterLink} to="/" aria-label="Automation Hub, all projects"
                            sx={{borderRadius: 1, gap: 1, p: 0.5}}>
                    <CustomIcon/>
                    <Typography variant="subtitle1" sx={{fontWeight: 600, display: {xs: 'none', sm: 'block'}}}>
                        Automation Hub
                    </Typography>
                </ButtonBase>
                <Box sx={{flexGrow: 1, minWidth: 0}}>
                    <NavbarBreadcrumbs/>
                </Box>
                <ColorModeIconDropdown/>
                {user && <Stack direction="row" alignItems="center" spacing={1}>
                    <Avatar alt={user.name} src={user.avatar} sx={{width: 32, height: 32}}/>
                    <Typography variant="body2" sx={{display: {xs: 'none', md: 'block'}}}>{user.name}</Typography>
                    <OptionsMenu/>
                </Stack>}
            </Stack>
        </Box>
    );
}
