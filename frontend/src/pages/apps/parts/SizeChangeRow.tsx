import React from 'react';
import {Box, TableCell, TableRow, Typography} from '@mui/material';
import {SizeRow} from '../../../types/bundle.compare';
import {byteFormat} from '../../tests/value_formatter';
import {percent, signedBytes} from '../compareRows';
import {mono} from './mono';

const hideOnPhone = {display: {xs: 'none', sm: 'table-cell'}, whiteSpace: 'nowrap'};

const Missing: React.FC<{ text: string }> = ({text}) =>
    <Typography component="span" variant="caption" color="text.disabled">{text}</Typography>;

const sizeOrMissing = (size: number | undefined, missing: string): React.ReactNode =>
    size === undefined ? <Missing text={missing}/> : byteFormat(size);

const deltaColor = (delta: number): string => delta > 0 ? 'warning.main' : delta < 0 ? 'success.main' : 'text.secondary';

const DeltaText: React.FC<{ row: SizeRow }> = ({row}) => {
    if (row.delta === undefined) return <Missing text="unknown"/>;
    const share = row.a !== undefined && row.b !== undefined ? percent(row.a, row.delta) : undefined;
    return <>
        <Typography component="span" variant="body2" sx={{fontWeight: 600, color: deltaColor(row.delta)}}>{signedBytes(row.delta)}</Typography>
        {share && <Typography component="span" variant="caption" color="text.secondary" sx={{ml: 0.75}}>{share}</Typography>}
    </>;
};

interface DeltaBarProps {
    delta: number | undefined,
    max: number,
}

// DeltaBar grows to the right and shrinks to the left of the middle
const DeltaBar: React.FC<DeltaBarProps> = ({delta, max}) => {
    const width = `${Math.abs(delta ?? 0) / max * 100}%`;
    return <Box sx={{display: 'flex', height: 8}} aria-hidden={true}>
        <Box sx={{flex: 1, display: 'flex', justifyContent: 'flex-end'}}>
            {delta !== undefined && delta < 0 && <Box sx={{width, bgcolor: 'success.main', borderRadius: '4px 0 0 4px'}}/>}
        </Box>
        <Box sx={{width: '1px', bgcolor: 'divider'}}/>
        <Box sx={{flex: 1}}>
            {delta !== undefined && delta > 0 && <Box sx={{width, height: '100%', bgcolor: 'warning.main', borderRadius: '0 4px 4px 0'}}/>}
        </Box>
    </Box>;
};

interface SizeChangeRowProps {
    row: SizeRow,
    max: number,
    // what to say for a side that has no entry
    missing: string,
}

export const SizeChangeRow: React.FC<SizeChangeRowProps> = ({row, max, missing}) =>
    <TableRow>
        <TableCell sx={{pl: 2, ...mono}}>{row.path}</TableCell>
        <TableCell align="right" sx={hideOnPhone}>{sizeOrMissing(row.a, missing)}</TableCell>
        <TableCell align="right" sx={hideOnPhone}>{sizeOrMissing(row.b, missing)}</TableCell>
        <TableCell align="right" sx={{whiteSpace: 'nowrap'}}><DeltaText row={row}/></TableCell>
        <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}><DeltaBar delta={row.delta} max={max}/></TableCell>
    </TableRow>;
