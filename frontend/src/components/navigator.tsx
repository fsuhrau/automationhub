import React, { FC } from 'react';
import clsx from 'clsx';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import Divider from '@material-ui/core/Divider';
import Drawer, { DrawerProps } from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import HomeIcon from '@material-ui/icons/Home';
import PeopleIcon from '@material-ui/icons/People';
import DnsRoundedIcon from '@material-ui/icons/DnsRounded';
import PermMediaOutlinedIcon from '@material-ui/icons/PhotoSizeSelectActual';
import SettingsEthernetIcon from '@material-ui/icons/SettingsEthernet';
import TimerIcon from '@material-ui/icons/Timer';
import SettingsIcon from '@material-ui/icons/Settings';
import PhonelinkSetupIcon from '@material-ui/icons/PhonelinkSetup';
import { Omit } from '@material-ui/types';
import { NavLink } from 'react-router-dom';
import { Link } from '@material-ui/core';

const categories = [
    {
        id: 'Test Center',
        children: [
            { id: 'Tests', ref: '/tests', icon: <PermMediaOutlinedIcon/> },
            { id: 'Results', ref: '/results', icon: <SettingsEthernetIcon/> },
            { id: 'Performance', ref: '/performance', icon: <TimerIcon/> },
        ],
    },
    {
        id: 'Management',
        children: [
            { id: 'Settings', ref: '/settings', icon: <SettingsIcon/> },
            { id: 'Apps', ref: '/apps', icon: <DnsRoundedIcon/> },
            { id: 'User', ref: '/users', icon: <PeopleIcon/> },
            { id: 'Devices', ref: '/devices', icon: <PhonelinkSetupIcon/> },
        ],
    },
];

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        categoryHeader: {
            paddingTop: theme.spacing(2),
            paddingBottom: theme.spacing(2),
        },
        categoryHeaderPrimary: {
            color: theme.palette.common.white,
        },
        item: {
            paddingTop: 1,
            paddingBottom: 1,
            color: 'rgba(255, 255, 255, 0.7)',
            textDecoration: 'none',
            '&:hover,&:focus': {
                backgroundColor: 'rgba(255, 255, 255, 0.08)',
            },
        },
        itemCategory: {
            backgroundColor: '#232f3e',
            boxShadow: '0 -1px 0 #404854 inset',
            paddingTop: theme.spacing(2),
            paddingBottom: theme.spacing(2),
        },
        firebase: {
            fontSize: 24,
            color: theme.palette.common.white,
        },
        activeLink: {
            '& $item': {
                color: '#4fc3f7',
            },
        },
        itemPrimary: {
            fontSize: 'inherit',
        },
        itemIcon: {
            minWidth: 'auto',
            marginRight: theme.spacing(2),
        },
        divider: {
            marginTop: theme.spacing(2),
        },
    });

export interface NavigatorProps extends Omit<DrawerProps, 'classes'>, WithStyles<typeof styles> {
}

const Navigator: FC<NavigatorProps> = (props) => {
    const { classes, ...other } = props;

    return (
        <Drawer variant="permanent" {...other}>
            <List disablePadding={true}>
                <ListItem className={clsx(classes.firebase, classes.item, classes.itemCategory)}>
                    Automation Hub
                </ListItem>
                <ListItem className={clsx(classes.item, classes.itemCategory)}>
                    <ListItemIcon className={classes.itemIcon}>
                        <HomeIcon/>
                    </ListItemIcon>
                    <ListItemText
                        classes={{
                            primary: classes.itemPrimary,
                        }}
                    >
                        Project Overview
                    </ListItemText>
                </ListItem>
                {categories.map(({ id, children }) => (
                    <React.Fragment key={id}>
                        <ListItem className={classes.categoryHeader}>
                            <ListItemText
                                classes={{
                                    primary: classes.categoryHeaderPrimary,
                                }}
                            >
                                {id}
                            </ListItemText>
                        </ListItem>
                        {children.map(({ id: childId, ref, icon }) => (
                            <Link
                                key={childId}
                                component={NavLink}
                                to={ref}
                                activeClassName={classes.activeLink}
                                underline="none"
                            >
                                <ListItem
                                    key={childId}
                                    button={true}
                                    className={classes.item}
                                >
                                    <ListItemIcon className={classes.itemIcon}>{icon}</ListItemIcon>
                                    <ListItemText
                                        classes={{
                                            primary: classes.itemPrimary,
                                        }}
                                    >
                                        {childId}
                                    </ListItemText>
                                </ListItem>
                            </Link>
                        ))}
                        <Divider className={classes.divider}/>
                    </React.Fragment>
                ))}
            </List>
        </Drawer>
    );
};

export default withStyles(styles)(Navigator);
