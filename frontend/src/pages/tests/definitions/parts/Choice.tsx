import React from 'react';
import {Box, FormControlLabel, Radio, Typography} from '@mui/material';

interface ChoiceProps {
    value: string,
    label: string,
    hint: string,
}

// Choice is a radio choice with a line that says what it means
export const Choice: React.FC<ChoiceProps> = ({value, label, hint}) =>
    <FormControlLabel value={value} control={<Radio size="small"/>} sx={{alignItems: 'flex-start', mb: 0.5, '& .MuiRadio-root': {pt: 0.25}}}
        label={<Box>
            <Typography variant="body2" sx={{fontWeight: 600}}>{label}</Typography>
            <Typography variant="caption" color="text.secondary">{hint}</Typography>
        </Box>}/>;
