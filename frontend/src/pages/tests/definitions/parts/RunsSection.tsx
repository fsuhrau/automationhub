import React from 'react';
import {Alert, Button, LinearProgress, Stack, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography} from '@mui/material';
import {PlayArrowRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {Section} from '../../../../components/common';
import {TestRoutes} from '../../../../lib/routes';
import {ITestRunListEntry} from '../../../../services/test.run.service';
import {TestRuns} from '../useTestRuns';
import {RunRow} from './RunRow';
import {useAgoClock} from '../useAgoClock';

interface RunsSectionProps {
    runs: TestRuns,
    testRoutes: TestRoutes,
    canRun: boolean,
    onRun: () => void,
}

const md = {display: {xs: 'none', md: 'table-cell'}};

const RunsTable: React.FC<{ runs: ITestRunListEntry[], testRoutes: TestRoutes }> = ({runs, testRoutes}) => {
    const now = useAgoClock();
    return <TableContainer sx={{mx: -2, width: 'auto'}}>
        <Table size="small" aria-label="Runs of this test">
            <TableHead>
                <TableRow>
                    <TableCell sx={{pl: 2}}>Run</TableCell>
                    <TableCell>Started</TableCell>
                    <TableCell sx={{display: {xs: 'none', sm: 'table-cell'}}}>Took</TableCell>
                    <TableCell sx={md}>Results</TableCell>
                    <TableCell sx={md}>Bundle and parameters</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {runs.map(r => <RunRow key={r.id} run={r} runRoutes={testRoutes.run(r.id)} now={now}/>)}
            </TableBody>
        </Table>
    </TableContainer>;
};

const NeverRan: React.FC<Pick<RunsSectionProps, 'canRun' | 'onRun'>> = ({canRun, onRun}) =>
    <Stack alignItems="flex-start" spacing={1}>
        <Typography variant="body2" color="text.secondary">This test hasn't run yet.</Typography>
        <Button variant="outlined" startIcon={<PlayArrowRounded/>} disabled={!canRun} onClick={onRun}>Run it</Button>
    </Stack>;

const RunsList: React.FC<RunsSectionProps> = ({runs: state, testRoutes, canRun, onRun}) => {
    const {runs, more, loading, loadOlder} = state;
    if (runs === null) return <LinearProgress/>;
    if (runs.length === 0) return <NeverRan canRun={canRun} onRun={onRun}/>;
    return <>
        <RunsTable runs={runs} testRoutes={testRoutes}/>
        {more && <Button size="small" sx={{mt: 1}} disabled={loading} onClick={loadOlder}>{loading ? 'Loading…' : 'Older runs'}</Button>}
        {!runs.some(r => r.status !== undefined) && <Typography variant="caption" color="text.secondary" component="div" sx={{mt: 1}}>
            The master doesn't summarize runs yet; open a run for its results.
        </Typography>}
    </>;
};

// RunsSection lists the runs of the test, newest first
export const RunsSection: React.FC<RunsSectionProps> = props => {
    const {runs, error, reload} = props.runs;
    const latest = runs && runs.length > 0 &&
        <Button size="small" component={RouterLink} to={props.testRoutes.lastRun()}>Latest run</Button>;
    return <Section title="Runs" action={latest}>
        {error
            ? <Alert severity="error" action={<Button color="inherit" size="small" onClick={reload}>Retry</Button>}>{error}</Alert>
            : <RunsList {...props}/>}
    </Section>;
};
