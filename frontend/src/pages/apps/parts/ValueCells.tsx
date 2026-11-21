import React from 'react';
import {TableCell, Typography} from '@mui/material';
import {ValueRow} from '../../../types/bundle.compare';
import {percent, signedBytes, signedCount} from '../compareRows';

const Delta: React.FC<{ row: ValueRow }> = ({row}) => {
    if (row.delta === undefined || row.delta === 0) return null;
    const text = row.unit === 'bytes' ? signedBytes(row.delta) : signedCount(row.delta);
    const share = percent(row.from, row.delta);
    return <Typography component="span" variant="caption" sx={{ml: 1, fontWeight: 600, color: row.delta > 0 ? 'warning.main' : 'success.main'}}>
        {text}{share && ` (${share})`}
    </Typography>;
};

const cell = (missing: boolean): object => ({verticalAlign: 'top', color: missing ? 'text.disabled' : undefined, wordBreak: 'break-word'});

// ValueCells are the From and To cells of an attribute
export const ValueCells: React.FC<{ row: ValueRow }> = ({row}) => <>
    <TableCell sx={cell(row.a === undefined)}>{row.a ?? '–'}</TableCell>
    <TableCell sx={{...cell(row.b === undefined), fontWeight: row.change === 'same' ? undefined : 600}}>
        {row.b ?? '–'}<Delta row={row}/>
    </TableCell>
</>;
