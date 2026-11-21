import React from 'react';
import {Chip, CircularProgress, Link, Stack, TableCell, TableRow, Tooltip, Typography} from '@mui/material';
import {Link as RouterLink, useNavigate} from 'react-router-dom';
import {ITestRunListEntry} from '../../../../services/test.run.service';
import {formatDateTime} from '../../../../components/DateTime';
import {RunRoutes} from '../../../../lib/routes';
import {ago} from '../../../apps/bundlesModel';
import {between, formatDuration, parseParameters} from '../../testRunModel';
import {STATUS} from '../../testsListModel';
import {runResults} from '../testShowModel';

const sm = {display: {xs: 'none', sm: 'table-cell'}, whiteSpace: 'nowrap'};
const md = {display: {xs: 'none', md: 'table-cell'}};

const duration = (r: ITestRunListEntry): string => {
    if (r.status === 'running') return 'running';
    const ms = between(r.createdAt, r.endedAt);
    return ms !== null ? formatDuration(ms) : '–';
};

const BundleCell: React.FC<{ run: ITestRunListEntry }> = ({run}) => {
    const params = parseParameters(run.parameter);
    const bundle = run.bundle;
    return <TableCell sx={{...md, maxWidth: 320}}>
        <Typography variant="body2" noWrap={true} title={bundle ? `${bundle.version} (#${bundle.id})` : undefined}>
            {bundle ? bundle.version || bundle.name : '–'}
        </Typography>
        {params.length > 0 && <Typography variant="caption" color="text.secondary" noWrap={true} component="div"
            title={params.map(p => `${p.name}=${p.value}`).join('\n')}>
            {params.map(p => p.value ? `${p.name}=${p.value}` : p.name).join(' · ')}
        </Typography>}
    </TableCell>;
};

const RunCell: React.FC<{ run: ITestRunListEntry, url: string }> = ({run, url}) => {
    const s = run.status ? STATUS[run.status] : null;
    return <TableCell sx={{pl: 2, whiteSpace: 'nowrap'}}>
        <Stack direction="row" spacing={1} alignItems="center">
            {s && <Chip size="small" color={s.tone} label={s.label}
                icon={run.status === 'running' ? <CircularProgress size={12} color="inherit"/> : undefined}/>}
            <Link component={RouterLink} to={url} underline="hover" onClick={e => e.stopPropagation()}>#{run.id}</Link>
        </Stack>
    </TableCell>;
};

// RunRow is a run of the test; the row opens it
export const RunRow: React.FC<{ run: ITestRunListEntry, runRoutes: RunRoutes, now: number }> = ({run, runRoutes, now}) => {
    const navigate = useNavigate();
    return <TableRow hover={true} sx={{cursor: 'pointer'}} onClick={() => navigate(runRoutes.show())}>
        <RunCell run={run} url={runRoutes.show()}/>
        <TableCell sx={{whiteSpace: 'nowrap'}}>
            <Tooltip title={formatDateTime(run.createdAt)}><span>{ago(run.createdAt, now)}</span></Tooltip>
        </TableCell>
        <TableCell sx={sm}>{duration(run)}</TableCell>
        <TableCell sx={md}>
            <Typography variant="body2" color="text.secondary">{runResults(run) || '–'}</Typography>
        </TableCell>
        <BundleCell run={run}/>
    </TableRow>;
};
