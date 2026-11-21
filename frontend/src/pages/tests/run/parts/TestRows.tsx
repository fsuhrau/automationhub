import React from 'react';
import {IconButton, TableCell, TableRow, Tooltip, Typography} from '@mui/material';
import {SlowMotionVideoRounded} from '@mui/icons-material';
import {ResultIcon} from '../../../../components/common';
import {between, formatDuration, TestRow} from '../../testRunModel';
import {TestNameCell} from './TestNameCell';

export interface TestRowsProps {
    row: TestRow,
    running: boolean,
    expanded: boolean,
    onToggle: () => void,
    onOpen: (protocolId: number) => void,
    // opens the replay where the test started; only for recorded runs
    onReplay?: (row: TestRow) => void,
    nested?: boolean,
}

const DurationCell: React.FC<{ row: TestRow, running: boolean }> = ({row, running}) => {
    const ms = between(row.startedAt, row.endedAt);
    return <TableCell align="right" sx={{whiteSpace: 'nowrap', px: 1}}>
        <Typography variant="body2" color={ms === null ? 'text.secondary' : 'text.primary'}>
            {ms === null ? (running ? 'running' : '–') : formatDuration(ms)}
        </Typography>
    </TableCell>;
};

const ReplayCell: React.FC<{ row: TestRow, onReplay?: (row: TestRow) => void }> = ({row, onReplay}) =>
    <TableCell align="right" sx={{whiteSpace: 'nowrap', pl: 0, pr: 1}}>
        {onReplay && row.children.length === 0 && <Tooltip title="Replay this test">
            <IconButton size="small" aria-label={`Replay ${row.name}`}
                onClick={e => {
                            e.stopPropagation();
                            onReplay(row);
                        }}>
                <SlowMotionVideoRounded fontSize="small"/>
            </IconButton>
        </Tooltip>}
    </TableCell>;

const keyboardOpen = (open: () => void) => (e: React.KeyboardEvent): void => {
    if (e.target === e.currentTarget && (e.key === 'Enter' || e.key === ' ')) {
        e.preventDefault();
        open();
    }
};

// TestRows is a row of the tests table and, expanded, the rows of its children
export const TestRows: React.FC<TestRowsProps> = ({row, running, expanded, onToggle, onOpen, onReplay, nested}) => {
    const hasChildren = row.children.length > 0;
    const open = (): void => hasChildren ? onToggle() : onOpen(row.protocol.id as number);
    return <>
        <TableRow hover={true} onClick={open} aria-expanded={hasChildren ? expanded : undefined} tabIndex={0}
            onKeyDown={keyboardOpen(open)} sx={{cursor: 'pointer', '& > td': {borderBottomColor: 'divider'}}}>
            <TableCell sx={{width: 40, pl: nested ? 4 : 2, pr: 1}}>
                <ResultIcon result={row.result} running={running}/>
            </TableCell>
            <TestNameCell row={row} hasChildren={hasChildren} expanded={expanded}/>
            <DurationCell row={row} running={running}/>
            <ReplayCell row={row} onReplay={onReplay}/>
        </TableRow>
        {hasChildren && expanded && row.children.map(child =>
            <TestRows key={child.key} row={child} running={running} expanded={false} onToggle={() => undefined}
                onOpen={onOpen} onReplay={onReplay} nested={true}/>)}
    </>;
};
