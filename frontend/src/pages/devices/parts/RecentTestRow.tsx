import React from 'react';
import {TableCell, TableRow, Typography} from '@mui/material';
import {useNavigate} from 'react-router-dom';
import ITestProtocolData from '../../../types/test.protocol';
import {ResultIcon} from '../../../components/common';
import {formatDateTime} from '../../../components/DateTime';
import {between, formatDuration, splitTestName} from '../../tests/testRunModel';
import {recentTestLink} from '../deviceDetailsModel';

interface RecentTestRowProps {
    projectId: string,
    protocol: ITestProtocolData,
}

// one test result that ran on the device; a click opens its test run
export const RecentTestRow: React.FC<RecentTestRowProps> = ({projectId, protocol: p}) => {
    const navigate = useNavigate();
    const {name, suite} = splitTestName(p.testName);
    const test = p.testRun?.test;
    const ms = between(p.startedAt, p.endedAt);
    const url = recentTestLink(projectId, p);
    return <TableRow hover={!!url} sx={{cursor: url ? 'pointer' : undefined}} onClick={() => url && navigate(url)}>
        <TableCell sx={{width: 36, pl: 0}}><ResultIcon result={p.testResult}/></TableCell>
        <TableCell>
            <Typography variant="body2" sx={{fontWeight: 600, wordBreak: 'break-word'}}>{name}</Typography>
            <Typography variant="caption" color="text.secondary">{test ? `${test.name} · run ${p.testRunId}` : suite}</Typography>
        </TableCell>
        <TableCell align="right" sx={{whiteSpace: 'nowrap', pr: 0}}>
            <Typography variant="body2">{formatDateTime(p.startedAt, 'YYYY/MM/DD HH:mm')}</Typography>
            <Typography variant="caption" color="text.secondary">{ms === null ? 'not finished' : formatDuration(ms)}</Typography>
        </TableCell>
    </TableRow>;
};
