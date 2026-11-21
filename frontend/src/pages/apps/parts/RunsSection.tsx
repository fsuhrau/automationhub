import React from 'react';
import {Chip, Link, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Tooltip, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import {Section} from '../../../components/common';
import {formatDateTime} from '../../../components/DateTime';
import {BundleRun, IAppBundleDetail} from '../../../types/app';
import {RunSummaryStatus} from '../../../types/test';
import {STATUS} from '../../tests/testsListModel';
import {runSummary} from '../bundleDetailModel';
import {ago} from '../bundlesModel';

interface RunRowProps {
    run: BundleRun,
    to: string,
    now: number,
}

const RunRow: React.FC<RunRowProps> = ({run, to, now}) => {
    const status = STATUS[run.status as RunSummaryStatus] ?? {label: run.status, tone: 'default'};
    return <TableRow hover={true}>
        <TableCell sx={{pl: 0}}><Link component={RouterLink} to={to}>#{run.id}</Link></TableCell>
        <TableCell>{run.testName || `Test ${run.testId}`}</TableCell>
        <TableCell>
            <Chip size="small" color={status.tone} variant={status.tone === 'default' ? 'outlined' : 'filled'} label={status.label}/>
            <Typography component="span" variant="caption" color="text.secondary" sx={{ml: 1}}>{runSummary(run)}</Typography>
        </TableCell>
        <TableCell sx={{display: {xs: 'none', sm: 'table-cell'}, whiteSpace: 'nowrap'}}>
            <Tooltip title={formatDateTime(run.createdAt)}><span>{ago(run.createdAt, now)}</span></Tooltip>
        </TableCell>
    </TableRow>;
};

interface RunsSectionProps {
    detail: IAppBundleDetail,
    runUrl: (testId: number, runId: number) => string,
    now: number,
}

const RunsTable: React.FC<RunsSectionProps> = ({detail, runUrl, now}) =>
    <TableContainer>
        <Table size="small" aria-label="Test runs with this bundle">
            <TableHead>
                <TableRow>
                    <TableCell sx={{pl: 0}}>Run</TableCell>
                    <TableCell>Test</TableCell>
                    <TableCell>Result</TableCell>
                    <TableCell sx={{display: {xs: 'none', sm: 'table-cell'}}}>Started</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {detail.runs.map(r => <RunRow key={r.id} run={r} to={runUrl(r.testId, r.id)} now={now}/>)}
            </TableBody>
        </Table>
    </TableContainer>;

export const RunsSection: React.FC<RunsSectionProps> = props => {
    const {detail} = props;
    return <Section title={`Test runs (${detail.runCount})`}
        description={detail.runCount > detail.runs.length ? `The latest ${detail.runs.length}` : undefined}>
        {detail.runs.length === 0
            ? <Typography variant="body2" color="text.secondary">No test ran with this bundle yet.</Typography>
            : <RunsTable {...props}/>}
    </Section>;
};
