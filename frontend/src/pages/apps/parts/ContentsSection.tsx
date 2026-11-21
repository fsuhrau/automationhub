import React from 'react';
import {Grid, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography} from '@mui/material';
import {Section} from '../../../components/common';
import {Contents, SizeEntry} from '../../../types/bundle.info';
import {byteFormat} from '../../tests/value_formatter';
import {SizeBars} from './SizeBars';
import {mono} from './mono';

const hideOnPhone = {display: {xs: 'none', sm: 'table-cell'}};

const LargestFiles: React.FC<{ files: SizeEntry[] }> = ({files}) =>
    <TableContainer>
        <Table size="small" aria-label="Largest files">
            <TableHead>
                <TableRow>
                    <TableCell sx={{pl: 0}}>File</TableCell>
                    <TableCell align="right">Installed</TableCell>
                    <TableCell align="right" sx={hideOnPhone}>In the file</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {files.map(f => <TableRow key={f.path}>
                    <TableCell sx={{pl: 0, ...mono}}>{f.path}</TableCell>
                    <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>{byteFormat(f.size)}</TableCell>
                    <TableCell align="right" sx={{whiteSpace: 'nowrap', ...hideOnPhone}}>{byteFormat(f.compressed)}</TableCell>
                </TableRow>)}
            </TableBody>
        </Table>
    </TableContainer>;

const description = (c: Contents): string =>
    `${c.files.toLocaleString()} files · ${byteFormat(c.size)} installed · ${byteFormat(c.compressed)} compressed`;

export const ContentsSection: React.FC<{ contents: Contents }> = ({contents}) =>
    <Section title="Contents" description={description(contents)}>
        <Grid container={true} spacing={3}>
            <Grid size={{xs: 12, lg: 6}}>
                <Typography variant="subtitle2" sx={{mb: 0.5}}>Size by part</Typography>
                <SizeBars contents={contents}/>
            </Grid>
            <Grid size={{xs: 12, lg: 6}}>
                <Typography variant="subtitle2" sx={{mb: 0.5}}>Largest files</Typography>
                <LargestFiles files={contents.largest ?? []}/>
            </Grid>
        </Grid>
    </Section>;
