import * as React from 'react';
import {styled} from '@mui/material/styles';
import Divider, {dividerClasses} from '@mui/material/Divider';
import Menu from '@mui/material/Menu';
import MuiMenuItem from '@mui/material/MenuItem';
import {paperClasses} from '@mui/material/Paper';
import {listClasses} from '@mui/material/List';
import ListItemText from '@mui/material/ListItemText';
import ListItemIcon, {listItemIconClasses} from '@mui/material/ListItemIcon';
import LogoutRoundedIcon from '@mui/icons-material/LogoutRounded';
import MoreVertRoundedIcon from '@mui/icons-material/MoreVertRounded';
import MenuButton from './MenuButton';
import {useAuth} from "../hooks/AuthProvider";
import {useNavigate} from "react-router-dom";
import {routes} from "../lib/routes";

const MenuItem = styled(MuiMenuItem)({
    margin: '2px 0',
});

const menuSx = {
    [`& .${listClasses.root}`]: {padding: '4px'},
    [`& .${paperClasses.root}`]: {padding: 0},
    [`& .${dividerClasses.root}`]: {margin: '4px -4px'},
};

const logoutSx = {[`& .${listItemIconClasses.root}`]: {ml: 'auto', minWidth: 0}};

export default function OptionsMenu(): React.ReactElement {
    const auth = useAuth();
    const navigate = useNavigate();
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
    const close = (): void => setAnchorEl(null);

    return (
        <React.Fragment>
            <MenuButton aria-label="Account menu" onClick={e => setAnchorEl(e.currentTarget)} sx={{borderColor: 'transparent'}}>
                <MoreVertRoundedIcon/>
            </MenuButton>
            <Menu anchorEl={anchorEl} id="menu" open={Boolean(anchorEl)} onClose={close} onClick={close}
                transformOrigin={{horizontal: 'right', vertical: 'top'}}
                anchorOrigin={{horizontal: 'right', vertical: 'bottom'}} sx={menuSx}>
                <MenuItem onClick={() => navigate(routes.profile())}>Profile and access tokens</MenuItem>
                <MenuItem onClick={() => navigate(routes.projects())}>All projects</MenuItem>
                <Divider/>
                <MenuItem onClick={() => auth.logOut()} sx={logoutSx}>
                    <ListItemText>Logout</ListItemText>
                    <ListItemIcon>
                        <LogoutRoundedIcon fontSize="small"/>
                    </ListItemIcon>
                </MenuItem>
            </Menu>
        </React.Fragment>
    );
}
