import React, {useEffect, useRef} from 'react';
import {Box, Button, Chip, InputAdornment, MenuItem, Select, Stack, TextField, Typography} from '@mui/material';
import {RefreshRounded, SearchRounded} from '@mui/icons-material';
import {red} from '../../../../shared-theme/themePrimitives';
import {logLevel} from '../../deviceControlModel';
import {LOG_LINE_CHOICES, LogsState} from '../useLogs';

const lineColor = (line: string): { error: boolean, color: string } => {
    const level = logLevel(line);
    const error = level === 'E' || level === 'F';
    return {error, color: error ? 'error.main' : level === 'W' ? 'warning.main' : 'text.primary'};
};

const LogLine: React.FC<{ line: string }> = ({line}) => {
    const {error, color} = lineColor(line);
    return <Box sx={t => ({
        whiteSpace: 'pre-wrap', wordBreak: 'break-all', color,
        // the lighter red of the ramp reads on the dark surface
        ...(error ? t.applyStyles('dark', {color: red[300]}) : {}),
    })}>{line}</Box>;
};

const LogFilters: React.FC<{ logs: LogsState }> = ({logs}) => (
    <>
        <TextField size="small" placeholder="Find" value={logs.filter} onChange={e => logs.setFilter(e.target.value)} sx={{width: 180}}
            slotProps={{
                       input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                       htmlInput: {'aria-label': 'Find in the system log'},
                   }}/>
        <Chip size="small" label="Warnings and errors" variant={logs.problemsOnly ? 'filled' : 'outlined'}
            color={logs.problemsOnly ? 'warning' : 'default'} onClick={logs.toggleProblems}/>
    </>
);

const LogLines: React.FC<{ logs: LogsState, all: string[] }> = ({logs, all}) => {
    const box = useRef<HTMLDivElement>(null);
    // the newest lines are at the end; start there after loading
    useEffect(() => {
        if (box.current) box.current.scrollTop = box.current.scrollHeight;
    }, [all]);
    const count = logs.visible.length === all.length ? `${all.length} lines` : `${logs.visible.length} of ${all.length} lines`;
    return (
        <>
            <Typography variant="caption" color="text.secondary">{count}</Typography>
            <Box ref={box} role="log" aria-label="System log" sx={{
                mt: 0.5, p: 1, maxHeight: '50vh', overflow: 'auto', fontFamily: 'monospace', fontSize: 12,
                bgcolor: 'background.default', border: 1, borderColor: 'divider', borderRadius: 1,
            }}>
                {/* lines repeat, and the list is replaced as a whole */}
                {/* eslint-disable-next-line react/no-array-index-key */}
                {logs.visible.map((line, i) => <LogLine key={i} line={line}/>)}
            </Box>
        </>
    );
};

// the system log of the device, loaded on demand and filtered here
const LogTab: React.FC<{ logs: LogsState, busy: boolean }> = ({logs, busy}) => (
    <Box>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true} sx={{mb: 1.5}}>
            <Button variant="outlined" startIcon={<RefreshRounded/>} onClick={logs.load} disabled={busy}>
                {logs.logs ? 'Refresh' : 'Load'}
            </Button>
            <Select size="small" value={logs.lines} onChange={e => logs.setLines(Number(e.target.value))}
                inputProps={{'aria-label': 'Lines'}}>
                {LOG_LINE_CHOICES.map(n => <MenuItem key={n} value={n}>last {n} lines</MenuItem>)}
            </Select>
            {logs.logs && <LogFilters logs={logs}/>}
        </Stack>
        {logs.logs === null
            ? <Typography variant="body2" color="text.secondary">The device's system log (logcat on Android), newest last.</Typography>
            : <LogLines logs={logs} all={logs.logs}/>}
    </Box>
);

export default LogTab;
