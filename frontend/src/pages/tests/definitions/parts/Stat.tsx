import React from 'react';
import {Box, Typography} from '@mui/material';

interface StatProps {
    label: string,
    value: React.ReactNode,
    note?: React.ReactNode,
}

// Stat is one figure with its label and a note under it
export const Stat: React.FC<StatProps> = ({label, value, note}) =>
    <Box sx={{p: 1.5, border: 1, borderColor: 'divider', borderRadius: 1, height: '100%'}}>
        <Typography variant="caption" color="text.secondary">{label}</Typography>
        <Typography variant="h6" component="div">{value}</Typography>
        {note && <Typography variant="caption" color="text.secondary" component="div">{note}</Typography>}
    </Box>;
