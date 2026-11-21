import React from 'react';
import {Box, Typography} from '@mui/material';
import {Code} from '../../../components/common';

export interface GuideStepProps {
    title: string,
    code?: string,
    children: React.ReactNode,
}

// GuideStep is one numbered step of the guide, with a command to copy
const GuideStep: React.FC<GuideStepProps> = ({title, code, children}) =>
    <Box component="li" sx={{
        counterIncrement: 'step', position: 'relative', pl: 4.5, pb: 2,
        '&::before': {
            content: 'counter(step)', position: 'absolute', left: 0, top: 0, width: 24, height: 24, borderRadius: '50%',
            display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 13, fontWeight: 600,
            bgcolor: 'primary.main', color: 'primary.contrastText',
        },
    }}>
        <Typography variant="body2" sx={{fontWeight: 600, lineHeight: '24px'}}>{title}</Typography>
        <Typography variant="body2" color="text.secondary" sx={{mb: code ? 1 : 0}}>{children}</Typography>
        {code && <Code label="Copy the command">{code}</Code>}
    </Box>;

export default GuideStep;
