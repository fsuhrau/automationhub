import React, { useEffect, useState } from 'react';
import Grid from '@mui/material/Grid';
import { Outlet, useNavigate } from 'react-router-dom';
import { Card, CardContent, Link, Typography } from '@mui/material';
import { Box } from '@mui/system';
import IHubStatsData from '../../types/hub.stats';
import { getHubStats } from '../../services/hub.stats.service';
import TestStatusIconComponent from '../../components/test-status-icon.component';
import { prettySize } from '../../types/app';
import { duration } from '../../types/test.protocol';
import { useProjectContext } from "../../project/project.context";

const ProjectMainPage: React.FC = () => {

    const navigate = useNavigate();

    const {projectId} = useProjectContext();

    const [stats, setStats] = useState<IHubStatsData>();

    useEffect(() => {
        if (projectId !== null) {
            getHubStats(projectId as string).then(response => {
                setStats(response.data);
            }).catch(e => {
                console.log(e);
            });
        }
    }, [projectId]);

    return (
        <>
            <Box sx={ { maxWidth: 1200, margin: 'auto', overflow: 'hidden' } }>
                <Grid container={ true } spacing={ 5 }>
                    <Grid item={ true } xs={ 3 }>
                        <Card>
                            <CardContent>
                                <Typography gutterBottom={true} variant="h5" component="div">
                                    Apps
                                </Typography>
                                <Typography variant="h1" color="text.secondary">
                                    { stats?.AppsCount }
                                </Typography>
                            </CardContent>
                        </Card>
                    </Grid>
                    <Grid item={ true } xs={ 3 }>
                        <Card>
                            <CardContent>
                                <Typography gutterBottom={true} variant="h5" component="div">
                                    App Storage
                                </Typography>
                                <Typography variant="h1" color="text.secondary">
                                    { stats && prettySize(stats?.AppsStorageSize) }
                                </Typography>
                            </CardContent>
                        </Card>
                    </Grid>
                    <Grid item={ true } xs={ 3 }>
                        <Card>
                            <CardContent>
                                <Typography gutterBottom={true} variant="h5" component="div">
                                    Devices
                                </Typography>
                                <Typography variant="h1" color="text.secondary">
                                    { stats?.DeviceCount }
                                </Typography>
                            </CardContent>
                        </Card>
                    </Grid>
                    <Grid item={ true } xs={ 3 }>
                        <Card>
                            <CardContent>
                                <Typography gutterBottom={true} variant="h5" component="div">
                                    Booted
                                </Typography>
                                <Typography variant="h1" color="text.secondary">
                                    { stats?.DeviceBooted }
                                </Typography>
                            </CardContent>
                        </Card>
                    </Grid>
                    <Grid item={ true } xs={ 6 }>
                        <Card>
                            <CardContent>
                                <Typography gutterBottom={true} variant="h5" component="div">
                                    Latest Tests
                                </Typography>
                                { stats?.TestsLastProtocols.map((data) => (
                                    <Grid container={ true } key={`test_protocol_${data.ID}`}>
                                        <Grid item={ true } xs={ 1 }>
                                            <TestStatusIconComponent status={ data.TestResult }/>
                                        </Grid>
                                        <Grid item={ true } xs={ 4 }>
                                            { data.Device?.Name }
                                        </Grid>
                                        <Grid item={ true } xs={ true }>
                                            <Link
                                                href={ `/test/0/run/${ data.TestRunID }/${ data.ID }` }
                                                underline="none">
                                                { data.TestName.split('/')[ 1 ] }
                                            </Link>
                                        </Grid>
                                        <Grid item={ true } xs={ 2 }>
                                            { duration(data.CreatedAt, data.EndedAt) }
                                        </Grid>
                                    </Grid>
                                )) }
                            </CardContent>
                        </Card>
                    </Grid>
                    <Grid item={ true } xs={ 6 }>
                        <Card>
                            <CardContent>
                                <Typography gutterBottom={true} variant="h5" component="div">
                                    Failed Tests
                                </Typography>
                                { stats?.TestsLastFailed.map((data) => (
                                    <Grid container={ true } key={`test_failed_${data.ID}`}>
                                        <Grid item={ true } xs={ 1 }>
                                            <TestStatusIconComponent status={ data.TestResult }/>
                                        </Grid>
                                        <Grid item={ true } xs={ 4 }>
                                            { data.Device?.Name }
                                        </Grid>
                                        <Grid item={ true } xs={ true }>
                                            <Link
                                                href={ `/test/0/run/${ data.TestRunID }/${ data.ID }` }
                                                underline="none">
                                                { data.TestName.split('/')[ 1 ] }
                                            </Link>
                                        </Grid>
                                        <Grid item={ true } xs={ 2 }>
                                            { duration(data.CreatedAt, data.EndedAt) }
                                        </Grid>
                                    </Grid>
                                )) }
                            </CardContent>
                        </Card>
                    </Grid>
                </Grid>
            </Box>
        </>
    );
};

export default ProjectMainPage;
