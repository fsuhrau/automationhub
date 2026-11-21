import React from 'react';
import {Chip, FormControlLabel, Stack, Switch, Typography} from '@mui/material';
import {IAppBundleDetail} from '../../../types/app';

interface CompareSummaryProps {
    from: IAppBundleDetail,
    to: IAppBundleDetail,
    changes: number,
    changesOnly: boolean,
    onChangesOnly: (on: boolean) => void,
}

const summaryText = (from: IAppBundleDetail, to: IAppBundleDetail, changes: number): string => {
    if (from.id === to.id) return 'The same bundle on both sides.';
    if (changes === 0) return 'No attribute changed.';
    return `${changes} ${changes === 1 ? 'attribute changed' : 'attributes changed'}`;
};

// CompareSummary says how much changed and filters to the changes
export const CompareSummary: React.FC<CompareSummaryProps> = ({from, to, changes, changesOnly, onChangesOnly}) =>
    <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap" useFlexGap={true} sx={{mb: 2}}>
        <Typography variant="body1" sx={{fontWeight: 600}} aria-live="polite">{summaryText(from, to, changes)}</Typography>
        {from.id > to.id && <Chip size="small" color="info" variant="outlined" label="From is the newer bundle"/>}
        <FormControlLabel sx={{ml: 'auto'}} label="Only changes"
            control={<Switch checked={changesOnly} onChange={e => onChangesOnly(e.target.checked)}/>}/>
    </Stack>;
