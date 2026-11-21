import React, {useMemo, useState} from 'react';
import {Box, Typography} from '@mui/material';
import {SseSourceFactory} from '../../../hooks/useSseLog';
import {chipHint, emptyLogText, logChip} from '../deviceLogModel';
import {filterLogLines} from '../deviceControlModel';
import {useDeviceLog} from '../useDeviceLog';
import {LogToolbar} from './LogToolbar';
import {LogLines} from './LogLines';

interface DeviceLogPanelProps {
    url: string,
    maxHeight: number | string,
    createSource?: SseSourceFactory,
}

// DeviceLogPanel follows the log stream at url: filter, problems only,
// pause (the shown lines stay, new ones are still collected) and clear
export const DeviceLogPanel: React.FC<DeviceLogPanelProps> = ({url, maxHeight, createSource}) => {
    const log = useDeviceLog(url, createSource);
    const [filter, setFilter] = useState('');
    const [problemsOnly, setProblemsOnly] = useState(false);
    const [frozen, setFrozen] = useState<string[] | null>(null);
    const shown = frozen ?? log.lines;
    const visible = useMemo(() => filterLogLines(shown, filter, problemsOnly), [shown, filter, problemsOnly]);

    const clear = (): void => {
        log.clear();
        setFrozen(f => f ? [] : null);
    };
    return <Box>
        <LogToolbar chip={logChip(log, frozen)} hint={chipHint(log)} paused={frozen !== null} filter={filter}
            problemsOnly={problemsOnly} onFilter={setFilter} onToggleProblems={() => setProblemsOnly(v => !v)}
            onTogglePause={() => setFrozen(f => f ? null : log.lines)} onClear={clear}/>
        {log.failure && <Typography variant="caption" color="error" component="div" sx={{mb: 0.5}}>{log.failure}</Typography>}
        <LogLines lines={visible} emptyText={emptyLogText(shown.length, log.status)} maxHeight={maxHeight}/>
    </Box>;
};
