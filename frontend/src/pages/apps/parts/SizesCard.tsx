import React, {useState} from 'react';
import {Button, Card, CardContent, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography} from '@mui/material';
import {SizeRow} from '../../../types/bundle.compare';
import {SizeChangeRow} from './SizeChangeRow';

const SHOWN = 15;
const hideOnPhone = {display: {xs: 'none', sm: 'table-cell'}};

interface SizesCardProps {
    title: string,
    subtitle: string,
    rows: SizeRow[],
    onlyChanged: boolean,
    missing: string,
}

const SizesTable: React.FC<{ title: string, rows: SizeRow[], missing: string }> = ({title, rows, missing}) => {
    const max = Math.max(1, ...rows.map(r => Math.abs(r.delta ?? 0)));
    return <TableContainer sx={{mx: -2, width: 'auto'}}>
        <Table size="small" aria-label={title}>
            <TableHead>
                <TableRow>
                    <TableCell sx={{pl: 2}}>Path</TableCell>
                    <TableCell align="right" sx={hideOnPhone}>From</TableCell>
                    <TableCell align="right" sx={hideOnPhone}>To</TableCell>
                    <TableCell align="right">Change</TableCell>
                    <TableCell sx={{width: '22%', display: {xs: 'none', md: 'table-cell'}}}/>
                </TableRow>
            </TableHead>
            <TableBody>{rows.map(r => <SizeChangeRow key={r.path} row={r} max={max} missing={missing}/>)}</TableBody>
        </Table>
    </TableContainer>;
};

// SizesCard shows how the size of each part or file changed, biggest change first
export const SizesCard: React.FC<SizesCardProps> = ({title, subtitle, rows, onlyChanged, missing}) => {
    const [all, setAll] = useState(false);
    if (rows.length === 0) return null;
    const filtered = onlyChanged ? rows.filter(r => r.change !== 'same') : rows;
    const shown = all ? filtered : filtered.slice(0, SHOWN);
    return <Card variant="outlined">
        <CardContent>
            <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
            <Typography variant="body2" color="text.secondary" sx={{mb: 1}}>{subtitle}</Typography>
            {filtered.length === 0
                ? <Typography variant="body2" color="text.secondary">No size changed.</Typography>
                : <SizesTable title={title} rows={shown} missing={missing}/>}
            {filtered.length > shown.length && <Button size="small" sx={{mt: 1}} onClick={() => setAll(true)}>Show all {filtered.length}</Button>}
        </CardContent>
    </Card>;
};
