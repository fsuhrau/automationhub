import Paper from "@mui/material/Paper";
import {Box, Typography} from "@mui/material";
import Grid from "@mui/material/Grid2";
import IconButton from "@mui/material/IconButton";
import {Edit} from "@mui/icons-material";
import {TitleCard} from "../../components/title.card.component";
import React, {useState} from "react";
import EditAttributePopup, {EditAttribute} from "./EditAttributePopup";
import {updateProject} from "../../project/project.service";
import IProject from "../../project/project";
import {HubStateActions} from "../../application/HubState";
import {useHubState} from "../../hooks/HubStateProvider";
import {useProjectContext} from "../../hooks/ProjectProvider";


const ProjectGroup: React.FC = () => {

    const {state, dispatch} = useHubState()
    const {project, projectIdentifier} = useProjectContext();

    const [editAttributeState, setEditAttributeState] = useState<EditAttribute>({
        attribute: null,
        value: '',
    });

    const onEditAttributeClose = () => {
        setEditAttributeState(prevState => ({...prevState, attribute: null, value: ''}))
    };

    const onEditAttributeSubmit = (attribute: string, value:string) => {
        updateProject(projectIdentifier as string, {
            ...project,
            [attribute]: value
        } as IProject).then(response => {
            dispatch({
                type: HubStateActions.ProjectAttributeUpdate,
                payload: {
                    projectIdentifier: projectIdentifier,
                    attribute: attribute,
                    value: value
                }
            })
        })
    };

    return (
        <TitleCard title={"Project"}>
            <EditAttributePopup attribute={editAttributeState.attribute} value={editAttributeState.value} onSubmit={onEditAttributeSubmit} onClose={onEditAttributeClose} />
            <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                <Box sx={{padding: 2}}>
                    <Grid container={true} spacing={4}>
                        <Grid size={{xs: 12, md: 1}}>
                            <Typography variant={"body1"} color={"dimgray"}>Project
                                name</Typography>
                        </Grid>
                        <Grid size={{xs: 12, md: 11}} container={true} spacing={2}>
                            {project.Name}<IconButton aria-label="edit" size={'small'}
                                                      onClick={() => setEditAttributeState(prevState => ({
                                                          ...prevState,
                                                          attribute: 'Name',
                                                          value: project.Name,
                                                      }))}><Edit/></IconButton>
                        </Grid>
                        <Grid size={{xs: 12, md: 1}}>
                            <Typography variant={"body1"} color={"dimgray"}>Project ID</Typography>
                        </Grid>
                        <Grid size={{xs: 12, md: 11}} container={true} spacing={2}>
                            {projectIdentifier}{projectIdentifier === "default_project" &&
                            <IconButton aria-label="edit" size={'small'}
                                        onClick={() => setEditAttributeState(prevState => ({
                                            ...prevState,
                                            attribute: 'Identifier',
                                            value: projectIdentifier,
                                        }))}><Edit/></IconButton>}
                        </Grid>
                    </Grid>
                </Box>
            </Paper>
        </TitleCard>
    )
}

export default ProjectGroup;