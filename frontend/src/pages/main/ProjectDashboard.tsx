import React, {useEffect, useState} from 'react';
import Grid from '@mui/material/Grid';
import {Box} from '@mui/system';
import {getHubStats} from '../../services/hub.stats.service';
import {prettySize} from '../../types/app';
import {useProjectContext} from "../../hooks/ProjectProvider";
import StatCard, {StatCardProps} from "../../components/StatCard";
import ProjectDashboardTestResultsDataGrid, {
    ProjectDashboardTestResultsData
} from "../../components/ProjectDashboardTestResultsDataGrid";
import {duration} from "../../types/test.protocol";
import {TitleCard} from "../../components/title.card.component";
import {useError} from "../../ErrorProvider";

const ProjectDashboard: React.FC = () => {

    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const [stats, setStats] = useState<StatCardProps[]>([]);
    const [resultData, setResultData] = useState<ProjectDashboardTestResultsData[]>([]);

    useEffect(() => {
        if (stats.length === 0) {
            getHubStats(projectIdentifier).then(response => {
                setStats([{
                    title: 'Apps',
                    value: response.data.AppsCount.toString(),
                }, {
                    title: 'App Storage',
                    value: prettySize(response.data.AppsStorageSize),
                }, {
                    title: 'Devices',
                    value: response.data.DeviceCount.toString(),
                }, {
                    title: 'Booted',
                    value: response.data.DeviceBooted.toString(),
                },
                ]);
                setResultData(response.data.TestsLastProtocols.map(d => {
                    const testName = d.TestName.split('/');
                    return {
                        id: d.ID!,
                        name: testName.length > 1 ? testName[1] : d.TestName,
                        result: d.TestResult,
                        fps: (+d.AvgFPS).toFixed(0),
                        mem: (+d.AvgMEM).toFixed(0),
                        cpu: (+d.AvgCPU).toFixed(0),
                        time: duration(d.CreatedAt, d.EndedAt),
                    }
                }));
            }).catch(e => {
                setError(e);
            });
        }
    }, []);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'Overview'}>
                <Grid
                    container
                    spacing={2}
                    columns={12}
                    sx={{mb: (theme) => theme.spacing(2)}}
                >
                    {stats.map((card, index) => (
                        <Grid key={index} size={{xs: 12, sm: 6, lg: 3}}>
                            <StatCard {...card} />
                        </Grid>
                    ))}
                    {/*
                <Grid size={{xs: 12, sm: 6, lg: 3}}>
                    <HighlightedCard/>
                </Grid>
                <Grid size={{xs: 12, md: 6}}>
                    <SessionsChart/>
                </Grid>
                <Grid size={{xs: 12, md: 6}}>
                    <PageViewsBarChart/>
                </Grid>
                */}
                </Grid>
            </TitleCard>

            <TitleCard title={'Details'}>
                <Grid container spacing={2} columns={12}>
                    <Grid size={{xs: 12, lg: 9}}>
                        <ProjectDashboardTestResultsDataGrid data={resultData}/>
                    </Grid>
                </Grid>
                <Grid container={true} spacing={5}>
                    {/*
                <Grid item={true} xs={6}>
                    <Card>
                        <CardContent>
                            <Typography gutterBottom={true} variant="h5" component="div">
                                Latest Tests
                            </Typography>
                            {stats?.TestsLastProtocols.map((data) => {
                                const testName = data.TestName.split('/');
                                return (
                                    <Grid container={true} key={`test_protocol_${data.ID}`}>
                                        <Grid item={true} xs={1}>
                                            <TestStatusIconComponent status={data.TestResult}/>
                                        </Grid>
                                        <Grid item={true} xs={true}>
                                            <Link
                                                href={`/project/${projectId}/app/${data.TestRun.Test.AppID}/test/4/run/${data.TestRunID}/${data.ID}`}
                                                underline="none">
                                                {testName.length > 1 ? testName[1] : data.TestName}
                                            </Link> <br/>
                                            {data.Device && (data.Device.Alias.length > 0 ? data.Device.Alias : data.Device.Name)}
                                        </Grid>
                                        <Grid item={true} xs={2}>
                                            {duration(data.CreatedAt, data.EndedAt)}
                                        </Grid>
                                    </Grid>
                                )
                            })}
                        </CardContent>
                    </Card>
                </Grid>
                <Grid item={true} xs={6}>
                    <Card>
                        <CardContent>
                            <Typography gutterBottom={true} variant="h5" component="div">
                                Failed Tests
                            </Typography>
                            {stats?.TestsLastFailed.map((data) => (
                                <Grid container={true} key={`test_failed_${data.ID}`}>
                                    <Grid item={true} xs={1}>
                                        <TestStatusIconComponent status={data.TestResult}/>
                                    </Grid>
                                    <Grid item={true} xs={true}>
                                        <Link
                                            href={`/test/0/run/${data.TestRunID}/${data.ID}`}
                                            underline="none">
                                            {data.TestName.split('/')[1]}
                                        </Link> <br/>
                                        {data.Device && (data.Device.Alias.length > 0 ? data.Device.Alias : data.Device.Name)}
                                    </Grid>
                                    <Grid item={true} xs={2}>
                                        {duration(data.CreatedAt, data.EndedAt)}
                                    </Grid>
                                </Grid>
                            ))}
                        </CardContent>
                    </Card>
                </Grid>
                */}
                </Grid>
            </TitleCard>
        </Box>
    );
};

export default ProjectDashboard;
