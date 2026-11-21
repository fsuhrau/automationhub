import React from 'react';
import {Button, Chip, CircularProgress, Stack, TableCell, TableRow, Tooltip, Typography} from '@mui/material';
import {PlayArrowRounded} from '@mui/icons-material';
import {Link as RouterLink, useNavigate} from 'react-router-dom';
import ITestData from '../../../../types/test';
import {formatDateTime} from '../../../../components/DateTime';
import {TestRoutes} from '../../../../lib/routes';
import {ago} from '../../../apps/bundlesModel';
import {describeConfig, STATUS} from '../../testsListModel';
import {RunHistory} from './RunHistory';

const hide = {display: {xs: 'none', md: 'table-cell'}};
const dim = (text: string): React.ReactNode => <Typography variant="body2" color="text.secondary">{text}</Typography>;

interface TestRowProps {
    test: ITestData,
    // whether any test of the list has its latest runs; else nothing is known
    historyKnown: boolean,
    now: number,
    testRoutes: TestRoutes,
    canRun: boolean,
    onRun: (test: ITestData) => void,
}

const LastRun: React.FC<Pick<TestRowProps, 'test' | 'historyKnown' | 'now' | 'testRoutes'>> = ({test, historyKnown, now, testRoutes}) => {
    const last = test.recentRuns?.[0];
    if (!last) return dim(historyKnown ? 'Never ran' : '–');
    const s = STATUS[last.status];
    return <Stack direction="row" spacing={1} alignItems="center">
        <Chip size="small" color={s.tone} label={s.label}
            icon={last.status === 'running' ? <CircularProgress size={12} color="inherit"/> : undefined}
            component={RouterLink} to={testRoutes.run(last.id).show()} clickable={true}
            onClick={(e: React.MouseEvent) => e.stopPropagation()}/>
        <Tooltip title={formatDateTime(last.createdAt)}><Typography variant="body2" color="text.secondary">{ago(last.createdAt, now)}</Typography></Tooltip>
    </Stack>;
};

// TestRow is a test in the list: what it runs, how its last run went, and Run
export const TestRow: React.FC<TestRowProps> = ({test, historyKnown, now, testRoutes, canRun, onRun}) => {
    const navigate = useNavigate();
    const runs = test.recentRuns;
    return <TableRow hover={true} sx={{cursor: 'pointer'}} onClick={() => navigate(testRoutes.show())}>
        <TableCell sx={{pl: 2}}>
            <Typography variant="body2" sx={{fontWeight: 600}}>{test.name}</Typography>
            <Typography variant="caption" color="text.secondary">{describeConfig(test.testConfig)}</Typography>
        </TableCell>
        <TableCell sx={{whiteSpace: 'nowrap'}}>
            <LastRun test={test} historyKnown={historyKnown} now={now} testRoutes={testRoutes}/>
        </TableCell>
        <TableCell sx={hide}>
            {runs && runs.length > 0 ? <RunHistory runs={runs} runUrl={id => testRoutes.run(id).show()}/> : dim('–')}
        </TableCell>
        <TableCell align="right" sx={{pr: 2}}>
            <Button size="small" variant="outlined" startIcon={<PlayArrowRounded/>} disabled={!canRun}
                onClick={e => {
                        e.stopPropagation();
                        onRun(test);
                    }}>Run</Button>
        </TableCell>
    </TableRow>;
};
