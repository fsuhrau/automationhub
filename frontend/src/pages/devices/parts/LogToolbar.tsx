import React from 'react';
import {Chip, IconButton, InputAdornment, Stack, TextField, Tooltip} from '@mui/material';
import {DeleteSweepRounded, PauseRounded, PlayArrowRounded, SearchRounded} from '@mui/icons-material';
import {LogChip} from '../deviceLogModel';

interface LogToolbarProps {
    chip: LogChip,
    hint: string,
    paused: boolean,
    filter: string,
    problemsOnly: boolean,
    onFilter: (text: string) => void,
    onToggleProblems: () => void,
    onTogglePause: () => void,
    onClear: () => void,
}

const FilterField: React.FC<Pick<LogToolbarProps, 'filter' | 'onFilter'>> = ({filter, onFilter}) =>
    <TextField size="small" placeholder="Filter, e.g. Unity" value={filter} onChange={e => onFilter(e.target.value)}
        sx={{flex: '1 1 160px', minWidth: 140}}
        slotProps={{
                   input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                   htmlInput: {'aria-label': 'Filter the log'},
               }}/>;

const PauseButton: React.FC<Pick<LogToolbarProps, 'paused' | 'onTogglePause'>> = ({paused, onTogglePause}) =>
    <Tooltip title={paused ? 'Show new lines again' : 'Hold the log to read it'}>
        <IconButton size="small" aria-label={paused ? 'Resume the log' : 'Pause the log'} onClick={onTogglePause}>
            {paused ? <PlayArrowRounded fontSize="small"/> : <PauseRounded fontSize="small"/>}
        </IconButton>
    </Tooltip>;

// LogToolbar: the state of the stream and what to do with the shown lines
export const LogToolbar: React.FC<LogToolbarProps> = props =>
    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true} sx={{mb: 1}}>
        <Tooltip title={props.hint}>
            <Chip size="small" label={props.chip.label} color={props.chip.color} variant="outlined"/>
        </Tooltip>
        <Chip size="small" label="Problems only" color={props.problemsOnly ? 'error' : 'default'}
            variant={props.problemsOnly ? 'filled' : 'outlined'} onClick={props.onToggleProblems}/>
        <FilterField filter={props.filter} onFilter={props.onFilter}/>
        <PauseButton paused={props.paused} onTogglePause={props.onTogglePause}/>
        <Tooltip title="Clear the shown lines">
            <IconButton size="small" aria-label="Clear the log" onClick={props.onClear}><DeleteSweepRounded fontSize="small"/></IconButton>
        </Tooltip>
    </Stack>;
