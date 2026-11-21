import React from 'react';
import {Alert, Box, Stack} from '@mui/material';
import Grid from '@mui/material/Grid';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {routes} from '../../lib/routes';
import {RunLifecycle} from '../../types/run.lifecycle';
import ITestRunData from '../../types/test.run';
import {deviceCount} from './run/liveDevicesModel';
import {BundleSection, DevicesSection, ParametersSection} from './run/parts/RunDetailSections';
import {ExecutionLog} from './run/parts/ExecutionLog';
import {RunHeader} from './run/parts/RunHeader';
import {DeviceLogSection, LiveScreenSection} from './run/parts/LiveSections';
import {TestsSection} from './run/parts/TestsSection';
import {useLiveDevices} from './run/useLiveDevices';
import {useRunActions} from './run/useRunActions';
import {useRunLinks} from './run/useRunLinks';
import {useRunStream} from './run/useRunStream';
import {useRunView} from './run/useRunView';


interface TestRunPageProps {
    testRun: ITestRunData,
    nextRunId: number | null,
    prevRunId: number | null,
    state?: RunLifecycle,
    // devices the run holds when the page loaded; undefined on older masters
    devices?: number[],
}

const INTERRUPTED = 'Some tests never finished. If the master restarted during the run, their results are lost; run the test again.';

// TestRunPage shows one run. The loader mounts it with key={testRun.id}, so
// all state below belongs to exactly one run.
const TestRunPage: React.FC<TestRunPageProps> = ({testRun, nextRunId, prevRunId, state, devices}) => {
    const {projectIdentifier: projectId} = useProjectContext();
    const {appId} = useApplicationContext();
    const test = {projectId, appId: appId as number, testId: testRun.testId};
    const stream = useRunStream(testRun);
    const view = useRunView(testRun, state, stream);
    const live = useLiveDevices({run: testRun, protocols: stream.protocols, running: view.running, test, initialHeld: devices});
    const links = useRunLinks(testRun, test);
    const actions = useRunActions(testRun, test);
    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <RunHeader run={testRun} view={view} devices={deviceCount(stream.protocols, testRun.deviceStatus)}
                prevRunId={prevRunId} nextRunId={nextRunId} actions={actions} links={links}/>
            {/* three columns on wide screens: the tests, what the run logs, and the devices with the run's details */}
            <Grid container={true} spacing={2}>
                <Grid size={{xs: 12, lg: 4}}>
                    <TestsSection view={view} onOpen={links.openProtocol} onReplay={view.hasReplay ? links.openReplay : undefined}/>
                </Grid>
                <Grid size={{xs: 12, lg: 5}}>
                    <Stack spacing={2}>
                        <ExecutionLog log={stream.log} errors={view.logErrors} running={view.running}/>
                        <DeviceLogSection projectId={projectId} device={live.watched}/>
                    </Stack>
                </Grid>
                <Grid size={{xs: 12, lg: 3}}>
                    <Stack spacing={2}>
                        {view.running && <LiveScreenSection projectId={projectId} live={live}/>}
                        <DevicesSection run={testRun} running={view.running} projectId={projectId}/>
                        <ParametersSection run={testRun}/>
                        <BundleSection run={testRun} bundlesUrl={routes.project(projectId).app(test.appId).bundles()}/>
                    </Stack>
                </Grid>
            </Grid>
            {view.status === 'interrupted' && !view.running && <Alert severity="info" sx={{mt: 2}}>{INTERRUPTED}</Alert>}
        </Box>
    );
};

export default TestRunPage;
