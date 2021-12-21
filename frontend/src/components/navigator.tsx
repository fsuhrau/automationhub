import React from 'react';
import Divider from '@mui/material/Divider';
import Drawer, { DrawerProps } from '@mui/material/Drawer';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import HomeIcon from '@mui/icons-material/Home';
import DnsRoundedIcon from '@mui/icons-material/DnsRounded';
import PermMediaOutlinedIcon from '@mui/icons-material/PhotoSizeSelectActual';
import PhonelinkSetupIcon from '@mui/icons-material/PhonelinkSetup';
import { DistributiveOmit } from '@mui/types';
import { NavLink } from 'react-router-dom';
import { Box, Link, ListItemButton } from '@mui/material';

const categories = [
    {
        id: 'Test Center',
        children: [
            { id: 'Tests', ref: '/web/tests', icon: <PermMediaOutlinedIcon/> },
            // { id: 'Results', ref: '/web/results', icon: <SettingsEthernetIcon/> },
            // { id: 'Performance', ref: '/web/performance', icon: <TimerIcon/> },
        ],
    },
    {
        id: 'Management',
        children: [
            // { id: 'Settings', ref: '/web/settings', icon: <SettingsIcon/> },
            { id: 'Apps', ref: '/web/apps', icon: <DnsRoundedIcon/> },
            // { id: 'User', ref: '/web/users', icon: <PeopleIcon/> },
            { id: 'Devices', ref: '/web/devices', icon: <PhonelinkSetupIcon/> },
        ],
    },
];

const item = {
    py: '2px',
    px: 3,
    color: 'rgba(255, 255, 255, 0.7)',
    '&:hover, &:focus': {
        bgcolor: 'rgba(255, 255, 255, 0.08)',
    },
};

const itemCategory = {
    boxShadow: '0 -1px 0 rgb(255,255,255,0.1) inset',
    py: 1.5,
    px: 3,
};

export type NavigatorProps = DistributiveOmit<DrawerProps, 'classes'>;

const Navigator: React.FC<NavigatorProps> = (props) => {

    const { ...other } = props;

    return (
        <Drawer variant="permanent" { ...other }>
            <List disablePadding={ true }>
                <ListItem sx={{ ...item, ...itemCategory, fontSize: 22, color: '#fff' }}>
                    Automation Hub
                </ListItem>
                <ListItem sx={{ ...item, ...itemCategory }}>
                    <ListItemIcon>
                        <HomeIcon/>
                    </ListItemIcon>
                    <ListItemText>
                        Project Overview
                    </ListItemText>
                </ListItem>
                { categories.map(({ id, children }) => (
                    <Box key={id} sx={{ bgcolor: '#101F33' }}>
                        <ListItem sx={{ py: 2, px: 3 }} >
                            <ListItemText sx={{ color: '#fff' }}>
                                { id }
                            </ListItemText>
                        </ListItem>
                        { children.map(({ id: childId, ref, icon }) => (
                            <Link
                                key={ childId }
                                component={ NavLink }
                                to={ ref }
                                underline="none"
                            >
                                <ListItem disablePadding={true} key={childId}>
                                    <ListItemButton sx={item}>
                                        <ListItemIcon>{icon}</ListItemIcon>
                                        <ListItemText>{childId}</ListItemText>
                                    </ListItemButton>
                                </ListItem>
                            </Link>
                        )) }
                        <Divider sx={{ mt: 2 }} />
                    </Box>
                )) }
            </List>
        </Drawer>
    );
};

export default Navigator;
