import React from 'react';
import {Button, Chip, Stack, Tooltip, Typography} from '@mui/material';
import {
    ArrowBackRounded, ArrowForwardRounded, FiberManualRecordRounded, ReplayRounded, SlowMotionVideoRounded, StopCircleRounded,
} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {RunActions} from '../useRunActions';
import {RunLinks} from '../useRunLinks';
import {RunView} from '../useRunView';

interface RunNavProps {
    prevRunId: number | null,
    nextRunId: number | null,
    onOpen: (runId: number) => void,
}

// the way to the runs of the test before and after this one
const RunNav: React.FC<RunNavProps> = ({prevRunId, nextRunId, onOpen}) => {
    if (!prevRunId && !nextRunId) return null;
    return <Stack direction="row" spacing={0.5} alignItems="center" role="navigation" aria-label="Runs of this test">
        {prevRunId ? <Tooltip title="Previous run of this test">
            <Button size="small" variant="text" startIcon={<ArrowBackRounded/>} onClick={() => onOpen(prevRunId)}>Run {prevRunId}</Button>
        </Tooltip> : null}
        {nextRunId
            ? <Tooltip title="Next run of this test">
                <Button size="small" variant="text" endIcon={<ArrowForwardRounded/>} onClick={() => onOpen(nextRunId)}>Run {nextRunId}</Button>
            </Tooltip>
            : <Typography variant="caption" color="text.secondary" sx={{px: 1}}>Latest run</Typography>}
    </Stack>;
};

const CancelButton: React.FC<{ actions: RunActions }> = ({actions}) =>
    <Button size="small" variant="outlined" color="error" startIcon={<StopCircleRounded/>}
        disabled={actions.cancelling} onClick={actions.cancel}>
        {actions.cancelling ? 'Cancelling…' : 'Cancel run'}
    </Button>;

const RecordingChip: React.FC = () =>
    <Tooltip title="The screens of the devices are recorded for the replay">
        <Chip size="small" color="error" variant="outlined" label="Recording" icon={<FiberManualRecordRounded sx={{fontSize: 12}}/>}/>
    </Tooltip>;

const ReplayButton: React.FC<{ to: string }> = ({to}) =>
    <Tooltip title="Watch the recording next to the log and the test steps">
        <Button size="small" variant="outlined" startIcon={<SlowMotionVideoRounded/>} component={RouterLink} to={to}>Replay</Button>
    </Tooltip>;

interface RunToolbarProps {
    view: RunView,
    prevRunId: number | null,
    nextRunId: number | null,
    actions: RunActions,
    links: RunLinks,
}

export const RunToolbar: React.FC<RunToolbarProps> = ({view, prevRunId, nextRunId, actions, links}) =>
    <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {md: 'auto'}, flexWrap: 'wrap'}} useFlexGap={true}>
        <RunNav prevRunId={prevRunId} nextRunId={nextRunId} onOpen={links.openRun}/>
        {view.running && <CancelButton actions={actions}/>}
        {view.running && view.recording && <RecordingChip/>}
        {view.hasReplay && <ReplayButton to={links.replayUrl()}/>}
        <Tooltip title="Start this test again with the same bundle and parameters">
            <Button size="small" variant="contained" startIcon={<ReplayRounded/>} onClick={actions.rerun}>Run again</Button>
        </Tooltip>
    </Stack>;
