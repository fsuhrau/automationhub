import React, {useEffect, useState} from 'react';
import {useNavigate, useParams} from 'react-router-dom';
import {ButtonBase, CardContent, IconButton, Typography} from '@mui/material';
import {ApplicationProps} from "../application/ApplicationProps";
import CreateProjectDialog from "./create.project.dialog";
import IProject from "./project";
import {createProject, getProjects} from "./project.service";
import CssBaseline from "@mui/material/CssBaseline";
import AppTheme from "../shared-theme/AppTheme";
import {styled} from "@mui/material/styles";
import MuiCard from "@mui/material/Card";
import Stack from "@mui/material/Stack";
import ColorModeSelect from "../shared-theme/ColorModeSelect";
import {ApplicationStateActions} from "../application/ApplicationState";
import {Box} from "@mui/system";
import PlatformTypeIcon from "../components/PlatformTypeIcon";

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
    [theme.breakpoints.up('sm')]: {
        width: '450px',
    },
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

const ProjectsIndexPage: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const {appState, dispatch} = props;

    const navigate = useNavigate();

    let params = useParams();

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
        getProjects().then(response => {
            dispatch({type: ApplicationStateActions.UpdateProjects, payload: response.data})
        })
    }, [])

    return (
        <AppTheme {...props}>
            <CssBaseline enableColorScheme/>
            <ColorModeSelect sx={{position: 'fixed', top: '1rem', right: '1rem'}}>
            </ColorModeSelect>
            <ProjectContainer direction="row" justifyContent="space-between">
                <Typography variant={"h3"}>Your Projects:</Typography>
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
                {
                    appState.projects.map(project => (
                        <Card variant="outlined" key={`project_${project.Identifier}`} onClick={() => navigate(`/project/${project.Identifier}`)}>
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
                    ))
                }
            </ProjectContainer>
        </AppTheme>
    );
};

export default ProjectsIndexPage;
