import React, {memo} from 'react';
import {Box, Tooltip, Typography} from '@mui/material';
import {formatDateTime} from '../../../../components/DateTime';
import {eventIsError, formatOffset, ReplayEvent} from '../../replayModel';
import {eventText} from '../eventView';
import {EventIcon} from './EventIcon';

interface EventRowProps {
    event: ReplayEvent,
    // ms since the start of the replay
    offset: number,
    active: boolean,
    future: boolean,
    onSeek: (time: number) => void,
    rowRef?: React.Ref<HTMLDivElement>,
}

const EventBody: React.FC<{ event: ReplayEvent }> = ({event}) => {
    const test = event.kind === 'test-start' || event.kind === 'test-end';
    return <Box sx={{minWidth: 0}}>
        <Typography variant="body2" sx={{
            fontFamily: test ? undefined : 'monospace', fontSize: test ? undefined : 13, fontWeight: test ? 600 : undefined,
            wordBreak: 'break-word', color: eventIsError(event) ? 'error.main' : 'text.primary',
        }}>
            {event.source && event.kind !== 'step' && <Typography component="span" variant="caption" color="text.secondary" sx={{mr: 1}}>
                {event.source}
            </Typography>}
            {eventText(event)}
        </Typography>
        {event.image && <Box component="img" src={event.image} alt={event.text} loading="lazy"
            sx={{mt: 0.5, maxHeight: 120, maxWidth: '100%', borderRadius: 0.5, border: 1, borderColor: 'divider'}}/>}
    </Box>;
};

const isActivation = (e: React.KeyboardEvent): boolean => e.key === 'Enter' || e.key === ' ';

export const EventRow = memo<EventRowProps>(({event, offset, active, future, onSeek, rowRef}) =>
    <Box ref={rowRef} role="button" tabIndex={0} aria-current={active ? 'true' : undefined}
        onClick={() => onSeek(event.time)}
        onKeyDown={e => isActivation(e) && (e.preventDefault(), onSeek(event.time))}
        sx={theme => ({
            display: 'grid', gridTemplateColumns: '64px 18px 1fr', gap: 1, py: 0.5, px: 1, mx: -1,
            cursor: 'pointer', borderRadius: 1, borderLeft: 3,
            borderColor: active ? (theme.vars ?? theme).palette.primary.main : 'transparent',
            bgcolor: active ? (theme.vars ?? theme).palette.action.selected : 'transparent',
            opacity: future && !active ? 0.55 : 1,
            '&:hover': {bgcolor: (theme.vars ?? theme).palette.action.hover},
            // long logs: let the browser skip rows out of view
            contentVisibility: 'auto', containIntrinsicSize: 'auto 26px',
        })}>
        <Tooltip title={formatDateTime(new Date(event.time), 'HH:mm:ss.SSS')} placement="left">
            <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>
                {formatOffset(offset)}
            </Typography>
        </Tooltip>
        <EventIcon event={event}/>
        <EventBody event={event}/>
    </Box>);
