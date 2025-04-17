import React, {useEffect, useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {ButtonGroup, Typography} from '@mui/material';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {PlatformType} from "../../types/platform.type.enum";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {Box} from "@mui/system";
import Grid from "@mui/material/Grid";
import PlatformTypeIcon from "../../components/PlatformTypeIcon";
import {useError} from "../../ErrorProvider";
import AppBundlesTable from "./AppBundlesTable";

const AppBundlesPage: React.FC = () => {

    const {project, projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {setError} = useError()

    const navigate = useNavigate();

    function newAppClick(): void {
        navigate('new');
    }

    const app = project.Apps.find(a => a.ID === appId);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'App Bundles'}>
                <Grid container={true}>
                    <Grid size={12} container={true} spacing={1} sx={{
                        padding: 1,
                    }}>
                        <Grid>
                            <PlatformTypeIcon platformType={app?.Platform as PlatformType}/>
                        </Grid>
                        <Grid>
                            <Typography variant={"body1"}>{app?.Name}{' / '}{app?.Identifier}</Typography>
                        </Grid>
                        <Grid container={true} justifyContent={"flex-end"}>
                        </Grid>
                    </Grid>
                    <Grid container={true} size={{xs: 12, md: 12}}>
                        <AppBundlesTable appId={appId}></AppBundlesTable>
                    </Grid>
                </Grid>
            </TitleCard>
        </Box>
    );
};

export default AppBundlesPage;
