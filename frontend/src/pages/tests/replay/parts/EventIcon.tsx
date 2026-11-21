import React from 'react';
import {Box} from '@mui/material';
import {
    CameraAltRounded,
    CheckCircleRounded,
    ErrorRounded,
    PlayCircleOutlineRounded,
    TouchAppRounded,
    WarningRounded,
} from '@mui/icons-material';
import {eventIsError, ReplayEvent} from '../../replayModel';
import {resultTone, Tone} from '../eventView';

const sx = {fontSize: 16, mt: '2px'};

const TONE_ICON: Record<Tone, React.ReactElement> = {
    error: <ErrorRounded color="error" sx={sx}/>,
    warning: <WarningRounded color="warning" sx={sx}/>,
    success: <CheckCircleRounded color="success" sx={sx}/>,
};

const KIND_ICON: Partial<Record<ReplayEvent['kind'], React.ReactElement>> = {
    'test-start': <PlayCircleOutlineRounded color="primary" sx={sx}/>,
    step: <TouchAppRounded color="action" sx={sx}/>,
    screen: <CameraAltRounded color="action" sx={sx}/>,
};

export const EventIcon: React.FC<{ event: ReplayEvent }> = ({event}) => {
    if (event.kind === 'test-end') return TONE_ICON[resultTone(event.result)];
    if (KIND_ICON[event.kind]) return KIND_ICON[event.kind] as React.ReactElement;
    return eventIsError(event) ? TONE_ICON.error : <Box sx={{width: 16}}/>;
};
