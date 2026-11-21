import React from 'react';
import {Box, Typography} from '@mui/material';
import ITestProtocolData from '../../../../types/test.protocol';
import {averageOf, compare, Metric, Verdict} from '../performanceModel';
import {Comparison} from '../protocolViewModel';

const VerdictNote: React.FC<{ verdict: Verdict, text: string, same: string }> = ({verdict, text, same}) => {
    if (verdict === null) return null;
    const color = verdict === 'better' ? 'success.main' : verdict === 'worse' ? 'warning.main' : 'text.secondary';
    return <Typography variant="caption" sx={{color}}>
        {verdict === 'same' ? same : verdict} · {text}
    </Typography>;
};

// the average the protocol is held against: the picked run's, or the usual
// one of earlier runs; none without a comparison
const usualOf = (protocol: ITestProtocolData, m: Metric, against: Comparison | null): number | null | undefined =>
    against?.averages ? averageOf(against.averages, m.key) : against ? m.usual(protocol) : null;

const MetricTile: React.FC<{ protocol: ITestProtocolData, metric: Metric, against: Comparison | null }> = ({protocol, metric, against}) => {
    const value = metric.avg(protocol);
    const usual = usualOf(protocol, metric, against);
    const picked = !!against?.averages;
    const text = usual ? (picked ? `run ${against?.runId}: ${metric.format(usual)}` : `usually ${metric.format(usual)}`) : '';
    return <Box sx={{p: 1.5, border: 1, borderColor: 'divider', borderRadius: 1}}>
        <Typography variant="caption" color="text.secondary">{metric.label} (average)</Typography>
        <Typography variant="h6" component="div">{value ? metric.format(value) : '–'}</Typography>
        <VerdictNote verdict={compare(value, usual, metric.higherIsBetter)} same={picked ? 'same' : 'as usual'} text={text}/>
    </Box>;
};

interface MetricTilesProps {
    protocol: ITestProtocolData,
    metrics: Metric[],
    against: Comparison | null,
}

// MetricTiles compares only with real earlier runs: older masters averaged
// in the protocol itself and its runner
export const MetricTiles: React.FC<MetricTilesProps> = ({protocol, metrics, against}) =>
    <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 1.5}}>
        {metrics.map(m => <MetricTile key={m.key} protocol={protocol} metric={m} against={against}/>)}
    </Box>;
