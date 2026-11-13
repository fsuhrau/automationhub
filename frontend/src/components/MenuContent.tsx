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
import KeyboardDoubleArrowLeftRoundedIcon from '@mui/icons-material/KeyboardDoubleArrowLeftRounded';
import KeyboardDoubleArrowRightRoundedIcon from '@mui/icons-material/KeyboardDoubleArrowRightRounded';
import AppsRoundedIcon from '@mui/icons-material/AppsRounded';
import {Link as RouterLink, useLocation, useNavigate} from "react-router-dom";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import {useSelectedApp} from "../hooks/useSelectedApp";
import ApplicationSelection from "./ApplicationSelection";
import HubLogo from "./HubLogo";
import PlatformTypeIcon from "./PlatformTypeIcon";

interface NavItem {
    text: string,
    ref: string,
    icon: React.ReactElement,
    selected: boolean,
    hint?: string,
}

// railItem centers an item's icon when the menu is collapsed to icons
const railItem = {justifyContent: 'center', px: 1.5} as const;

// MenuItemButton is one menu entry; collapsed it shows only its icon and
// names itself in a tooltip and to screen readers.
const MenuItemButton: React.FC<{
    text: string,
    hint?: string,
    icon: React.ReactNode,
    collapsed?: boolean,
    selected?: boolean,
    onClick?: () => void,
    href?: string,
}> = ({text, hint, icon, collapsed, selected, onClick, href}) => {
    const link = href ? {component: 'a' as const, href, target: '_blank', rel: 'noreferrer'} : {};
    return <ListItem disablePadding sx={{display: 'block'}}>
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

const NavList: React.FC<{ items: NavItem[], collapsed?: boolean, onNavigate?: () => void }> = ({items, collapsed, onNavigate}) => {
    const navigate = useNavigate();
    return <>
        {items.map(item => (
            <MenuItemButton key={item.text} text={item.text} hint={item.hint} icon={item.icon}
                            collapsed={collapsed} selected={item.selected}
                            onClick={() => {
                                navigate(item.ref);
                                onNavigate?.();
                            }}/>
        ))}
    </>;
};

// MenuContent is the project navigation. It is split by scope, so it's clear
// what a page belongs to: the project pages (overview, devices, settings) are
// the same for every app, while tests and bundles belong to the app picked in
// the "App" section. Collapsed (onToggleCollapsed is given only where the
// menu can collapse) it is a rail of icons.
export default function MenuContent({onNavigate, collapsed = false, onToggleCollapsed, onExpand}: {
    onNavigate?: () => void,
    collapsed?: boolean,
    onToggleCollapsed?: () => void,
    // opens the collapsed menu, e.g. to pick another app
    onExpand?: () => void,
}) {
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

    const toggleItem = onToggleCollapsed &&
        <MenuItemButton text={collapsed ? 'Expand menu' : 'Collapse menu'} collapsed={collapsed}
                        icon={collapsed ? <KeyboardDoubleArrowRightRoundedIcon/> : <KeyboardDoubleArrowLeftRoundedIcon/>}
                        onClick={onToggleCollapsed}/>;
    const aboutItem = <MenuItemButton text="About" icon={<InfoRoundedIcon/>} collapsed={collapsed}
                                      href="https://github.com/fsuhrau/automationhub"/>;

    if (collapsed) {
        return (
            <Stack sx={{flexGrow: 1, justifyContent: 'space-between', overflowY: 'auto', overflowX: 'hidden'}}>
                <Box>
                    <Stack alignItems="center" spacing={0.5}
                           sx={{pt: 'calc(var(--template-frame-height, 0px) + 12px)', pb: 1}}>
                        <Tooltip title={`Project ${project.name}`} placement="right">
                            <Box sx={{display: 'flex', py: 0.5}}><HubLogo sx={{fontSize: 28}}/></Box>
                        </Tooltip>
                        <Tooltip title="Switch project" placement="right">
                            <IconButton size="small" component={RouterLink} to="/" aria-label="Switch project"
                                        onClick={onNavigate}>
                                <SwapHorizRoundedIcon fontSize="small"/>
                            </IconButton>
                        </Tooltip>
                    </Stack>
                    <Divider/>
                    <List dense sx={{p: 1}}>
                        <NavList items={projectItems} collapsed onNavigate={onNavigate}/>
                    </List>
                    <Divider/>
                    <List dense sx={{p: 1}} aria-label="App">
                        {/* the app picker needs the room of the open menu */}
                        <MenuItemButton collapsed onClick={onExpand}
                                        text={app ? `App ${app.name}: open the menu to switch` : 'Pick an app'}
                                        icon={app ? <PlatformTypeIcon platformType={app.platform}/> : <AppsRoundedIcon/>}/>
                        <NavList items={appItems} collapsed onNavigate={onNavigate}/>
                    </List>
                </Box>
                <List dense sx={{p: 1}}>
                    {aboutItem}
                    {toggleItem}
                </List>
            </Stack>
        );
    }

    return (
        <Stack sx={{flexGrow: 1, justifyContent: 'space-between', overflowY: 'auto'}}>
            <Box>
                <Stack direction="row" alignItems="center" spacing={1.5}
                       sx={{px: 2, pt: 'calc(var(--template-frame-height, 0px) + 12px)', pb: 1.5}}>
                    <HubLogo sx={{fontSize: 28}}/>
                    <Box sx={{minWidth: 0, flexGrow: 1}}>
                        <Typography variant="caption" color="text.secondary">Project</Typography>
                        {/* up to two lines, so the logo doesn't cut the name short */}
                        <Typography variant="subtitle1" title={project.name}
                                    sx={{fontWeight: 600, lineHeight: 1.3, display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden'}}>
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
                            <Link component={RouterLink} to={`${base}/settings?tab=apps`} onClick={onNavigate}>
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
                {aboutItem}
                {toggleItem}
            </List>
        </Stack>
    );
}
