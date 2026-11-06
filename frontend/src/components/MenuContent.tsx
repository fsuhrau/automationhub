import * as React from 'react';
import Box from '@mui/material/Box';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import ListSubheader from '@mui/material/ListSubheader';
import Stack from '@mui/material/Stack';
import Divider from '@mui/material/Divider';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Link from '@mui/material/Link';
import DashboardRoundedIcon from '@mui/icons-material/DashboardRounded';
import AnalyticsRoundedIcon from '@mui/icons-material/AnalyticsRounded';
import Inventory2RoundedIcon from '@mui/icons-material/Inventory2Rounded';
import SettingsRoundedIcon from '@mui/icons-material/SettingsRounded';
import InfoRoundedIcon from '@mui/icons-material/InfoRounded';
import DevicesRoundedIcon from "@mui/icons-material/DevicesRounded";
import SwapHorizRoundedIcon from '@mui/icons-material/SwapHorizRounded';
import {Link as RouterLink, useLocation, useNavigate} from "react-router-dom";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import {useSelectedApp} from "../hooks/useSelectedApp";
import ApplicationSelection from "./ApplicationSelection";

interface NavItem {
    text: string,
    ref: string,
    icon: React.ReactElement,
    selected: boolean,
    hint?: string,
}

const NavList: React.FC<{ items: NavItem[], onNavigate?: () => void }> = ({items, onNavigate}) => {
    const navigate = useNavigate();
    return <>
        {items.map(item => (
            <ListItem key={item.text} disablePadding sx={{display: 'block'}}>
                <ListItemButton selected={item.selected} aria-current={item.selected ? 'page' : undefined}
                                onClick={() => {
                                    navigate(item.ref);
                                    onNavigate?.();
                                }}>
                    <ListItemIcon>{item.icon}</ListItemIcon>
                    <ListItemText primary={item.text} secondary={item.hint}/>
                </ListItemButton>
            </ListItem>
        ))}
    </>;
};

// MenuContent is the project navigation. It is split by scope, so it's clear
// what a page belongs to: the project pages (overview, devices, settings) are
// the same for every app, while tests and bundles belong to the app picked in
// the "App" section.
export default function MenuContent({onNavigate}: { onNavigate?: () => void }) {
    const {project, projectIdentifier} = useProjectContext();
    const {state} = useHubState();
    const {app} = useSelectedApp();
    const {pathname} = useLocation();

    const base = `/project/${projectIdentifier}`;
    const appBase = app ? `${base}/app:${app.id}` : null;

    const projectItems: NavItem[] = [
        {text: 'Overview', ref: base, icon: <DashboardRoundedIcon/>, selected: pathname === base || pathname.endsWith('/home')},
        {text: 'Device Pool', ref: `${base}/devices`, icon: <DevicesRoundedIcon/>, selected: pathname.startsWith(`${base}/device`)},
        {text: 'Settings', ref: `${base}/settings`, icon: <SettingsRoundedIcon/>, selected: pathname.startsWith(`${base}/settings`)},
    ];
    const appItems: NavItem[] = appBase ? [
        {text: 'Tests', ref: `${appBase}/tests`, icon: <AnalyticsRoundedIcon/>, selected: pathname.startsWith(`${appBase}/test`)},
        {text: 'Bundles', ref: `${appBase}/bundles`, icon: <Inventory2RoundedIcon/>, selected: pathname.startsWith(`${appBase}/bundles`)},
    ] : [];

    return (
        <Stack sx={{flexGrow: 1, justifyContent: 'space-between', overflowY: 'auto'}}>
            <Box>
                <Stack direction="row" alignItems="center" spacing={1}
                       sx={{px: 2, pt: 'calc(var(--template-frame-height, 0px) + 12px)', pb: 1.5}}>
                    <Box sx={{minWidth: 0, flexGrow: 1}}>
                        <Typography variant="caption" color="text.secondary">Project</Typography>
                        <Typography variant="subtitle1" noWrap sx={{fontWeight: 600, lineHeight: 1.3}}
                                    title={project.name}>
                            {project.name}
                        </Typography>
                    </Box>
                    <Tooltip title="Switch project">
                        <IconButton size="small" component={RouterLink} to="/" aria-label="Switch project"
                                    onClick={onNavigate}>
                            <SwapHorizRoundedIcon fontSize="small"/>
                        </IconButton>
                    </Tooltip>
                </Stack>
                <Divider/>
                <List dense sx={{p: 1}}>
                    <NavList items={projectItems} onNavigate={onNavigate}/>
                </List>
                <Divider/>
                <List dense sx={{p: 1}}
                      subheader={<ListSubheader disableSticky sx={{lineHeight: '28px', px: 1, bgcolor: 'transparent'}}>
                          App
                      </ListSubheader>}>
                    {state.apps !== null && state.apps.length === 0
                        ? <Typography variant="body2" color="text.secondary" sx={{px: 1, pb: 1}}>
                            This project has no apps yet.{' '}
                            <Link component={RouterLink} to={`${base}/settings`} onClick={onNavigate}>
                                Add one in Settings
                            </Link> to create tests and upload bundles.
                        </Typography>
                        : <>
                            <Box sx={{px: 0.5, pb: 1}}>
                                <ApplicationSelection/>
                            </Box>
                            <NavList items={appItems} onNavigate={onNavigate}/>
                        </>}
                </List>
            </Box>

            <List dense sx={{p: 1}}>
                <ListItem disablePadding sx={{display: 'block'}}>
                    <ListItemButton component="a" href="https://github.com/fsuhrau/automationhub" target="_blank"
                                    rel="noreferrer">
                        <ListItemIcon><InfoRoundedIcon/></ListItemIcon>
                        <ListItemText primary="About"/>
                    </ListItemButton>
                </ListItem>
            </List>
        </Stack>
    );
}
