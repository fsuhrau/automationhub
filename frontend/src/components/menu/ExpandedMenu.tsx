import * as React from 'react';
import {Link as RouterLink} from 'react-router-dom';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Link from '@mui/material/Link';
import List from '@mui/material/List';
import ListSubheader from '@mui/material/ListSubheader';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import HubLogo from '../HubLogo';
import ApplicationSelection from '../ApplicationSelection';
import NavList from './NavList';
import MenuFooter from './MenuFooter';
import SwitchProjectButton from './SwitchProjectButton';
import {MenuViewProps} from './menuView';
import {useHubState} from '../../hooks/HubStateProvider';
import {routes} from '../../lib/routes';

const nameStyle = {fontWeight: 600, lineHeight: 1.3, display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden'} as const;

const ProjectTitle: React.FC<Pick<MenuViewProps, 'project' | 'onNavigate'>> = ({project, onNavigate}) => (
    <Stack direction="row" alignItems="center" spacing={1.5} sx={{px: 2, pt: 'calc(var(--template-frame-height, 0px) + 12px)', pb: 1.5}}>
        <HubLogo sx={{fontSize: 28}}/>
        <Box sx={{minWidth: 0, flexGrow: 1}}>
            <Typography variant="caption" color="text.secondary">Project</Typography>
            {/* up to two lines, so the logo doesn't cut the name short */}
            <Typography variant="subtitle1" title={project.name} sx={nameStyle}>{project.name}</Typography>
        </Box>
        <SwitchProjectButton onNavigate={onNavigate}/>
    </Stack>
);

const NoApps: React.FC<Pick<MenuViewProps, 'project' | 'onNavigate'>> = ({project, onNavigate}) => (
    <Typography variant="body2" color="text.secondary" sx={{px: 1, pb: 1}}>
        This project has no apps yet.{' '}
        <Link component={RouterLink} to={routes.project(project.identifier).settings('apps')} onClick={onNavigate}>
            Add one in Settings
        </Link> to create tests and upload bundles.
    </Typography>
);

const AppSection: React.FC<Pick<MenuViewProps, 'project' | 'appItems' | 'onNavigate'>> = ({project, appItems, onNavigate}) => {
    const {state} = useHubState();
    const subheader = <ListSubheader disableSticky={true} sx={{lineHeight: '28px', px: 1, bgcolor: 'transparent'}}>App</ListSubheader>;
    const noApps = state.apps !== null && state.apps.length === 0;
    return <List dense={true} sx={{p: 1}} subheader={subheader}>
        {noApps ? <NoApps project={project} onNavigate={onNavigate}/> : <>
            <Box sx={{px: 0.5, pb: 1}}><ApplicationSelection/></Box>
            <NavList items={appItems} onNavigate={onNavigate}/>
        </>}
    </List>;
};

// the menu with names
const ExpandedMenu: React.FC<MenuViewProps> = ({project, projectItems, appItems, onNavigate, onToggleCollapsed}) => (
    <Stack sx={{flexGrow: 1, justifyContent: 'space-between', overflowY: 'auto'}}>
        <Box>
            <ProjectTitle project={project} onNavigate={onNavigate}/>
            <Divider/>
            <List dense={true} sx={{p: 1}}>
                <NavList items={projectItems} onNavigate={onNavigate}/>
            </List>
            <Divider/>
            <AppSection project={project} appItems={appItems} onNavigate={onNavigate}/>
        </Box>
        <MenuFooter collapsed={false} onToggleCollapsed={onToggleCollapsed}/>
    </Stack>
);

export default ExpandedMenu;
