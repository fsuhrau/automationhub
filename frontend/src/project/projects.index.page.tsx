import React, { useEffect, useState } from 'react';
import Grid from '@mui/material/Grid';
import { useNavigate, useParams } from 'react-router-dom';
import { ButtonBase, Card, CardActions, CardContent, IconButton, Typography } from '@mui/material';
import { Box } from '@mui/system';
import { ApplicationStateActions } from "../application/application.state";
import AndroidRoundedIcon from "@mui/icons-material/AndroidRounded";
import AppleIcon from "@mui/icons-material/Apple";
import { PlatformType } from "../types/platform.type.enum";
import { ApplicationProps } from "../application/application.props";
import CreateProjectDialog from "./create.project.dialog";
import { isBooleanObject } from "util/types";
import IProject from "./project";
import { createProject } from "./project.service";

const styles = {
    cardAction: {
        display: 'block',
        textAlign: 'initial',
    }
}


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

    return (
        <Box sx={ {maxWidth: 1200, margin: 'auto', overflow: 'hidden'} }>
            <Grid container={ true } spacing={ 5 }>
                <Grid item={ true } xs={ 12 }>
                    <Typography variant={ "h3" }>Your Projects:</Typography>
                </Grid>
                <Grid item={ true } xs={ 3 }>
                    <CreateProjectDialog open={openNewProjectDialog} onClose={() => setOpenNewProjectDialog(false) } onSubmit={onCreateNewProject} />
                    <ButtonBase sx={ {display: 'block', width: '100%', backgroundColor: 'transparent'} }
                                onClick={ handleCreateNewProject }>
                        <Card sx={ {height: 200, padding: 2, margin: 'auto', flexDirection: 'column'} }>
                            <CardContent>
                                <Typography variant={ "h5" } color="text.primary" gutterBottom>
                                    +
                                </Typography>
                                <Typography variant={ "h5" } color="text.primary" gutterBottom>
                                    Create new Project
                                </Typography>
                            </CardContent>
                        </Card>
                    </ButtonBase>
                </Grid>
                {
                    appState.projects.map(project => (
                        <Grid key={ `project_${ project.Identifier }` } item={ true } xs={ 3 }>
                            <ButtonBase sx={ {display: 'block', width: '100%'} } onClick={ () => navigate(`/project/${project.Identifier}`) }>
                                <Card sx={ {height: 200, padding: 2, margin: 'auto', flexDirection: 'column'} }>
                                    <CardContent>
                                        <Typography variant={ "h5" } color="text.primary" gutterBottom>
                                            { project.Name }
                                        </Typography>
                                        <Typography sx={ {fontSize: 14} } color="text.secondary" gutterBottom>
                                            { project.Identifier }
                                        </Typography>
                                    </CardContent>
                                    <CardActions>
                                        {
                                            project.Apps.map(app => (
                                                <IconButton key={ `project_app_${ app.ID }` }>
                                                    { app.Platform === PlatformType.Android &&
                                                        <AndroidRoundedIcon/> }
                                                    { app.Platform === PlatformType.iOS && <AppleIcon/> }
                                                </IconButton>
                                            ))
                                        }
                                    </CardActions>
                                </Card>
                            </ButtonBase>
                        </Grid>
                    ))
                }
            </Grid>
        </Box>
    );
};

export default ProjectsIndexPage;
