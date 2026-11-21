import React, {useState} from 'react';
import {Stack} from '@mui/material';
import Grid from '@mui/material/Grid';
import {Section} from '../../../../components/common';
import {IBaselineCandidate} from '../../../../types/protocol.baseline';
import {ProtocolRef} from '../../../../types/test.run.refs';
import BaselinePicker from '../../BaselinePicker';
import {checkpointRows} from '../performanceModel';
import {ComparisonState} from '../useComparison';
import {ProtocolView} from '../useProtocolView';
import {CheckpointsTable} from './CheckpointsTable';
import {CompareBar} from './CompareBar';
import {MetricChart} from './MetricChart';
import {MetricTiles} from './MetricTiles';

interface PerformanceTabProps {
    refs: ProtocolRef,
    view: ProtocolView,
    state: ComparisonState,
    // picks the run to compare with, null for the latest earlier one
    onBaseline: (candidate: IBaselineCandidate | null) => void,
    // whether a run is picked in the url
    picked: boolean,
}

const Charts: React.FC<{ view: ProtocolView, state: ComparisonState }> = ({view, state}) =>
    <Grid container={true} spacing={2}>
        {view.metrics.map(m => <Grid key={m.key} size={{xs: 12, md: 6}}>
            <Section title={`${m.label}${m.unit ? ` (${m.unit})` : ''}`}>
                <MetricChart metric={m} protocol={view.protocol} earlier={state.comparison}/>
            </Section>
        </Grid>)}
    </Grid>;

// the performance of the test: averages, one chart per metric and the checkpoints
export const PerformanceTab: React.FC<PerformanceTabProps> = ({refs, view, state, onBaseline, picked}) => {
    const [picking, setPicking] = useState(false);
    const {protocol, metrics, earlier} = view;
    const {comparison} = state;
    const checkpoints = checkpointRows(protocol.performance, comparison?.performance);
    return <Stack spacing={2}>
        <CompareBar baseline={state.baseline} loading={state.loading} failed={state.failed} earlierRunId={earlier?.testRunId ?? null}
            onChange={() => setPicking(true)} onReset={picked ? () => onBaseline(null) : null}/>
        <BaselinePicker open={picking} onClose={() => setPicking(false)} refs={refs} deviceId={protocol.deviceId}
            onPick={c => {
                            setPicking(false);
                            onBaseline(c);
                        }}
            selectedRunId={comparison?.runId ?? null} metrics={metrics}/>
        {metrics.length > 0 && <Section title="Averages"><MetricTiles protocol={protocol} metrics={metrics} against={comparison}/></Section>}
        <Charts view={view} state={state}/>
        {checkpoints.length > 0 && <CheckpointsTable rows={checkpoints} comparison={comparison}/>}
    </Stack>;
};
