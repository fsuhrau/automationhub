import React, {useEffect, useState} from 'react';
import Paper from '@mui/material/Paper';
import {useNavigate} from 'react-router-dom';
import TableContainer from '@mui/material/TableContainer';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import {deleteAppBundle, getAppBundles} from '../../services/app.service';
import {IAppBinaryData, prettySize} from '../../types/app';
import Moment from 'react-moment';
import {ButtonGroup, Typography} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import IconButton from '@mui/material/IconButton';
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {PlatformType} from "../../types/platform.type.enum";
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import {Box} from "@mui/system";
import Grid from "@mui/material/Grid2";
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
