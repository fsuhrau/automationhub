import React from 'react';
import {Alert, Box, Button, Chip, CircularProgress, List, ListItemButton, Stack, Typography} from '@mui/material';
import {formatDateTime} from '../../../../components/DateTime';
import {IBaselineCandidate} from '../../../../types/protocol.baseline';
import {parseTags} from '../../../apps/bundlesModel';
import {averageOf, bundleLabel, Metric} from '../performanceModel';
import {BaselineCandidates} from '../useBaselineCandidates';

interface CandidateProps {
    candidate: IBaselineCandidate,
    runId: number,
    deviceId: number | null | undefined,
    selected: boolean,
    metrics: Metric[],
    onPick: (c: IBaselineCandidate) => void,
}

const Averages: React.FC<{ candidate: IBaselineCandidate, metrics: Metric[] }> = ({candidate, metrics}) =>
    <Stack direction="row" spacing={2} sx={{flexShrink: 0, display: {xs: 'none', sm: 'flex'}}}>
        {metrics.slice(0, 3).map(m => {
            const v = averageOf(candidate, m.key);
            return <Box key={m.key} sx={{textAlign: 'right', minWidth: 64}}>
                <Typography variant="caption" color="text.secondary" component="div">{m.label}</Typography>
                <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{v ? m.format(v) : '–'}</Typography>
            </Box>;
        })}
    </Stack>;

const CandidateRow: React.FC<CandidateProps> = ({candidate: c, runId, deviceId, selected, metrics, onPick}) =>
    <ListItemButton selected={selected} onClick={() => onPick(c)} sx={{borderRadius: 1, alignItems: 'flex-start', gap: 2}}>
        <Box sx={{flex: 1, minWidth: 0}}>
            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true}>
                <Typography variant="body2" sx={{fontWeight: 600}}>Run {c.testRunId}{c.testRunId > runId && ' (later)'}</Typography>
                <Typography variant="body2">{bundleLabel(c.bundle) || 'unknown bundle'}</Typography>
                {parseTags(c.bundle?.tags).map(t => <Chip key={t} size="small" label={t} variant="outlined"/>)}
            </Stack>
            <Typography variant="caption" color="text.secondary" sx={{wordBreak: 'break-word'}}>
                {formatDateTime(c.startedAt, 'YYYY/MM/DD HH:mm')}
                {c.device && <> · {c.device}{c.deviceId !== deviceId && ' (other device)'}</>}
                {c.parameter && <> · {c.parameter}</>}
            </Typography>
        </Box>
        <Averages candidate={c} metrics={metrics}/>
    </ListItemButton>;

const Empty: React.FC<{ search: BaselineCandidates }> = ({search}) =>
    <Typography variant="body2" color="text.secondary" sx={{py: 2}}>
        No other run of this test measured performance{search.tag ? ` with a bundle tagged ${search.tag}` : ''}{search.q ? ` matching "${search.q}"` : ''}
        {search.anyDevice ? '.' : ' on this device. Try any device.'}
    </Typography>;

interface CandidatesProps extends Omit<CandidateProps, 'candidate' | 'selected'> {
    search: BaselineCandidates,
    selectedRunId: number | null,
}

// the candidate runs, or why there are none yet
export const Candidates: React.FC<CandidatesProps> = ({search, selectedRunId, ...row}) => {
    const {list, error, loading, retry} = search;
    if (error) {
        return <Alert severity="error" action={<Button size="small" onClick={retry}>Retry</Button>}>The runs could not be loaded.</Alert>;
    }
    if (!list) return <Box sx={{display: 'flex', justifyContent: 'center', py: 4}}><CircularProgress size={24}/></Box>;
    if (list.candidates.length === 0) return <Empty search={search}/>;
    return <List dense={true} disablePadding={true} sx={{opacity: loading ? 0.6 : 1}} aria-label="Runs to compare with">
        {list.candidates.map(c => <CandidateRow key={c.protocolId} {...row} candidate={c} selected={c.testRunId === selectedRunId}/>)}
    </List>;
};
