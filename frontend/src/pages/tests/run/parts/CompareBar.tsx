import React from 'react';
import {Box, Button, Card, CardContent, Chip, Stack, Typography} from '@mui/material';
import {CompareArrowsRounded} from '@mui/icons-material';
import {formatDateTime} from '../../../../components/DateTime';
import {IBaseline} from '../../../../types/protocol.baseline';
import {parseTags} from '../../../apps/bundlesModel';
import {bundleLabel} from '../performanceModel';

const PickedRun: React.FC<{ baseline: IBaseline }> = ({baseline}) =>
    <>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true}>
            <Typography variant="body2">Compared with <b>run {baseline.testRunId}</b></Typography>
            {baseline.bundle && <Typography variant="body2">{bundleLabel(baseline.bundle)}</Typography>}
            {parseTags(baseline.bundle?.tags).map(t => <Chip key={t} size="small" variant="outlined" label={t}/>)}
        </Stack>
        <Typography variant="caption" color="text.secondary">
            {formatDateTime(baseline.startedAt, 'YYYY/MM/DD HH:mm')}{baseline.device && ` · ${baseline.device}`}
        </Typography>
    </>;

const EarlierRun: React.FC<{ earlierRunId: number | null }> = ({earlierRunId}) =>
    earlierRunId
        ? <Typography variant="body2">Compared with <b>run {earlierRunId}</b>, the latest earlier run on this device</Typography>
        : <Typography variant="body2" color="text.secondary">No earlier run on this device to compare with</Typography>;

interface CompareBarProps {
    baseline: IBaseline | null,
    loading: boolean,
    failed: boolean,
    earlierRunId: number | null,
    onChange: () => void,
    // back to the latest earlier run; null while no run is picked
    onReset: (() => void) | null,
}

const Status: React.FC<CompareBarProps> = ({baseline, loading, failed, earlierRunId, onReset}) => {
    if (failed) return <Typography variant="body2" color="error">The picked run could not be loaded.</Typography>;
    if (onReset && (loading || !baseline)) return <Typography variant="body2" color="text.secondary">Loading the picked run…</Typography>;
    return baseline ? <PickedRun baseline={baseline}/> : <EarlierRun earlierRunId={earlierRunId}/>;
};

// CompareBar says what the performance is compared with and offers to pick
// another run
export const CompareBar: React.FC<CompareBarProps> = (props) =>
    <Card variant="outlined">
        <CardContent sx={{py: 1.5, '&:last-child': {pb: 1.5}}}>
            <Stack direction={{xs: 'column', sm: 'row'}} spacing={1.5} alignItems={{sm: 'center'}} useFlexGap={true}>
                <CompareArrowsRounded color="action" sx={{display: {xs: 'none', sm: 'block'}}}/>
                <Box sx={{minWidth: 0}}><Status {...props}/></Box>
                <Stack direction="row" spacing={1} sx={{ml: {sm: 'auto'}, flexShrink: 0}}>
                    {props.onReset && <Button size="small" onClick={props.onReset}>Latest earlier run</Button>}
                    <Button size="small" variant="outlined" onClick={props.onChange}>Compare with…</Button>
                </Stack>
            </Stack>
        </CardContent>
    </Card>;
