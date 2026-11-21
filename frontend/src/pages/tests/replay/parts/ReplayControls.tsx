import React from 'react';
import {IconButton, MenuItem, Select, Stack, Tooltip, Typography} from '@mui/material';
import {Forward5Rounded, PauseRounded, PlayArrowRounded, Replay5Rounded, SkipNextRounded} from '@mui/icons-material';
import {formatOffset, ReplayEvent} from '../../replayModel';
import {eventText} from '../eventView';
import {RATES, ReplayClock} from '../useReplayClock';

const SKIP_MS = 5000;

interface ReplayControlsProps {
    clock: ReplayClock,
    start: number,
    end: number,
    nextMark: ReplayEvent | undefined,
}

const nextTitle = (nextMark: ReplayEvent | undefined): string =>
    nextMark ? `Next test or error: ${eventText(nextMark)}` : 'No test or error after this point';

export const ReplayControls: React.FC<ReplayControlsProps> = ({clock, start, end, nextMark}) => {
    const {position, playing, rate, setRate, seek, toggle} = clock;
    return <Stack direction="row" alignItems="center" spacing={0.5} flexWrap="wrap" useFlexGap={true}>
        <Tooltip title={playing ? 'Pause (space)' : 'Play (space)'}>
            <IconButton onClick={toggle} aria-label={playing ? 'Pause' : 'Play'} color="primary">
                {playing ? <PauseRounded/> : <PlayArrowRounded/>}
            </IconButton>
        </Tooltip>
        <Tooltip title="Back 5 s (←)">
            <IconButton size="small" onClick={() => seek(position - SKIP_MS)} aria-label="Back 5 seconds"><Replay5Rounded/></IconButton>
        </Tooltip>
        <Tooltip title="Ahead 5 s (→)">
            <IconButton size="small" onClick={() => seek(position + SKIP_MS)} aria-label="Ahead 5 seconds"><Forward5Rounded/></IconButton>
        </Tooltip>
        <Tooltip title={nextTitle(nextMark)}>
            <span>
                <IconButton size="small" disabled={!nextMark} onClick={() => nextMark && seek(nextMark.time)}
                    aria-label="Next test or error"><SkipNextRounded/></IconButton>
            </span>
        </Tooltip>
        <Typography variant="body2" sx={{fontFamily: 'monospace', mx: 1}}>
            {formatOffset(position - start)} / {formatOffset(end - start)}
        </Typography>
        <Select size="small" value={rate} onChange={e => setRate(Number(e.target.value))}
            inputProps={{'aria-label': 'Speed'}} sx={{ml: 'auto', minWidth: 80}}>
            {RATES.map(r => <MenuItem key={r} value={r}>{r}×</MenuItem>)}
        </Select>
    </Stack>;
};
