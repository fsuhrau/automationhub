import React from 'react';
import {Alert, Button, Card, LinearProgress, Typography} from '@mui/material';
import Grid from '@mui/material/Grid';
import {AddRounded} from '@mui/icons-material';
import IProject from '../project';
import ProjectCard from './ProjectCard';

export interface ProjectsListProps {
    // null while the projects load
    projects: IProject[] | null,
    // the projects matching the search
    visible: IProject[],
    last: string | null,
    error: string | null,
    onRetry: () => void,
    onNew: () => void,
}

const NoProjects: React.FC<{ onNew: () => void }> = ({onNew}) =>
    <Card variant="outlined" sx={{p: 4, textAlign: 'center'}}>
        <Typography variant="subtitle1" gutterBottom={true}>No projects yet</Typography>
        <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
            Create one for your game; then add its apps and tests.
        </Typography>
        <Button variant="contained" startIcon={<AddRounded/>} onClick={onNew}>New project</Button>
    </Card>;

// ProjectsList is the project cards, or what stands in for them: the error,
// the loading bar, or a note that there is nothing to show
const ProjectsList: React.FC<ProjectsListProps> = ({projects, visible, last, error, onRetry, onNew}) => {
    if (error) return <Alert severity="error" action={<Button color="inherit" size="small" onClick={onRetry}>Retry</Button>}>{error}</Alert>;
    if (projects === null) return <LinearProgress/>;
    if (projects.length === 0) return <NoProjects onNew={onNew}/>;
    if (visible.length === 0) return <Typography variant="body2" color="text.secondary">No project matches.</Typography>;
    return <Grid container={true} spacing={2}>
        {visible.map(p => <Grid key={p.id} size={{xs: 12, sm: 6, md: 4}}>
            <ProjectCard project={p} isLast={p.identifier === last}/>
        </Grid>)}
    </Grid>;
};

export default ProjectsList;
