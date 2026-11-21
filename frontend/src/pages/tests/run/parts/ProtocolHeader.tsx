import React from 'react';
import {Box, Chip, CircularProgress, Link, Stack, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import {formatDateTime} from '../../../../components/DateTime';
import {routes} from '../../../../lib/routes';
import {RunRef} from '../../../../types/test.run.refs';
import {between, formatDuration, TestRow} from '../../testRunModel';
import {deviceLabel, resultLabel, resultPalette} from '../runStatusModel';
import {ProtocolView} from '../useProtocolView';
import {TestNav} from './TestNav';

const DeviceLink: React.FC<{ projectId: string, view: ProtocolView }> = ({projectId, view}) => {
    const {device, deviceId} = view.protocol;
    if (!device) return null;
    return <> · <Link component={RouterLink} to={routes.project(projectId).device(deviceId as number).show()} underline="hover" color="inherit">
        {deviceLabel(device)}</Link>
        {device.osVersion && ` (${device.os} ${device.osVersion})`}</>;
};

const Meta: React.FC<{ run: RunRef, view: ProtocolView }> = ({run, view}) => {
    // the runner times a test as the run does; the test's own start is when
    // the app connected
    const timing = view.runner ?? view.protocol;
    const ms = view.running ? null : between(timing.startedAt, timing.endedAt);
    const runUrl = routes.project(run.projectId).app(run.appId).test(run.testId).run(run.runId).show();
    return <Typography variant="body2" color="text.secondary" sx={{mt: 0.5}}>
        Started {formatDateTime(timing.startedAt)}
        {ms !== null && <> · took {formatDuration(ms)}</>}
        <DeviceLink projectId={run.projectId} view={view}/>
        {' · '}<Link component={RouterLink} to={runUrl} underline="hover" color="inherit">Run {run.runId}</Link>
    </Typography>;
};

const Title: React.FC<{ view: ProtocolView }> = ({view}) =>
    <Stack direction="row" alignItems="center" spacing={1.5} flexWrap="wrap" useFlexGap={true}>
        <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{view.name}</Typography>
        <Chip size="small" label={resultLabel(view.testResult, true)} color={resultPalette(view.testResult)}
            icon={view.running ? <CircularProgress size={12} color="inherit"/> : undefined}/>
    </Stack>;

interface ProtocolHeaderProps {
    run: RunRef,
    view: ProtocolView,
    onOpen: (row: TestRow) => void,
}

// which test, how it went, and the way to the other tests of the run
export const ProtocolHeader: React.FC<ProtocolHeaderProps> = ({run, view, onOpen}) =>
    <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
        <Box sx={{minWidth: 0}}>
            <Title view={view}/>
            {view.suite && <Typography variant="body2" color="text.secondary" sx={{wordBreak: 'break-all'}}>{view.suite}</Typography>}
            <Meta run={run} view={view}/>
        </Box>
        <TestNav place={view.place} onOpen={onOpen}/>
    </Stack>;
