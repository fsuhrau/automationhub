import React from 'react';
import {Box} from '@mui/material';
import {useTheme} from '@mui/material/styles';
import {ResultCounts} from '../../testRunModel';
import {PaletteKey, RESULT_ORDER, resultLabel, resultPalette} from '../runStatusModel';

// ResultBar shows the share of each result as one stacked bar; the chips
// below it are the filter of the test list.
export const ResultBar: React.FC<{ counts: ResultCounts }> = ({counts}) => {
    const theme = useTheme();
    const color = (key: PaletteKey): string => (theme.vars ?? theme).palette[key].main;
    return <Box role="img" aria-label={RESULT_ORDER.map(r => `${counts[r]} ${resultLabel(r, false)}`).join(', ')}
        sx={{display: 'flex', height: 8, borderRadius: 4, overflow: 'hidden', gap: '2px'}}>
        {RESULT_ORDER.filter(r => counts[r] > 0).map(r =>
            <Box key={r} sx={{flexGrow: counts[r], bgcolor: color(resultPalette(r))}}/>)}
    </Box>;
};
