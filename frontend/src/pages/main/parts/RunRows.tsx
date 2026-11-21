import React from 'react';
import {Chip, CircularProgress, Table, TableBody, TableCell, TableRow, Tooltip, Typography} from '@mui/material';
import {useNavigate} from 'react-router-dom';
import {IOverviewRun} from '../../../types/overview.run';
import {formatDateTime} from '../../../components/DateTime';
import {ago} from '../../apps/bundlesModel';
import {STATUS} from '../../tests/testsListModel';
import {runCounts} from '../dashboardModel';

interface RunRowsProps {
    runs: IOverviewRun[],
    runUrl: (r: IOverviewRun) => string,
    appName: (id: number) => React.ReactNode,
    now: number,
}

const RunRow: React.FC<RunRowsProps & { run: IOverviewRun }> = ({run, runUrl, appName, now}) => {
    const navigate = useNavigate();
    const status = STATUS[run.status];
    const counts = runCounts(run);
    return <TableRow hover={true} sx={{cursor: 'pointer', '&:last-child td': {borderBottom: 0}}} onClick={() => navigate(runUrl(run))}>
        <TableCell sx={{pl: 0, width: 120}}>
            <Chip size="small" color={status.tone} label={status.label}
                icon={run.status === 'running' ? <CircularProgress size={12} color="inherit"/> : undefined}/>
        </TableCell>
        <TableCell>
            <Typography variant="body2" sx={{fontWeight: 600}}>{run.testName}</Typography>
            <Typography variant="caption" color="text.secondary">{appName(run.appId)} · run {run.id}{counts && ` · ${counts}`}</Typography>
        </TableCell>
        <TableCell align="right" sx={{pr: 0, whiteSpace: 'nowrap'}}>
            <Tooltip title={formatDateTime(run.createdAt)}><Typography variant="body2" color="text.secondary">{ago(run.createdAt, now)}</Typography></Tooltip>
        </TableCell>
    </TableRow>;
};

const RunRows: React.FC<RunRowsProps> = props => <Table size="small">
    <TableBody>
        {props.runs.map(run => <RunRow key={run.id} {...props} run={run}/>)}
    </TableBody>
</Table>;

export default RunRows;
