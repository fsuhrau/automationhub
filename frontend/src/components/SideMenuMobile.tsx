import * as React from 'react';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import Drawer, {drawerClasses} from '@mui/material/Drawer';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import LogoutRoundedIcon from '@mui/icons-material/LogoutRounded';
import PersonRoundedIcon from '@mui/icons-material/PersonRounded';
import {useNavigate} from 'react-router-dom';

import MenuContent from './MenuContent';
import {useAuth} from "../hooks/AuthProvider";

interface SideMenuMobileProps {
    open: boolean | undefined;
    toggleDrawer: (newOpen: boolean) => () => void;
}

export default function SideMenuMobile({open, toggleDrawer}: SideMenuMobileProps) {
    const auth = useAuth();
    const navigate = useNavigate();
    const close = toggleDrawer(false);
    return (
        <Drawer
            anchor="right"
            open={open}
            onClose={close}
            sx={{
                zIndex: (theme) => theme.zIndex.drawer + 1,
                [`& .${drawerClasses.paper}`]: {
                    backgroundImage: 'none',
                    backgroundColor: 'background.paper',
                },
            }}
        >
            <Stack sx={{width: 280, maxWidth: '85dvw', height: '100%'}}>
                <MenuContent onNavigate={close}/>
                {auth.user && <>
                    <Divider/>
                    <Stack spacing={1} sx={{p: 2}}>
                        <Typography variant="body2" sx={{fontWeight: 500}}>{auth.user.name}</Typography>
                        <Button variant="outlined" fullWidth startIcon={<PersonRoundedIcon/>}
                                onClick={() => {
                                    close();
                                    navigate('/profile');
                                }}>
                            Profile
                        </Button>
                        <Button variant="outlined" fullWidth startIcon={<LogoutRoundedIcon/>}
                                onClick={() => auth.logOut()}>
                            Logout
                        </Button>
                    </Stack>
                </>}
            </Stack>
        </Drawer>
    );
}
