import React from 'react';
import {Link as RouterLink} from 'react-router-dom';
import {ButtonBase, Card, Chip, IconButton, Stack, Tooltip, Typography} from '@mui/material';
import {SettingsRounded} from '@mui/icons-material';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import {routes} from '../../lib/routes';
import {IAppData} from '../../types/app';
import IProject from '../project';
import {APPS_SHOWN} from '../projectsModel';

const AppNames: React.FC<{ apps: IAppData[] }> = ({apps}) => {
    if (apps.length === 0) return <Typography variant="body2" color="text.secondary">No apps yet</Typography>;
    return <Stack spacing={0.75}>
        {apps.slice(0, APPS_SHOWN).map(a => <Stack key={a.id} direction="row" spacing={1} alignItems="center">
            <PlatformTypeIcon platformType={a.platform} fontSize="small" color="action"/>
            <Typography variant="body2" noWrap={true}>{a.name}</Typography>
        </Stack>)}
        {apps.length > APPS_SHOWN && <Typography variant="caption" color="text.secondary">and {apps.length - APPS_SHOWN} more</Typography>}
    </Stack>;
};

export interface ProjectCardProps {
    project: IProject,
    // the project opened last gets a chip
    isLast: boolean,
}

// ProjectCard opens a project; its settings icon opens the project's settings
const ProjectCard: React.FC<ProjectCardProps> = ({project, isLast}) => {
    const projectRoutes = routes.project(project.identifier);
    return <Card variant="outlined" sx={{height: '100%', position: 'relative'}}>
        <ButtonBase component={RouterLink} to={projectRoutes.overview()}
            sx={{display: 'block', textAlign: 'left', p: 2.5, height: '100%', width: '100%'}}>
            <Stack direction="row" spacing={1} alignItems="center" sx={{pr: 4}}>
                <Typography component="h2" variant="h6" noWrap={true}>{project.name}</Typography>
                {isLast && <Chip size="small" variant="outlined" label="last opened"/>}
            </Stack>
            <Typography variant="body2" color="text.secondary" noWrap={true} sx={{fontFamily: 'monospace', mb: 2}}>{project.identifier}</Typography>
            <AppNames apps={project.apps ?? []}/>
        </ButtonBase>
        <Tooltip title="Settings">
            <IconButton size="small" aria-label={`Settings of ${project.name}`} sx={{position: 'absolute', top: 12, right: 12}}
                component={RouterLink} to={projectRoutes.settings()}>
                <SettingsRounded fontSize="small"/>
            </IconButton>
        </Tooltip>
    </Card>;
};

export default ProjectCard;
