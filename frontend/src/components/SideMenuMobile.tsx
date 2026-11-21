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
import {routes} from "../lib/routes";

interface SideMenuMobileProps {
    open: boolean | undefined;
    toggleDrawer: (newOpen: boolean) => () => void;
}

const drawerSx = {
    zIndex: (theme: { zIndex: { drawer: number } }) => theme.zIndex.drawer + 1,
    [`& .${drawerClasses.paper}`]: {backgroundImage: 'none', backgroundColor: 'background.paper'},
};

// who is signed in, with the profile and logout
const UserActions: React.FC<{ close: () => void }> = ({close}) => {
    const auth = useAuth();
    const navigate = useNavigate();
    if (!auth.user) return null;
    const openProfile = (): void => {
        close();
        navigate(routes.profile());
    };
    return <>
        <Divider/>
        <Stack spacing={1} sx={{p: 2}}>
            <Typography variant="body2" sx={{fontWeight: 500}}>{auth.user.name}</Typography>
            <Button variant="outlined" fullWidth={true} startIcon={<PersonRoundedIcon/>} onClick={openProfile}>
                Profile
            </Button>
            <Button variant="outlined" fullWidth={true} startIcon={<LogoutRoundedIcon/>} onClick={() => auth.logOut()}>
                Logout
            </Button>
        </Stack>
    </>;
};

export default function SideMenuMobile({open, toggleDrawer}: SideMenuMobileProps): React.ReactElement {
    const close = toggleDrawer(false);
    return (
        <Drawer anchor="right" open={open} onClose={close} sx={drawerSx}>
            <Stack sx={{width: 280, maxWidth: '85dvw', height: '100%'}}>
                <MenuContent onNavigate={close}/>
                <UserActions close={close}/>
            </Stack>
        </Drawer>
    );
}
