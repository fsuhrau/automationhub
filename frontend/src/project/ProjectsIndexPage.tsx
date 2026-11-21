import React, {useState} from 'react';
import {Box} from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import TopBar from '../components/TopBar';
import NewProjectDialog from './parts/NewProjectDialog';
import ProjectsHeader from './parts/ProjectsHeader';
import ProjectsList from './parts/ProjectsList';
import {SEARCH_FROM} from './projectsModel';
import {useProjectsIndex} from './useProjectsIndex';

// ProjectsIndexPage is the start: the projects with their apps, the last
// opened one first, and creating a new one
const ProjectsIndexPage: React.FC = () => {
    const index = useProjectsIndex();
    const [creating, setCreating] = useState(false);
    const openDialog = (): void => setCreating(true);

    return (
        <>
            <CssBaseline enableColorScheme={true}/>
            <TopBar/>
            <Box sx={{maxWidth: 1200, mx: 'auto', px: {xs: 2, sm: 3}, py: 4}}>
                <ProjectsHeader searchable={(index.projects?.length ?? 0) > SEARCH_FROM} search={index.search}
                    onSearch={index.setSearch} onNew={openDialog}/>
                <ProjectsList projects={index.projects} visible={index.visible} last={index.last}
                    error={index.error} onRetry={index.retry} onNew={openDialog}/>
            </Box>
            <NewProjectDialog open={creating} onCreate={index.create} onClose={() => setCreating(false)}/>
        </>
    );
};

export default ProjectsIndexPage;
