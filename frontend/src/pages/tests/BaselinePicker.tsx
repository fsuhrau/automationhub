import React, {useEffect, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Chip,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControlLabel,
    InputAdornment,
    List,
    ListItemButton,
    Stack,
    Switch,
    TextField,
    Typography,
} from '@mui/material';
import {SearchRounded} from '@mui/icons-material';
import {getBaselines, IBaselineCandidate} from '../../services/test.run.service';
import {useLoader} from '../../hooks/useLoader';
import {formatDateTime} from '../../components/DateTime';
import {parseTags} from '../apps/bundlesModel';
import {averageOf, bundleLabel, Metric} from './protocolModel';

interface BaselinePickerProps {
    open: boolean,
    onClose: () => void,
    onPick: (c: IBaselineCandidate) => void,
    projectIdentifier: string,
    appId: number,
    testId: number,
    runId: number,
    protocolId: number,
    // the device of the protocol; runs on others are marked
    deviceId: number | null | undefined,
    // the run compared with now, if one is picked
    selectedRunId: number | null,
    metrics: Metric[],
}

// BaselinePicker lists the other runs of the test that measured performance,
// to pick one to compare with: by bundle name, version, tag or run id, on the
// same device or on any
// the master lists this many by default
const SHOWN = 50;

const BaselinePicker: React.FC<BaselinePickerProps> = (props) => {
    const {open, onClose, onPick, projectIdentifier, appId, testId, runId, protocolId, deviceId, selectedRunId, metrics} = props;
    const [text, setText] = useState('');
    const [q, setQ] = useState('');
    const [tag, setTag] = useState<string | null>(null);
    const [anyDevice, setAnyDevice] = useState(false);

    // search once typing pauses
    useEffect(() => {
        const t = setTimeout(() => setQ(text.trim()), 300);
        return () => clearTimeout(t);
    }, [text]);

    const {data, error, loading, retry} = useLoader(
        () => open ? getBaselines(projectIdentifier, appId, testId, runId, protocolId, {q, tag: tag ?? undefined, anyDevice}) : null,
        [open, projectIdentifier, appId, testId, runId, protocolId, q, tag, anyDevice]);
    const shown = metrics.slice(0, 3);

    return <Dialog open={open} onClose={onClose} fullWidth maxWidth="md" aria-labelledby="baseline-title">
        <DialogTitle id="baseline-title">Compare with another run</DialogTitle>
        <DialogContent dividers>
            <Stack spacing={1.5}>
                <Stack direction={{xs: 'column', sm: 'row'}} spacing={1.5} alignItems={{sm: 'center'}} useFlexGap>
                    <TextField size="small" autoFocus fullWidth value={text} onChange={e => setText(e.target.value)}
                               placeholder="Bundle name, version, tag, parameter or run id"
                               aria-label="Search runs"
                               slotProps={{input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>}}}/>
                    <FormControlLabel sx={{flexShrink: 0, mr: 0}} label="Any device"
                                      control={<Switch checked={anyDevice} onChange={e => setAnyDevice(e.target.checked)}/>}/>
                </Stack>
                {(data?.tags.length ?? 0) > 0 && <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap aria-label="Bundle tags">
                    {data!.tags.map(t => <Chip key={t} size="small" label={t}
                                               color={tag === t ? 'primary' : 'default'} variant={tag === t ? 'filled' : 'outlined'}
                                               onClick={() => setTag(tag === t ? null : t)}/>)}
                </Stack>}

                {error ? <Alert severity="error" action={<Button size="small" onClick={retry}>Retry</Button>}>
                        The runs could not be loaded.
                    </Alert>
                    : !data ? <Box sx={{display: 'flex', justifyContent: 'center', py: 4}}><CircularProgress size={24}/></Box>
                        : data.candidates.length === 0
                            ? <Typography variant="body2" color="text.secondary" sx={{py: 2}}>
                                No other run of this test measured performance{tag ? ` with a bundle tagged ${tag}` : ''}{q ? ` matching "${q}"` : ''}
                                {anyDevice ? '.' : ' on this device. Try any device.'}
                            </Typography>
                            : <List dense disablePadding sx={{opacity: loading ? 0.6 : 1}} aria-label="Runs to compare with">
                                {data.candidates.map(c => {
                                    const tags = parseTags(c.bundle?.tags);
                                    const selected = c.testRunId === selectedRunId;
                                    return <ListItemButton key={c.protocolId} selected={selected} onClick={() => onPick(c)}
                                                           sx={{borderRadius: 1, alignItems: 'flex-start', gap: 2}}>
                                        <Box sx={{flex: 1, minWidth: 0}}>
                                            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                                                <Typography variant="body2" sx={{fontWeight: 600}}>
                                                    Run {c.testRunId}{c.testRunId > runId && ' (later)'}
                                                </Typography>
                                                <Typography variant="body2">{bundleLabel(c.bundle) || 'unknown bundle'}</Typography>
                                                {tags.map(t => <Chip key={t} size="small" label={t} variant="outlined"/>)}
                                            </Stack>
                                            <Typography variant="caption" color="text.secondary" sx={{wordBreak: 'break-word'}}>
                                                {formatDateTime(c.startedAt, 'YYYY/MM/DD HH:mm')}
                                                {c.device && <> · {c.device}{c.deviceId !== deviceId && ' (other device)'}</>}
                                                {c.parameter && <> · {c.parameter}</>}
                                            </Typography>
                                        </Box>
                                        <Stack direction="row" spacing={2} sx={{flexShrink: 0, display: {xs: 'none', sm: 'flex'}}}>
                                            {shown.map(m => {
                                                const v = averageOf(c, m.key);
                                                return <Box key={m.key} sx={{textAlign: 'right', minWidth: 64}}>
                                                    <Typography variant="caption" color="text.secondary" component="div">{m.label}</Typography>
                                                    <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{v ? m.format(v) : '–'}</Typography>
                                                </Box>;
                                            })}
                                        </Stack>
                                    </ListItemButton>;
                                })}
                            </List>}
            </Stack>
        </DialogContent>
        <DialogActions>
            {data && data.candidates.length >= SHOWN && <Typography variant="caption" color="text.secondary" sx={{mr: 'auto', pl: 1}}>
                The newest {data.candidates.length}; search to find older runs.
            </Typography>}
            <Button onClick={onClose}>Cancel</Button>
        </DialogActions>
    </Dialog>;
};

export default BaselinePicker;
