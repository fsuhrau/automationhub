import React from 'react';
import {Box, Card, Typography} from '@mui/material';
import {formatRuntime, LogEntry} from '../logModel';
import {Screenshot} from './Screenshot';

interface ScreenshotsTabProps {
    screens: LogEntry[],
    onViewer: (index: number) => void,
}

// a grid of every screenshot the test and its runner took
export const ScreenshotsTab: React.FC<ScreenshotsTabProps> = ({screens, onViewer}) =>
    <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 2}}>
        {screens.map((s, i) => <Card key={s.id} variant="outlined">
            <Screenshot entry={s} height={360} onClick={() => onViewer(i)}/>
            <Box sx={{p: 1}}>
                <Typography variant="caption" color="text.secondary">
                    {formatRuntime(s.runtime)}{s.phase && ` · ${s.phase}`}
                </Typography>
                {s.message && <Typography variant="body2">{s.message}</Typography>}
            </Box>
        </Card>)}
    </Box>;
