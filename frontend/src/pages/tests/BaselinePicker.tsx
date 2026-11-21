import React from 'react';
import {
    Button, Chip, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, InputAdornment, Stack, Switch, TextField, Typography,
} from '@mui/material';
import {SearchRounded} from '@mui/icons-material';
import {IBaselineCandidate} from '../../types/protocol.baseline';
import {ProtocolRef} from '../../types/test.run.refs';
import {Metric} from './run/performanceModel';
import {Candidates} from './run/parts/BaselineCandidates';
import {BaselineCandidates, useBaselineCandidates} from './run/useBaselineCandidates';

interface BaselinePickerProps {
    open: boolean,
    onClose: () => void,
    onPick: (c: IBaselineCandidate) => void,
    // the protocol that is compared
    refs: ProtocolRef,
    // the device of the protocol; runs on others are marked
    deviceId: number | null | undefined,
    // the run compared with now, if one is picked
    selectedRunId: number | null,
    metrics: Metric[],
}

// the master lists this many by default
const SHOWN = 50;

const SearchBar: React.FC<{ search: BaselineCandidates }> = ({search}) =>
    <Stack direction={{xs: 'column', sm: 'row'}} spacing={1.5} alignItems={{sm: 'center'}} useFlexGap={true}>
        <TextField size="small" autoFocus={true} fullWidth={true} value={search.text} onChange={e => search.setText(e.target.value)}
            placeholder="Bundle name, version, tag, parameter or run id" aria-label="Search runs"
            slotProps={{input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>}}}/>
        <FormControlLabel sx={{flexShrink: 0, mr: 0}} label="Any device"
            control={<Switch checked={search.anyDevice} onChange={e => search.setAnyDevice(e.target.checked)}/>}/>
    </Stack>;

const TagChips: React.FC<{ search: BaselineCandidates }> = ({search}) => {
    const tags = search.list?.tags ?? [];
    if (tags.length === 0) return null;
    return <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap={true} aria-label="Bundle tags">
        {tags.map(t => <Chip key={t} size="small" label={t} color={search.tag === t ? 'primary' : 'default'}
            variant={search.tag === t ? 'filled' : 'outlined'}
            onClick={() => search.setTag(search.tag === t ? null : t)}/>)}
    </Stack>;
};

// BaselinePicker lists the other runs of the test that measured performance,
// to pick one to compare with: by bundle name, version, tag or run id, on the
// same device or on any
const BaselinePicker: React.FC<BaselinePickerProps> = ({open, onClose, onPick, refs, deviceId, selectedRunId, metrics}) => {
    const search = useBaselineCandidates(open, refs);
    const count = search.list?.candidates.length ?? 0;
    return <Dialog open={open} onClose={onClose} fullWidth={true} maxWidth="md" aria-labelledby="baseline-title">
        <DialogTitle id="baseline-title">Compare with another run</DialogTitle>
        <DialogContent dividers={true}>
            <Stack spacing={1.5}>
                <SearchBar search={search}/>
                <TagChips search={search}/>
                <Candidates search={search} selectedRunId={selectedRunId} runId={Number(refs.runId)} deviceId={deviceId}
                    metrics={metrics} onPick={onPick}/>
            </Stack>
        </DialogContent>
        <DialogActions>
            {count >= SHOWN && <Typography variant="caption" color="text.secondary" sx={{mr: 'auto', pl: 1}}>
                The newest {count}; search to find older runs.
            </Typography>}
            <Button onClick={onClose}>Cancel</Button>
        </DialogActions>
    </Dialog>;
};

export default BaselinePicker;
