import React from 'react';
import {Card, CardContent, LinearProgress, Typography} from '@mui/material';
import {BundleCompare} from '../useBundleCompare';
import {ComparePickers} from './ComparePicker';

// ComparePickerCard is the card with the two pickers, or why there are none
export const ComparePickerCard: React.FC<{ compare: BundleCompare, now: number }> = ({compare, now}) => {
    const {bundles, selection, pick, swap} = compare;
    return <Card variant="outlined" sx={{mb: 2}}>
        <CardContent>
            {bundles === null && <LinearProgress/>}
            {bundles !== null && bundles.length < 2
                && <Typography variant="body2" color="text.secondary">The app needs two bundles to compare.</Typography>}
            {bundles !== null && bundles.length >= 2
                && <ComparePickers bundles={bundles} from={selection.a} to={selection.b} now={now} onPick={pick} onSwap={swap}/>}
        </CardContent>
    </Card>;
};
