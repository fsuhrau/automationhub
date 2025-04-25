import React, {useEffect, useState} from 'react';
import Grid from '@mui/material/Grid';
import {Box} from '@mui/system';
import {getHubStats} from '../../services/hub.stats.service';
import {useProjectContext} from "../../hooks/ProjectProvider";
import StatCard, {StatCardProps} from "../../components/StatCard";
import ProjectDashboardTestResultsDataGrid, {
    ProjectDashboardTestResultsData
} from "../../components/ProjectDashboardTestResultsDataGrid";
import {duration} from "../../types/test.protocol";
import {TitleCard} from "../../components/title.card.component";
import {useError} from "../../ErrorProvider";
import {byteFormat} from "../tests/value_formatter";

const ProjectDashboard: React.FC = () => {

    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const [stats, setStats] = useState<StatCardProps[]>([]);
    const [resultData, setResultData] = useState<ProjectDashboardTestResultsData[]>([]);

    useEffect(() => {
        if (stats.length === 0) {
            getHubStats(projectIdentifier).then(state => {
                setStats([{
                    title: 'Apps',
                    value: state.appsCount.toString(),
                }, {
                    title: 'App Storage',
                    value: byteFormat(state.appsStorageSize),
                }, {
                    title: 'Devices',
                    value: state.deviceCount.toString(),
                }, {
                    title: 'Booted',
                    value: state.deviceBooted.toString(),
                },
                ]);
                setResultData(state.testsLastProtocols.map(d => {
                    const testName = d.testName.split('/');
                    return {
                        id: d.id!,
                        name: testName.length > 1 ? testName[1] : d.testName,
                        result: d.testResult,
                        fps: (+d.avgFps).toFixed(0),
                        mem: (+d.avgMem).toFixed(0),
                        cpu: (+d.avgCpu).toFixed(0),
                        time: duration(d.createdAt, d.endedAt),
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
                            {stats?.testsLastProtocols.map((data) => {
                                const testName = data.testName.split('/');
                                return (
                                    <Grid container={true} key={`test_protocol_${data.id}`}>
                                        <Grid item={true} xs={1}>
                                            <TestStatusIconComponent status={data.testResult}/>
                                        </Grid>
                                        <Grid item={true} xs={true}>
                                            <Link
                                                href={`/project/${projectId}/app/${data.testRun.test.appId}/test/4/run/${data.testRunId}/${data.id}`}
                                                underline="none">
                                                {testName.length > 1 ? testName[1] : data.testName}
                                            </Link> <br/>
                                            {data.device && (data.device.alias.length > 0 ? data.device.alias : data.device.name)}
                                        </Grid>
                                        <Grid item={true} xs={2}>
                                            {duration(data.createdAt, data.endedAt)}
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
                            {stats?.testsLastFailed.map((data) => (
                                <Grid container={true} key={`test_failed_${data.id}`}>
                                    <Grid item={true} xs={1}>
                                        <TestStatusIconComponent status={data.testResult}/>
                                    </Grid>
                                    <Grid item={true} xs={true}>
                                        <Link
                                            href={`/test/0/run/${data.testRunId}/${data.id}`}
                                            underline="none">
                                            {data.testName.split('/')[1]}
                                        </Link> <br/>
                                        {data.device && (data.device.alias.length > 0 ? data.device.alias : data.device.name)}
                                    </Grid>
                                    <Grid item={true} xs={2}>
                                        {duration(data.createdAt, data.endedAt)}
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
