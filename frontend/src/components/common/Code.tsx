import React from 'react';
import {Box} from '@mui/material';
import {CopyButton} from './CopyButton';

// Code shows text to copy into a terminal or a config file
export const Code: React.FC<{ children: string, label?: string }> = ({children, label = 'Copy'}) =>
    <Box sx={{display: 'flex', alignItems: 'flex-start', gap: 1}}>
        <Box component="pre" sx={{
            m: 0, p: 1.5, flex: 1, overflowX: 'auto', fontSize: 13, whiteSpace: 'pre',
            bgcolor: 'background.default', border: 1, borderColor: 'divider', borderRadius: 1,
        }}>{children}</Box>
        <CopyButton text={children} label={label}/>
    </Box>;
