import React from 'react';
import {Box, Stack, Typography} from '@mui/material';
import {CloudUploadRounded} from '@mui/icons-material';

// DropOverlay covers the page while files are dragged over it
export const DropOverlay: React.FC<{ appName?: string, accept: string }> = ({appName, accept}) =>
    <Box sx={{
        position: 'absolute', inset: -8, zIndex: 2, border: '2px dashed', borderColor: 'primary.main', borderRadius: 2,
        bgcolor: 'background.default', opacity: 0.94, display: 'flex', alignItems: 'center', justifyContent: 'center',
    }}>
        <Stack alignItems="center" spacing={1}>
            <CloudUploadRounded color="primary" sx={{fontSize: 48}}/>
            <Typography variant="h6">Drop to upload to {appName ?? 'this app'}</Typography>
            <Typography variant="body2" color="text.secondary">{accept}</Typography>
        </Stack>
    </Box>;
