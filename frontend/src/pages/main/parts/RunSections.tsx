import React from 'react';
import {Stack, Typography} from '@mui/material';
import Grid from '@mui/material/Grid';
import {CheckCircleRounded} from '@mui/icons-material';
import {IProjectOverview} from '../../../types/overview.project';
import {Section} from '../../../components/common';
import RunRows from './RunRows';

export interface RunListProps {
    overview: IProjectOverview,
    runUrl: React.ComponentProps<typeof RunRows>['runUrl'],
    appName: React.ComponentProps<typeof RunRows>['appName'],
    now: number,
}

type ListProps = Omit<RunListProps, 'overview'>;

const AllPassing: React.FC = () => (
    <Stack direction="row" spacing={1} alignItems="center">
        <CheckCircleRounded color="success" fontSize="small"/>
        <Typography variant="body2">No test failed in its latest run.</Typography>
    </Stack>
);

const NeedsAttention: React.FC<ListProps & { overview: IProjectOverview }> = ({overview, ...rows}) => (
    <Section title="Needs attention" fullHeight={true}>
        {overview.failing.length === 0 ? <AllPassing/> : <>
            <Typography variant="caption" color="text.secondary">Tests whose latest run failed, was unstable or didn't run.</Typography>
            <RunRows runs={overview.failing} {...rows}/>
        </>}
    </Section>
);

const LatestRuns: React.FC<ListProps & { overview: IProjectOverview }> = ({overview, ...rows}) => (
    <Section title="Latest runs" fullHeight={true}>
        {overview.latest.length === 0
            ? <Typography variant="body2" color="text.secondary">Nothing ran yet. Start a test from Tests.</Typography>
            : <RunRows runs={overview.latest} {...rows}/>}
    </Section>
);

// what needs attention, what runs now, and what ran
const RunSections: React.FC<RunListProps> = ({overview, ...rows}) => (
    <Grid container={true} spacing={2}>
        <Grid size={{xs: 12, lg: 6}}>
            <Stack spacing={2}>
                <NeedsAttention overview={overview} {...rows}/>
                {overview.running.length > 0 && <Section title="Running now" fullHeight={true}>
                    <RunRows runs={overview.running} {...rows}/>
                </Section>}
            </Stack>
        </Grid>
        <Grid size={{xs: 12, lg: 6}}>
            <LatestRuns overview={overview} {...rows}/>
        </Grid>
    </Grid>
);

export default RunSections;
