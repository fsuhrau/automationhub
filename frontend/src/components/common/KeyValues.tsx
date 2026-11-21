import React from 'react';
import {Box, Typography} from '@mui/material';

export interface KeyValuesProps {
    rows: [string, React.ReactNode][],
    monospace?: boolean,
}

// KeyValues lists name and value pairs in two aligned columns
export const KeyValues: React.FC<KeyValuesProps> = ({rows, monospace}) =>
    <Box sx={{display: 'grid', gridTemplateColumns: 'minmax(110px, max-content) 1fr', columnGap: 2, rowGap: 0.75}}>
        {rows.map(([k, v]) => <React.Fragment key={k}>
            <Typography variant="body2" color="text.secondary">{k}</Typography>
            <Typography variant="body2" component="div" sx={{wordBreak: 'break-word', minWidth: 0, fontFamily: monospace ? 'monospace' : undefined}}>
                {v}
            </Typography>
        </React.Fragment>)}
    </Box>;
