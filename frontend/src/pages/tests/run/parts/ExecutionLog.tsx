import React, {useEffect, useRef, useState} from 'react';
import {Box, Chip, Tooltip, Typography} from '@mui/material';
import {useTheme} from '@mui/material/styles';
// error text needs the lighter red of the ramp to read on the dark surface
import {red} from '../../../../shared-theme/themePrimitives';
import {formatDateTime} from '../../../../components/DateTime';
import {Section} from '../../../../components/common';
import ITesRunLogEntryData from '../../../../types/test.run.log.entry';

const LogLine: React.FC<{ entry: ITesRunLogEntryData }> = ({entry}) => {
    const theme = useTheme();
    const isError = entry.level === 'error';
    return <Box sx={{
        display: 'grid', gridTemplateColumns: '72px 1fr', gap: 1.5, py: 0.5, borderLeft: 3, pl: 1, ml: -1.5,
        borderColor: isError ? (theme.vars ?? theme).palette.error.main : 'transparent',
    }}>
        <Tooltip title={formatDateTime(entry.createdAt)} placement="left">
            <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>
                {formatDateTime(entry.createdAt, 'HH:mm:ss')}
            </Typography>
        </Tooltip>
        <Typography variant="body2" sx={t => ({
            fontFamily: 'monospace', fontSize: 13, wordBreak: 'break-word',
            color: isError ? 'error.main' : 'text.primary',
            ...(isError ? t.applyStyles('dark', {color: red[300]}) : {}),
        })}>
            {entry.log}
        </Typography>
    </Box>;
};

// useStickToEnd keeps a box scrolled to its end while it grows and `active`,
// unless the user scrolled up
const useStickToEnd = (length: number, active: boolean): { box: React.RefObject<HTMLDivElement | null>, onScroll: React.UIEventHandler<HTMLDivElement> } => {
    const box = useRef<HTMLDivElement>(null);
    const atEnd = useRef(true);
    useEffect(() => {
        const el = box.current;
        if (el && active && atEnd.current) el.scrollTop = el.scrollHeight;
    }, [length, active]);
    const onScroll = (e: React.UIEvent<HTMLDivElement>): void => {
        const el = e.currentTarget;
        atEnd.current = el.scrollHeight - el.scrollTop - el.clientHeight < 24;
    };
    return {box, onScroll};
};

interface ExecutionLogProps {
    log: ITesRunLogEntryData[],
    errors: ITesRunLogEntryData[],
    running: boolean,
}

const ErrorsOnlyChip: React.FC<{ count: number, on: boolean, onToggle: () => void }> = ({count, on, onToggle}) =>
    <Chip size="small" color={on ? 'error' : 'default'} variant={on ? 'filled' : 'outlined'}
        label={`Errors only ${count}`} onClick={onToggle}/>;

// what the run wrote to its log, errors highlighted
export const ExecutionLog: React.FC<ExecutionLogProps> = ({log, errors, running}) => {
    const [errorsOnly, setErrorsOnly] = useState(false);
    const {box, onScroll} = useStickToEnd(log.length, running);
    return <Section title="Execution log" fullHeight={true} action={errors.length > 0 &&
        <ErrorsOnlyChip count={errors.length} on={errorsOnly} onToggle={() => setErrorsOnly(v => !v)}/>}>
        {log.length === 0
            ? <Typography variant="body2" color="text.secondary">No log yet.</Typography>
            : <Box ref={box} onScroll={onScroll} role="log" aria-label="Execution log"
                    sx={{maxHeight: {xs: 420, lg: running ? '45vh' : '75vh'}, overflowY: 'auto', mx: -2, px: 2}}>
                {(errorsOnly ? errors : log).map(entry => <LogLine key={entry.id} entry={entry}/>)}
            </Box>}
    </Section>;
};
