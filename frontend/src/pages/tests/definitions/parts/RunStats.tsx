import React, {useMemo} from 'react';
import Grid from '@mui/material/Grid';
import {ITestRunListEntry} from '../../../../types/test.run.list';
import {formatDateTime} from '../../../../components/DateTime';
import {ago} from '../../../apps/bundlesModel';
import {formatDuration} from '../../testRunModel';
import {runStats, STATUS} from '../../testsListModel';
import {Stat} from './Stat';
import {useAgoClock} from '../useAgoClock';

const cell = {xs: 6, md: 3};

// RunStats sums up the loaded runs: the last one, how many passed, when it
// last passed and how long it usually takes
export const RunStats: React.FC<{ runs: ITestRunListEntry[] }> = ({runs}) => {
    const stats = useMemo(() => runStats(runs), [runs]);
    const now = useAgoClock();
    const last = runs[0];
    const passedAt = stats.lastPassedAt;
    return <Grid container={true} spacing={1.5}>
        <Grid size={cell}>
            <Stat label="Last run" value={last.status ? STATUS[last.status].label : '–'} note={ago(last.createdAt, now)}/>
        </Grid>
        <Grid size={cell}>
            <Stat label="Passed" value={stats.judged ? `${stats.passed} of ${stats.judged}` : '–'} note="of the runs below with a result"/>
        </Grid>
        <Grid size={cell}>
            <Stat label="Last passed" value={passedAt ? ago(passedAt, now) : 'not yet'}
                note={passedAt ? formatDateTime(passedAt, 'YYYY/MM/DD HH:mm') : undefined}/>
        </Grid>
        <Grid size={cell}>
            <Stat label="Usually takes" value={stats.typicalMs !== null ? formatDuration(stats.typicalMs) : '–'} note="when it passes"/>
        </Grid>
    </Grid>;
};
