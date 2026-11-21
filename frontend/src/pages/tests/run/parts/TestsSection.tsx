import React from 'react';
import {Alert, Box, Link, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography} from '@mui/material';
import {Section} from '../../../../components/common';
import {TestResultState} from '../../../../types/test.result.state.enum';
import ITesRunLogEntryData from '../../../../types/test.run.log.entry';
import {filterRows, TestRow as Row} from '../../testRunModel';
import {RunView} from '../useRunView';
import {ResultBar} from './ResultBar';
import {ResultFilters} from './ResultFilters';
import {TestRows} from './TestRows';
import {TestFilters, useTestFilters} from './useTestFilters';

const visuallyHidden = {
    border: 0, clip: 'rect(0 0 0 0)', height: '1px', margin: '-1px', overflow: 'hidden',
    padding: 0, position: 'absolute', whiteSpace: 'nowrap', width: '1px',
} as const;

const Hidden: React.FC<{ children: string }> = ({children}) => <Box component="span" sx={visuallyHidden}>{children}</Box>;

const TestsHead: React.FC = () =>
    <TableHead>
        <TableRow>
            <TableCell sx={{pl: 2, pr: 1}}><Hidden>Result</Hidden></TableCell>
            <TableCell sx={{px: 1}}>Test</TableCell>
            <TableCell align="right" sx={{px: 1}}>Duration</TableCell>
            <TableCell sx={{pl: 0, pr: 1}}><Hidden>Actions</Hidden></TableCell>
        </TableRow>
    </TableHead>;

// what the list says while no test result exists
const NoResults: React.FC<{ view: RunView }> = ({view}) => {
    if (view.status !== 'not_run') {
        return <Typography variant="body2" color="text.secondary">
            {view.running ? 'Waiting for the first test to start…' : 'This run has no test results.'}
        </Typography>;
    }
    return <Alert severity="error" variant="outlined">
        <Typography variant="body2" sx={{fontWeight: 600, mb: 0.5}}>No test ran. The run stopped with:</Typography>
        {view.logErrors.slice(-3).map((e: ITesRunLogEntryData) =>
            <Typography key={e.id} variant="body2" sx={{fontFamily: 'monospace', fontSize: 13}}>{e.log}</Typography>)}
    </Alert>;
};

const NoMatch: React.FC<{ onClear: () => void }> = ({onClear}) =>
    <Typography variant="body2" color="text.secondary" sx={{py: 2}}>
        No tests match.{' '}
        <Link component="button" variant="body2" onClick={onClear}>Show all</Link>
    </Typography>;

interface TestsTableProps {
    rows: Row[],
    running: boolean,
    filters: TestFilters,
    onOpen: (protocolId: number) => void,
    onReplay?: (row: Row) => void,
}

const TestsTable: React.FC<TestsTableProps> = ({rows, running, filters, onOpen, onReplay}) =>
    <TableContainer sx={{mx: -2, width: 'auto'}}>
        <Table size="small" aria-label="Tests of this run">
            <TestsHead/>
            <TableBody>
                {rows.map(row => <TestRows key={row.key} row={row} running={running} expanded={filters.isExpanded(row.key)}
                    onToggle={() => filters.toggle(row.key)} onOpen={onOpen} onReplay={onReplay}/>)}
            </TableBody>
        </Table>
    </TableContainer>;

interface TestsSectionProps {
    view: RunView,
    onOpen: (protocolId: number) => void,
    onReplay?: (row: Row) => void,
}

const Progress: React.FC<{ view: RunView }> = ({view}) => {
    const {running, total, counts} = view;
    const done = total - counts[TestResultState.TestResultOpen];
    return <Typography variant="body2" color="text.secondary">
        {running ? `${done} of ${total} done` : `${total} ${total === 1 ? 'test' : 'tests'}`}
    </Typography>;
};

// the tests of the run: results at a glance, narrowed by result or name
export const TestsSection: React.FC<TestsSectionProps> = ({view, onOpen, onReplay}) => {
    const filters = useTestFilters();
    const visible = filterRows(view.rows, filters.result, filters.search);
    return <Section title="Tests" fullHeight={true} action={view.total > 0 && <Progress view={view}/>}>
        {view.total > 0 && <ResultBar counts={view.counts}/>}
        {view.total > 0 && <ResultFilters counts={view.counts} total={view.total} running={view.running} result={filters.result}
            onResult={filters.setResult} search={filters.search} onSearch={filters.setSearch}/>}
        {view.total === 0 && <NoResults view={view}/>}
        {view.total > 0 && visible.length === 0 && <NoMatch onClear={filters.clear}/>}
        {visible.length > 0 && <TestsTable rows={visible} running={view.running} filters={filters} onOpen={onOpen} onReplay={onReplay}/>}
    </Section>;
};
