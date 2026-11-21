import React from 'react';
import {Card, CardContent, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography} from '@mui/material';
import {CompareSection, Row} from '../../../types/bundle.compare';
import {ChangeMark, edgeColor} from './ChangeMark';
import {ListCells} from './ListCells';
import {ValueCells} from './ValueCells';

const hideOnPhone = {display: {xs: 'none', sm: 'table-cell'}};

const CompareRow: React.FC<{ row: Row }> = ({row}) =>
    <TableRow sx={{'& > td:first-of-type': {borderLeft: 3, borderLeftColor: edgeColor(row.change)}}}>
        <TableCell sx={{pl: 2, verticalAlign: 'top', color: 'text.secondary'}}>{row.label}</TableCell>
        {row.kind === 'value' ? <ValueCells row={row}/> : <ListCells row={row}/>}
        <TableCell sx={{verticalAlign: 'top', ...hideOnPhone}}><ChangeMark change={row.change}/></TableCell>
    </TableRow>;

// CompareSectionCard lists the attributes of one section with what changed
export const CompareSectionCard: React.FC<{ section: CompareSection }> = ({section}) =>
    <Card variant="outlined">
        <CardContent>
            <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600, mb: 1}}>{section.title}</Typography>
            <TableContainer sx={{mx: -2, width: 'auto'}}>
                <Table size="small" aria-label={section.title} sx={{tableLayout: 'fixed'}}>
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{pl: 2, width: {xs: '28%', md: '20%'}}}>Attribute</TableCell>
                            <TableCell>From</TableCell>
                            <TableCell>To</TableCell>
                            <TableCell sx={{width: 96, ...hideOnPhone}}/>
                        </TableRow>
                    </TableHead>
                    <TableBody>{section.rows.map(r => <CompareRow key={r.label} row={r}/>)}</TableBody>
                </Table>
            </TableContainer>
        </CardContent>
    </Card>;
