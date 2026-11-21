import React from 'react';
import {Alert, Box} from '@mui/material';
import Grid from '@mui/material/Grid';
import {useParams} from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import {useTestRunService} from '../../services/test.run.service';
import {routes} from '../../lib/routes';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useLoader} from '../../hooks/useLoader';
import PageLoader from '../../components/PageLoader';
import {EventsPanel} from './replay/parts/EventsPanel';
import {PlayerPanel} from './replay/parts/PlayerPanel';
import {ReplayHeader} from './replay/parts/ReplayHeader';
import {useReplayClock} from './replay/useReplayClock';
import {useReplayDevice} from './replay/useReplayDevice';
import {useReplayKeys} from './replay/useReplayKeys';

interface TestReplayPageProps {
    run: ITestRunData,
}

const TestReplayPage: React.FC<TestReplayPageProps> = ({run}) => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const device = useReplayDevice(run);
    const start = device.range?.start ?? 0;
    const end = device.range?.end ?? 0;
    const clock = useReplayClock({start, end}, device.initial);
    useReplayKeys(clock);
    const backTo = routes.project(projectIdentifier).app(appId as number).test(run.testId).run(run.id as number).show();

    return <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
        <ReplayHeader run={run} backTo={backTo}/>
        {!device.range
            ? <Alert severity="info">This run has nothing to replay yet.</Alert>
            : <Grid container={true} spacing={2}>
                <Grid size={{xs: 12, lg: 5}}>
                    <PlayerPanel device={device} clock={clock} start={start} end={end}/>
                </Grid>
                <Grid size={{xs: 12, lg: 7}}>
                    <EventsPanel events={device.events} start={start} position={clock.position}
                        playing={clock.playing} onSeek={clock.seek}/>
                </Grid>
            </Grid>}
    </Box>;
};

export const TestReplayPageLoader: React.FC = () => {
    const service = useTestRunService();
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {testId, runId} = useParams();
    const {data, error, retry} = useLoader(
        () => service.run({projectId: projectIdentifier, appId: appId as number, testId: testId as string, runId: runId as string}),
        [service, projectIdentifier, appId, testId, runId]);
    if (!data || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    return <TestReplayPage key={data.testRun.id} run={data.testRun}/>;
};

export default TestReplayPage;
