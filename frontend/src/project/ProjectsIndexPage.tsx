import React, {useEffect, useState} from 'react';
import {useLocation, useNavigate} from 'react-router-dom';
import {ButtonBase, CardContent, IconButton, Typography} from '@mui/material';
import CreateProjectDialog from "./create.project.dialog";
import IProject from "./project";
import {createProject, getProjects} from "./project.service";
import CssBaseline from "@mui/material/CssBaseline";
import {styled} from "@mui/material/styles";
import MuiCard from "@mui/material/Card";
import Stack from "@mui/material/Stack";
import ColorModeSelect from "../shared-theme/ColorModeSelect";
import {HubStateActions} from "../application/HubState";
import {Box} from "@mui/system";
import PlatformTypeIcon from "../components/PlatformTypeIcon";
import {useHubState} from "../hooks/HubStateProvider";
import Grid from "@mui/material/Grid2";

const Card = styled(MuiCard)(({theme}) => ({
    display: 'flex',
    flexDirection: 'column',
    alignSelf: 'center',
    width: '100%',
    padding: theme.spacing(2),
    gap: theme.spacing(2),
    margin: 'auto',
    boxShadow:
        'hsla(220, 30%, 5%, 0.05) 0px 5px 15px 0px, hsla(220, 25%, 10%, 0.05) 0px 15px 35px -5px',
    ...theme.applyStyles('dark', {
        boxShadow:
            'hsla(220, 30%, 5%, 0.5) 0px 5px 15px 0px, hsla(220, 25%, 10%, 0.08) 0px 15px 35px -5px',
    }),
}));

const ProjectContainer = styled(Stack)(({theme}) => ({
    height: 'calc((1 - var(--template-frame-height, 0)) * 100dvh)',
    minHeight: '100%',
    padding: theme.spacing(2),
    [theme.breakpoints.up('sm')]: {
        padding: theme.spacing(4),
    },
    '&::before': {
        content: '""',
        display: 'block',
        position: 'absolute',
        zIndex: -1,
        inset: 0,
        backgroundImage:
            'radial-gradient(ellipse at 50% 50%, hsl(210, 100%, 97%), hsl(0, 0%, 100%))',
        backgroundRepeat: 'no-repeat',
        ...theme.applyStyles('dark', {
            backgroundImage:
                'radial-gradient(at 50% 50%, hsla(210, 100%, 16%, 0.5), hsl(220, 30%, 5%))',
        }),
    },
}));

const ProjectsIndexPage: React.FC = (props: any) => {

    const {state, dispatch} = useHubState()

    const navigate = useNavigate();
    const location = useLocation()

    const [openNewProjectDialog, setOpenNewProjectDialog] = useState<boolean>(false);

    const handleCreateNewProject = () => {
        setOpenNewProjectDialog(true);
    };

    const onCreateNewProject = (project: IProject) => {
        createProject(project).then(resp => {
            navigate(`/project/${resp.data.Identifier}`)
        }).catch(ex => {
            console.log(ex)
        })
    };

    useEffect(() => {
        if (state.projects === null) {
            getProjects().then(response => {
                dispatch({type: HubStateActions.ProjectsUpdate, payload: response.data})
            })
        }
    }, [state.projects])

    return (
        <>
            <CssBaseline enableColorScheme/>
            <ColorModeSelect sx={{position: 'fixed', top: '1rem', right: '1rem'}}>
            </ColorModeSelect>
            <ProjectContainer justifyContent="space-between">
                <Grid size={12} container={true} spacing={2}>
                    <Grid size={12}>
                        <Typography variant={"h3"}>Your Projects:</Typography>
                    </Grid>
                    <Grid container={true} size={12} alignItems={"center"} justifyContent={"center"} sx={{height: '100%'}}>
                        <Grid size={3}>
                            <Card variant="outlined" onClick={handleCreateNewProject}>
                                <ButtonBase sx={{width: '100%', padding: 2, height: 200, display: 'block'}}>
                                    <Typography
                                        component="h3"
                                        variant="h3"
                                        sx={{width: '100%', fontSize: 'clamp(2rem, 10vw, 2.15rem)'}}
                                    >
                                        Create new Project
                                    </Typography>
                                    <Box
                                        sx={{display: 'flex', flexDirection: 'column', gap: 2}}
                                    >
                                        <CreateProjectDialog open={openNewProjectDialog}
                                                             onClose={() => setOpenNewProjectDialog(false)}
                                                             onSubmit={onCreateNewProject}/>
                                        <CardContent>
                                            <Typography
                                                component="h3"
                                                variant="h3"
                                                sx={{width: '100%', fontSize: 'clamp(1rem, 10vw, 1.15rem)'}}
                                            >
                                                +
                                            </Typography>
                                        </CardContent>
                                    </Box>
                                </ButtonBase>
                            </Card>
                        </Grid>

                        {
                            state.projects?.map(project => (
                                <Grid key={`projects_${project.ID}`} size={3}>
                                    <Card variant="outlined" key={`project_${project.Identifier}`}
                                          onClick={() => navigate(`/project/${project.Identifier}`)}>
                                        <ButtonBase sx={{width: '100%', padding: 2, height: 200, display: 'block'}}>
                                            <Typography
                                                component="h3"
                                                variant="h3"
                                                sx={{width: '100%', fontSize: 'clamp(2rem, 10vw, 2.15rem)'}}
                                            >
                                                {project.Name}
                                            </Typography>
                                            <Typography
                                                component="h3"
                                                variant="h3"
                                                sx={{width: '100%', fontSize: 'clamp(1rem, 10vw, 1.15rem)'}}
                                            >
                                                {project.Identifier}
                                            </Typography>
                                            <Box
                                                sx={{display: 'flex', flexDirection: 'row', gap: 2}}
                                            >
                                                {
                                                    project.Apps.map(app => (
                                                        <IconButton key={`project_app_${app.ID}`}>
                                                            <PlatformTypeIcon platformType={app.Platform}/>
                                                        </IconButton>
                                                    ))
                                                }
                                            </Box>
                                        </ButtonBase>
                                    </Card>
                                </Grid>
                            ))
                        }
                    </Grid>
                </Grid>
            </ProjectContainer>
        </>
    );
};

export default ProjectsIndexPage;
