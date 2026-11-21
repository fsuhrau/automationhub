import * as React from 'react';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import Stack from '@mui/material/Stack';
import Tooltip from '@mui/material/Tooltip';
import AppsRoundedIcon from '@mui/icons-material/AppsRounded';
import HubLogo from '../HubLogo';
import PlatformTypeIcon from '../PlatformTypeIcon';
import MenuItemButton from './MenuItemButton';
import NavList from './NavList';
import MenuFooter from './MenuFooter';
import SwitchProjectButton from './SwitchProjectButton';
import {MenuViewProps} from './menuView';

const AppEntry: React.FC<Pick<MenuViewProps, 'app' | 'onExpand'>> = ({app, onExpand}) => (
    // the app picker needs the room of the open menu
    <MenuItemButton collapsed={true} onClick={onExpand}
        text={app ? `App ${app.name}: open the menu to switch` : 'Pick an app'}
        icon={app ? <PlatformTypeIcon platformType={app.platform}/> : <AppsRoundedIcon/>}/>
);

// the menu as a rail of icons
const CollapsedMenu: React.FC<MenuViewProps> = ({project, app, projectItems, appItems, onNavigate, onExpand, onToggleCollapsed}) => (
    <Stack sx={{flexGrow: 1, justifyContent: 'space-between', overflowY: 'auto', overflowX: 'hidden'}}>
        <Box>
            <Stack alignItems="center" spacing={0.5} sx={{pt: 'calc(var(--template-frame-height, 0px) + 12px)', pb: 1}}>
                <Tooltip title={`Project ${project.name}`} placement="right">
                    <Box sx={{display: 'flex', py: 0.5}}><HubLogo sx={{fontSize: 28}}/></Box>
                </Tooltip>
                <SwitchProjectButton onNavigate={onNavigate} placement="right"/>
            </Stack>
            <Divider/>
            <List dense={true} sx={{p: 1}}>
                <NavList items={projectItems} collapsed={true} onNavigate={onNavigate}/>
            </List>
            <Divider/>
            <List dense={true} sx={{p: 1}} aria-label="App">
                <AppEntry app={app} onExpand={onExpand}/>
                <NavList items={appItems} collapsed={true} onNavigate={onNavigate}/>
            </List>
        </Box>
        <MenuFooter collapsed={true} onToggleCollapsed={onToggleCollapsed}/>
    </Stack>
);

export default CollapsedMenu;
