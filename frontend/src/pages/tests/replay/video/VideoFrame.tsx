import React from 'react';
import {Box, Typography} from '@mui/material';

export const VideoFrame: React.FC<{ width: number, height: number, children: React.ReactNode }> = ({width, height, children}) =>
    <Box sx={{
        position: 'relative', bgcolor: '#000', borderRadius: 1, overflow: 'hidden', mx: 'auto',
        // portrait phones would be very tall; keep the whole screen in view
        maxHeight: '70vh', aspectRatio: width > 0 && height > 0 ? `${width} / ${height}` : '9 / 16', maxWidth: '100%',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        '& > video, & > img': {width: '100%', height: '100%', objectFit: 'contain', display: 'block'},
    }}>{children}</Box>;

export const VideoMessage: React.FC<{ text: string, center?: boolean }> = ({text, center}) =>
    <Typography variant="body2" sx={{color: 'grey.400', p: 2, textAlign: center ? 'center' : undefined}}>{text}</Typography>;
