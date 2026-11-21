import React from 'react';
import {Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Tooltip, Typography} from '@mui/material';
import {NorthRounded, SouthRounded} from '@mui/icons-material';
import {Section} from '../../../../components/common';
import {formatRuntime} from '../logModel';
import {CheckpointRow, compare} from '../performanceModel';
import {Comparison} from '../protocolViewModel';

const VerdictArrow: React.FC<{ row: CheckpointRow, runId: number | undefined }> = ({row, runId}) => {
    const verdict = compare(row.took, row.earlierTook, false);
    if (verdict === 'worse') {
        return <Tooltip title={`Slower than run ${runId}`}><NorthRounded fontSize="inherit" color="warning" sx={{ml: 0.5}}/></Tooltip>;
    }
    return verdict === 'better'
        ? <Tooltip title={`Faster than run ${runId}`}><SouthRounded fontSize="inherit" color="success" sx={{ml: 0.5}}/></Tooltip>
        : null;
};

const CheckpointLine: React.FC<{ row: CheckpointRow, comparison: Comparison | null }> = ({row, comparison}) =>
    <TableRow>
        <TableCell sx={{wordBreak: 'break-word'}}>{row.checkpoint}</TableCell>
        <TableCell align="right" sx={{fontFamily: 'monospace'}}>{formatRuntime(row.at)}</TableCell>
        <TableCell align="right" sx={{fontFamily: 'monospace', whiteSpace: 'nowrap'}}>
            {row.took.toFixed(2)} s
            <VerdictArrow row={row} runId={comparison?.runId}/>
        </TableCell>
        {comparison && <TableCell align="right" sx={{fontFamily: 'monospace', color: 'text.secondary'}}>
            {row.earlierTook === null ? '–' : `${row.earlierTook.toFixed(2)} s`}
        </TableCell>}
    </TableRow>;

// the checkpoints of the test with the time each took, next to the compared run
export const CheckpointsTable: React.FC<{ rows: CheckpointRow[], comparison: Comparison | null }> = ({rows, comparison}) =>
    <Section title="Checkpoints" action={comparison &&
        <Typography variant="body2" color="text.secondary">compared with run {comparison.runId}</Typography>}>
        <TableContainer>
            <Table size="small" aria-label="Checkpoints">
                <TableHead>
                    <TableRow>
                        <TableCell>Checkpoint</TableCell>
                        <TableCell align="right">Reached at</TableCell>
                        <TableCell align="right">Took</TableCell>
                        {comparison && <TableCell align="right">Run {comparison.runId}</TableCell>}
                    </TableRow>
                </TableHead>
                <TableBody>
                    {rows.map(c => <CheckpointLine key={`${c.checkpoint}@${c.at}`} row={c} comparison={comparison}/>)}
                </TableBody>
            </Table>
        </TableContainer>
    </Section>;
