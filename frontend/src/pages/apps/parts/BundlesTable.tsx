import React from 'react';
import {Table, TableBody, TableCell, TableContainer, TableHead, TableRow} from '@mui/material';
import {IAppBinaryData} from '../../../types/app';
import {BundleRow, RowContext} from './BundleRow';
import {hideBelow} from './tableStyle';

interface BundlesTableProps {
    bundles: IAppBinaryData[],
    ctx: RowContext,
}

export const BundlesTable: React.FC<BundlesTableProps> = ({bundles, ctx}) =>
    <TableContainer sx={{mx: -2, width: 'auto'}}>
        <Table size="small" aria-label="Bundles">
            <TableHead>
                <TableRow>
                    <TableCell sx={{pl: 2}}>Version</TableCell>
                    <TableCell>Tags</TableCell>
                    <TableCell sx={hideBelow('md')}>Uploaded</TableCell>
                    <TableCell align="right" sx={hideBelow('sm')}>Size</TableCell>
                    <TableCell align="right" sx={hideBelow('md')}>Runs</TableCell>
                    <TableCell/>
                </TableRow>
            </TableHead>
            <TableBody>{bundles.map(b => <BundleRow key={b.id} bundle={b} ctx={ctx}/>)}</TableBody>
        </Table>
    </TableContainer>;
