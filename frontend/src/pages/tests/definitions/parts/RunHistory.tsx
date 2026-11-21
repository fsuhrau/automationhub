import React from 'react';
import {ButtonBase, Stack, Tooltip} from '@mui/material';
import {useTheme} from '@mui/material/styles';
import {Link as RouterLink} from 'react-router-dom';
import {IRunSummary} from '../../../../types/test';
import {formatDateTime} from '../../../../components/DateTime';
import {STATUS, Tone} from '../../testsListModel';

interface RunHistoryProps {
    runs: IRunSummary[],
    runUrl: (id: number) => string,
}

const countsOf = (r: IRunSummary): string =>
    [r.passed && `${r.passed} passed`, r.failed && `${r.failed} failed`, r.unstable && `${r.unstable} unstable`]
        .filter(Boolean).join(', ');

const RunTip: React.FC<{ run: IRunSummary }> = ({run}) => {
    const counts = countsOf(run);
    return <>Run {run.id}: {STATUS[run.status].label}{counts && ` (${counts})`}<br/>{formatDateTime(run.createdAt, 'YYYY/MM/DD HH:mm')}</>;
};

// RunHistory shows the latest runs oldest to newest, one square each; a square
// opens its run. Status is in the tooltip and the label, not only the color.
export const RunHistory: React.FC<RunHistoryProps> = ({runs, runUrl}) => {
    const theme = useTheme();
    const color = (tone: Tone): string => tone === 'default' ? theme.palette.action.disabled : (theme.vars ?? theme).palette[tone].main;
    return <Stack direction="row" spacing={0.5} role="list" aria-label="Latest runs, oldest first">
        {[...runs].reverse().map(r => <Tooltip key={r.id} title={<RunTip run={r}/>}>
            <ButtonBase component={RouterLink} to={runUrl(r.id)} role="listitem" aria-label={`Run ${r.id}: ${STATUS[r.status].label}`}
                onClick={e => e.stopPropagation()}
                sx={{width: 12, height: 20, borderRadius: 0.5, bgcolor: color(STATUS[r.status].tone)}}/>
        </Tooltip>)}
    </Stack>;
};
