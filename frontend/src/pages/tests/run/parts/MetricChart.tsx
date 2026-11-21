import React from 'react';
import {Box, Typography} from '@mui/material';
import {useColorScheme, useTheme} from '@mui/material/styles';
import {LineChart} from '@mui/x-charts/LineChart';
import {gray} from '../../../../shared-theme/themePrimitives';
import ITestProtocolData from '../../../../types/test.protocol';
import IProtocolPerformanceEntryData from '../../../../types/protocol.performance.entry';
import {formatRuntime} from '../logModel';
import {Metric} from '../performanceModel';
import {Comparison} from '../protocolViewModel';

// the value of the metric at the first time each checkpoint was reached
const valuesByCheckpoint = (perf: IProtocolPerformanceEntryData[], metric: Metric): Map<string, number> => {
    const values = new Map<string, number>();
    perf.forEach(p => {
        if (!values.has(p.checkpoint) && typeof p[metric.key] === 'number') values.set(p.checkpoint, p[metric.key]);
    });
    return values;
};

interface MetricChartProps {
    metric: Metric,
    protocol: ITestProtocolData,
    earlier: Comparison | null,
}

// the theme uses CSS variables, so theme.palette always holds the light
// scheme; the compared run's grey has to follow the active scheme itself
const useDark = (): boolean => {
    const {mode, systemMode} = useColorScheme();
    return (mode === 'system' ? systemMode : mode) === 'dark';
};

// MetricChart plots one metric over the checkpoints of the test, with the
// compared run as a dashed reference line where it reached the same checkpoint
export const MetricChart: React.FC<MetricChartProps> = ({metric, protocol, earlier}) => {
    const theme = useTheme();
    const dark = useDark();
    const perf = (protocol.performance ?? []).filter(p => p.checkpoint !== 'schedule');
    const before = valuesByCheckpoint(earlier?.performance ?? [], metric);
    const earlierData = perf.map(p => before.get(p.checkpoint) ?? null);
    const hasEarlier = earlierData.some(v => v !== null);
    const format = (v: number | null): string => v === null ? '–' : metric.format(v);
    const series = [
        {id: 'current', data: perf.map(p => p[metric.key] ?? null), label: 'This run', color: theme.palette.primary.main,
            showMark: perf.length <= 20, valueFormatter: format, curve: 'linear' as const},
        ...(hasEarlier ? [{id: 'earlier', data: earlierData, label: `Run ${earlier?.runId}`, color: dark ? gray[300] : gray[500],
            showMark: false, valueFormatter: format, connectNulls: true, curve: 'linear' as const}] : []),
    ];
    return <Box>
        <LineChart height={220} skipAnimation={true}
            margin={{top: hasEarlier ? 40 : 16, right: 16, bottom: 28, left: 56}}
            xAxis={[{scaleType: 'point', data: perf.map((_, i) => i), valueFormatter: (i: number, ctx) =>
                           ctx.location === 'tooltip' ? `${formatRuntime(perf[i]?.runtime)} · ${perf[i]?.checkpoint}` : `${i + 1}`}]}
            yAxis={[{valueFormatter: (v: number) => metric.format(v).replace(/ (fps|%)$/, '')}]}
            series={series}
            slotProps={{legend: {hidden: !hasEarlier}}}
            sx={{'& .MuiLineElement-series-earlier': {strokeDasharray: '5 4'}}}/>
        <Typography variant="caption" color="text.secondary">Checkpoints in order; hover for their names.</Typography>
    </Box>;
};
