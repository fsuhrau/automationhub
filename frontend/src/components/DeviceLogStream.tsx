import React, {useEffect, useMemo, useRef, useState} from 'react';
import {Box, Chip, IconButton, InputAdornment, Stack, TextField, Tooltip, Typography} from '@mui/material';
import {DeleteSweepRounded, PauseRounded, PlayArrowRounded, SearchRounded} from '@mui/icons-material';
import {red} from '../shared-theme/themePrimitives';
import {logStreamUrl} from '../services/device.control.service';
import {filterLogLines, logLevel} from '../pages/devices/deviceControlModel';

// the most lines kept in the browser
const MAX_LINES = 3000;

type StreamState = 'connecting' | 'live' | 'failing';

interface DeviceLogStreamProps {
    projectId: string,
    deviceId: number,
    maxHeight?: number | string,
}

// DeviceLogStream follows the system log of a device (Android logcat) live.
// It needs no lease, so it also works while a test run holds the device.
const DeviceLogStream: React.FC<DeviceLogStreamProps> = ({projectId, deviceId, maxHeight = 420}) => {
    const [lines, setLines] = useState<string[]>([]);
    const [state, setState] = useState<StreamState>('connecting');
    const [failure, setFailure] = useState<string | null>(null);
    const [filter, setFilter] = useState('');
    const [problemsOnly, setProblemsOnly] = useState(false);
    // paused: the shown lines stay, new ones are still collected
    const [frozen, setFrozen] = useState<string[] | null>(null);
    const box = useRef<HTMLDivElement>(null);
    const atEnd = useRef(true);

    useEffect(() => {
        setLines([]);
        setFrozen(null);
        setState('connecting');
        setFailure(null);
        const source = new EventSource(logStreamUrl(projectId, deviceId, 300), {withCredentials: true});
        source.addEventListener('lines', e => {
            const fresh: string[] = JSON.parse((e as MessageEvent).data);
            setState('live');
            setFailure(null);
            setLines(prev => {
                const next = prev.concat(fresh);
                return next.length > MAX_LINES ? next.slice(next.length - MAX_LINES) : next;
            });
        });
        source.addEventListener('failure', e => {
            setState('failing');
            setFailure(JSON.parse((e as MessageEvent).data).message);
        });
        // the browser reconnects by itself after a dropped connection
        source.onerror = () => setState(s => s === 'live' ? 'connecting' : s);
        source.onopen = () => setState(s => s === 'connecting' ? 'live' : s);
        return () => source.close();
    }, [projectId, deviceId]);

    const shown = frozen ?? lines;
    const visible = useMemo(() => filterLogLines(shown, filter, problemsOnly), [shown, filter, problemsOnly]);

    // stay at the end while new lines come, unless scrolled up
    useEffect(() => {
        const el = box.current;
        if (el && atEnd.current) el.scrollTop = el.scrollHeight;
    }, [visible.length]);

    const status = frozen
        ? {label: `Paused · ${lines.length - frozen.length} new`, color: 'default' as const}
        : state === 'live'
            ? {label: 'Live', color: 'success' as const}
            : state === 'failing'
                ? {label: 'No log', color: 'error' as const}
                : {label: 'Connecting…', color: 'default' as const};

    return <Box>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{mb: 1}}>
            <Tooltip title={failure ?? (state === 'live' ? 'New log lines appear as the device writes them' : '')}>
                <Chip size="small" label={status.label} color={status.color} variant="outlined"/>
            </Tooltip>
            <Chip size="small" label="Problems only" color={problemsOnly ? 'error' : 'default'}
                  variant={problemsOnly ? 'filled' : 'outlined'} onClick={() => setProblemsOnly(v => !v)}/>
            <TextField size="small" placeholder="Filter, e.g. Unity" value={filter} onChange={e => setFilter(e.target.value)}
                       sx={{flex: '1 1 160px', minWidth: 140}}
                       slotProps={{
                           input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                           htmlInput: {'aria-label': 'Filter the log'},
                       }}/>
            <Tooltip title={frozen ? 'Show new lines again' : 'Hold the log to read it'}>
                <IconButton size="small" aria-label={frozen ? 'Resume the log' : 'Pause the log'}
                            onClick={() => setFrozen(f => f ? null : lines)}>
                    {frozen ? <PlayArrowRounded fontSize="small"/> : <PauseRounded fontSize="small"/>}
                </IconButton>
            </Tooltip>
            <Tooltip title="Clear the shown lines">
                <IconButton size="small" aria-label="Clear the log" onClick={() => {
                    setLines([]);
                    setFrozen(f => f ? [] : null);
                }}><DeleteSweepRounded fontSize="small"/></IconButton>
            </Tooltip>
        </Stack>
        {failure && <Typography variant="caption" color="error" component="div" sx={{mb: 0.5}}>{failure}</Typography>}
        <Box ref={box} onScroll={e => {
            const el = e.currentTarget;
            atEnd.current = el.scrollHeight - el.scrollTop - el.clientHeight < 24;
        }} sx={{
            p: 1, maxHeight, minHeight: 120, overflow: 'auto', fontFamily: 'monospace', fontSize: 12,
            bgcolor: 'background.default', border: 1, borderColor: 'divider', borderRadius: 1,
        }} role="log" aria-label="Device log" aria-live="off">
            {visible.length === 0
                ? <Typography variant="body2" color="text.secondary">
                    {shown.length === 0 ? (state === 'failing' ? 'The device log could not be read.' : 'Waiting for log lines…') : 'No line matches.'}
                </Typography>
                : visible.map((line, i) => {
                    const level = logLevel(line);
                    const error = level === 'E' || level === 'F';
                    return <Box key={i} sx={t => ({
                        whiteSpace: 'pre-wrap', wordBreak: 'break-all',
                        color: error ? 'error.main' : level === 'W' ? 'warning.main' : 'text.primary',
                        // the lighter red of the ramp reads on the dark surface
                        ...(error ? t.applyStyles('dark', {color: red[300]}) : {}),
                    })}>{line}</Box>;
                })}
        </Box>
    </Box>;
};

export default DeviceLogStream;
