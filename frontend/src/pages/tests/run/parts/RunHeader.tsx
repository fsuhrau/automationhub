import React from 'react';
import {Box, Chip, CircularProgress, Stack, Tooltip, Typography} from '@mui/material';
import {formatDateTime} from '../../../../components/DateTime';
import ITestRunData from '../../../../types/test.run';
import {formatDuration} from '../../testRunModel';
import {STATUS} from '../runStatusModel';
import {RunActions} from '../useRunActions';
import {RunLinks} from '../useRunLinks';
import {RunView} from '../useRunView';
import {RunToolbar} from './RunToolbar';

const Meta: React.FC<{ run: ITestRunData, view: RunView, devices: number }> = ({run, view, devices}) =>
    <Typography variant="body2" color="text.secondary" sx={{mt: 0.5}}>
        Started {formatDateTime(run.createdAt, 'YYYY/MM/DD HH:mm')}
        {view.elapsed !== null && <> · {view.running ? 'running for' : 'took'} {formatDuration(view.elapsed)}</>}
        {devices > 0 && <> · {devices} {devices === 1 ? 'device' : 'devices'}</>}
        {run.appBinary && <> · {run.appBinary.name} {run.appBinary.version}</>}
    </Typography>;

const Title: React.FC<{ run: ITestRunData, view: RunView }> = ({run, view}) => {
    const info = STATUS[view.status];
    return <Stack direction="row" alignItems="center" spacing={1.5} flexWrap="wrap" useFlexGap={true}>
        <Typography component="h1" variant="h5">
            {run.test.name} <Typography component="span" variant="h5" color="text.secondary">· Run {run.id}</Typography>
        </Typography>
        <Tooltip title={info.hint}>
            <Chip label={info.label} color={info.color} size="small"
                icon={view.running ? <CircularProgress size={12} color="inherit"/> : undefined}/>
        </Tooltip>
    </Stack>;
};

interface RunHeaderProps {
    run: ITestRunData,
    view: RunView,
    devices: number,
    prevRunId: number | null,
    nextRunId: number | null,
    actions: RunActions,
    links: RunLinks,
}

// what run this is, how it went, and what can be done with it
export const RunHeader: React.FC<RunHeaderProps> = ({run, view, devices, prevRunId, nextRunId, actions, links}) =>
    <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
        <Box sx={{minWidth: 0}}>
            <Title run={run} view={view}/>
            <Meta run={run} view={view} devices={devices}/>
        </Box>
        <RunToolbar view={view} prevRunId={prevRunId} nextRunId={nextRunId} actions={actions} links={links}/>
    </Stack>;
