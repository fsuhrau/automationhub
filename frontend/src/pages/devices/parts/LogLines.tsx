import React, {useEffect, useRef} from 'react';
import {Box, Theme, Typography} from '@mui/material';
import {SystemStyleObject} from '@mui/system';
import {red} from '../../../shared-theme/themePrimitives';
import {LogLevel, logLevel} from '../deviceControlModel';

interface LogLinesProps {
    lines: string[],
    emptyText: string,
    maxHeight: number | string,
}

const isError = (level: LogLevel | null): boolean => level === 'E' || level === 'F';

const lineColor = (level: LogLevel | null): string => {
    if (isError(level)) return 'error.main';
    return level === 'W' ? 'warning.main' : 'text.primary';
};

const lineStyle = (level: LogLevel | null) => (t: Theme): SystemStyleObject<Theme> => ({
    whiteSpace: 'pre-wrap',
    wordBreak: 'break-all',
    color: lineColor(level),
    // the lighter red of the ramp reads on the dark surface
    ...(isError(level) ? t.applyStyles('dark', {color: red[300]}) : {}),
});

const AT_END_PX = 24;

// LogLines shows the lines, staying at the end while new ones come unless
// scrolled up
export const LogLines: React.FC<LogLinesProps> = ({lines, emptyText, maxHeight}) => {
    const box = useRef<HTMLDivElement>(null);
    const atEnd = useRef(true);
    useEffect(() => {
        const el = box.current;
        if (el && atEnd.current) el.scrollTop = el.scrollHeight;
    }, [lines.length]);
    return <Box ref={box} role="log" aria-label="Device log" aria-live="off"
        onScroll={e => {
                    const el = e.currentTarget;
                    atEnd.current = el.scrollHeight - el.scrollTop - el.clientHeight < AT_END_PX;
                }}
        sx={{
                    p: 1, maxHeight, minHeight: 120, overflow: 'auto', fontFamily: 'monospace', fontSize: 12,
                    bgcolor: 'background.default', border: 1, borderColor: 'divider', borderRadius: 1,
                }}>
        {lines.length === 0 && <Typography variant="body2" color="text.secondary">{emptyText}</Typography>}
        {/* eslint-disable-next-line react/no-array-index-key -- lines repeat and are only appended */}
        {lines.map((line, i) => <Box key={i} sx={lineStyle(logLevel(line))}>{line}</Box>)}
    </Box>;
};
