import React, {useEffect, useState} from 'react';
import {useLocation, useNavigate} from 'react-router-dom';
import {ButtonBase, CardContent, Tooltip, Typography} from '@mui/material';
import CreateProjectDialog from "./create.project.dialog";
import IProject from "./project";
import {createProject, getProjects} from "./project.service";
import CssBaseline from "@mui/material/CssBaseline";
import {styled} from "@mui/material/styles";
import MuiCard from "@mui/material/Card";
import Stack from "@mui/material/Stack";
import TopBar from "../components/TopBar";
import {HubStateActions} from "../application/HubState";
import {Box} from "@mui/system";
import PlatformTypeIcon from "../components/PlatformTypeIcon";
import {useHubState} from "../hooks/HubStateProvider";
import {useError} from "../ErrorProvider";
import Grid from "@mui/material/Grid";

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
    const {setError} = useError()

    const [openNewProjectDialog, setOpenNewProjectDialog] = useState<boolean>(false);

    const handleCreateNewProject = () => {
        setOpenNewProjectDialog(true);
    };

    const onCreateNewProject = (project: IProject) => {
        createProject(project).then(project => {
            navigate(`/project/${project.identifier}`)
        }).catch(ex => {
            setError(ex)
        })
    };

    useEffect(() => {
        if (state.projects === null) {
            getProjects().then(projects => {
                dispatch({type: HubStateActions.ProjectsUpdate, payload: projects})
            }).catch(ex => {
                setError(ex)
            })
        }
    }, [state.projects])

    return (
        <>
            <CssBaseline enableColorScheme/>
            <TopBar/>
            <ProjectContainer justifyContent="space-between">
                <Grid size={12} container={true} spacing={2} alignContent="flex-start" sx={{width: '100%', maxWidth: 1700, mx: 'auto'}}>
                    <Grid size={12}>
                        <Typography variant={"h5"} component="h1">Your Projects</Typography>
                        <Typography variant="body2" color="text.secondary">
                            Pick a project to see its apps, tests and devices.
                        </Typography>
                    </Grid>
                    <Grid container={true} size={12} spacing={2} alignItems={"stretch"} justifyContent={"flex-start"}>
                        <Grid size={{xs: 12, sm: 6, md: 4, lg: 3}}>
                            <Card variant="outlined" onClick={handleCreateNewProject}>
                                <ButtonBase sx={{width: '100%', padding: 2, minHeight: 140, display: 'block', textAlign: 'left'}}>
                                    <Typography component="h2" variant="h6">
                                        Create new Project
                                    </Typography>
                                    <Box
                                        sx={{display: 'flex', flexDirection: 'column', gap: 2}}
                                    >
                                        <CreateProjectDialog open={openNewProjectDialog}
                                                             onClose={() => setOpenNewProjectDialog(false)}
                                                             onSubmit={onCreateNewProject}/>
                                        <CardContent>
                                            <Typography variant="h4" component="span" color="text.secondary">
                                                +
                                            </Typography>
                                        </CardContent>
                                    </Box>
                                </ButtonBase>
                            </Card>
                        </Grid>

                        {
                            state.projects?.map(project => (
                                <Grid key={`projects_${project.id}`} size={{xs: 12, sm: 6, md: 4, lg: 3}}>
                                    <Card variant="outlined" key={`project_${project.identifier}`}
                                          onClick={() => navigate(`/project/${project.identifier}`)}>
                                        <ButtonBase sx={{width: '100%', padding: 2, minHeight: 140, display: 'block', textAlign: 'left'}}>
                                            <Typography component="h2" variant="h6" noWrap>
                                                {project.name}
                                            </Typography>
                                            <Typography variant="body2" color="text.secondary" noWrap sx={{mb: 1}}>
                                                {project.identifier}
                                            </Typography>
                                            <Box
                                                sx={{display: 'flex', flexDirection: 'row', gap: 2}}
                                            >
                                                {
                                                    project.apps.map(app => (
                                                        <Tooltip key={`project_app_${app.id}`} title={app.name}>
                                                            <Box component="span" sx={{display: 'inline-flex'}}>
                                                                <PlatformTypeIcon platformType={app.platform} color="action"/>
                                                            </Box>
                                                        </Tooltip>
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
