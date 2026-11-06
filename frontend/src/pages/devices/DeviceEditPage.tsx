import React, {useEffect, useMemo, useState} from 'react';
import {
    Box,
    Button,
    Card,
    CardContent,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    IconButton,
    Stack,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {AddRounded, DeleteOutlineRounded} from '@mui/icons-material';
import {useNavigate} from 'react-router-dom';
import IDeviceData from '../../types/device';
import {updateDevice} from '../../services/device.service';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {useError} from '../../ErrorProvider';
import {HubStateActions} from '../../application/HubState';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import {EditForm, isDirty, keyErrors, ParameterRow, toRequest} from './deviceEditModel';

interface DeviceEditPageProps {
    device: IDeviceData
}

let nextRow = 0;
const rowId = () => `row-${nextRow++}`;

const formOf = (device: IDeviceData): EditForm => ({
    alias: device.alias ?? '',
    parameters: (device.customParameter ?? []).map(p => ({rowId: rowId(), id: p.id, key: p.key, value: p.value})),
});

const Section: React.FC<{ title: string, description: React.ReactNode, children: React.ReactNode }> = ({title, description, children}) =>
    <Card variant="outlined">
        <CardContent>
            <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>{description}</Typography>
            {children}
        </CardContent>
    </Card>;

// DeviceEditPage changes what a person sets on a device: the alias and the
// custom parameters. The rest is reported by the device and isn't editable.
const DeviceEditPage: React.FC<DeviceEditPageProps> = ({device}) => {
    const navigate = useNavigate();
    const {projectIdentifier} = useProjectContext();
    const {state, dispatch} = useHubState();
    const {setError} = useError();

    const initial = useMemo(() => formOf(device), [device]);
    const [form, setForm] = useState<EditForm>(initial);
    const [saving, setSaving] = useState(false);
    const [leaving, setLeaving] = useState(false);

    const errors = keyErrors(form.parameters);
    const dirty = isDirty(initial, form);
    const valid = errors.size === 0;
    const detailsUrl = `/project/${projectIdentifier}/device/${device.id}`;

    // closing or reloading the tab with unsaved changes asks first
    useEffect(() => {
        if (!dirty) return;
        const warn = (e: BeforeUnloadEvent) => e.preventDefault();
        window.addEventListener('beforeunload', warn);
        return () => window.removeEventListener('beforeunload', warn);
    }, [dirty]);

    const setRow = (id: string, change: Partial<ParameterRow>) =>
        setForm(f => ({...f, parameters: f.parameters.map(p => p.rowId === id ? {...p, ...change} : p)}));
    const addRow = () => {
        const id = rowId();
        setForm(f => ({...f, parameters: [...f.parameters, {rowId: id, key: '', value: ''}]}));
        // focus the new key field once it is rendered
        setTimeout(() => document.getElementById(`param-key-${id}`)?.focus());
    };
    const removeRow = (id: string) => setForm(f => ({...f, parameters: f.parameters.filter(p => p.rowId !== id)}));

    const save = () => {
        if (!valid || !dirty) return;
        setSaving(true);
        updateDevice(projectIdentifier, device.id, toRequest(form)).then(saved => {
            // the device list keeps its live state; only what was edited changes
            const listed = state.devices?.find(d => d.id === device.id);
            if (listed && saved?.id) {
                dispatch({type: HubStateActions.UpdateDevice, payload: {...listed, alias: saved.alias, customParameter: saved.customParameter}});
            }
            navigate(detailsUrl);
        }).catch(e => {
            setError(e);
            setSaving(false);
        });
    };

    const cancel = () => dirty ? setLeaving(true) : navigate(detailsUrl);

    return (
        // full width like the other pages, so it lines up with the breadcrumbs;
        // the form itself stays narrow enough to read
        <Box component="form" sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2, '& > *': {maxWidth: 900}}} noValidate
             onSubmit={e => {
                 e.preventDefault();
                 save();
             }}
             onKeyDown={e => {
                 // Cmd/Ctrl+S saves, as in an editor
                 if ((e.metaKey || e.ctrlKey) && e.key === 's') {
                     e.preventDefault();
                     save();
                 }
             }}>
            <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} useFlexGap alignItems={{sm: 'center'}} sx={{mb: 2}}>
                <Stack direction="row" spacing={1.5} alignItems="center" sx={{minWidth: 0}}>
                    <Box sx={{display: 'flex', p: 1, borderRadius: 1, border: 1, borderColor: 'divider'}}>
                        <PlatformTypeIcon platformType={device.platformType}/>
                    </Box>
                    <Box sx={{minWidth: 0}}>
                        <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>
                            Edit {device.alias || device.name}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                            {dirty ? 'Unsaved changes' : 'No changes yet'}
                        </Typography>
                    </Box>
                </Stack>
                <Stack direction="row" spacing={1} sx={{ml: {sm: 'auto'}}}>
                    <Button onClick={cancel} disabled={saving}>Cancel</Button>
                    <Tooltip title={!valid ? 'Fix the marked parameters first' : !dirty ? 'Nothing to save yet' : 'Save (⌘S / Ctrl+S)'}>
                        <span><Button type="submit" variant="contained" disabled={!valid || !dirty || saving}
                                      startIcon={saving ? <CircularProgress size={16} color="inherit"/> : undefined}>
                            Save
                        </Button></span>
                    </Tooltip>
                </Stack>
            </Stack>

            <Stack spacing={2}>
                <Section title="Alias" description={<>
                    Shown instead of the device name <Box component="span" sx={{fontFamily: 'monospace'}}>{device.name}</Box> in
                    lists, test results and the device pool. Leave it empty to use the device name.
                </>}>
                    <TextField size="small" fullWidth label="Alias" placeholder={device.name} value={form.alias}
                               onChange={e => setForm(f => ({...f, alias: e.target.value}))}
                               slotProps={{htmlInput: {maxLength: 100}}}/>
                </Section>

                <Section title="Custom parameters" description={<>
                    Values the hub uses for this device and passes to tests, for example <Box component="span" sx={{fontFamily: 'monospace'}}>PIN</Box> to
                    unlock an Android phone. Keys must be unique.
                </>}>
                    {form.parameters.length === 0
                        ? <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>No custom parameters.</Typography>
                        : <Box sx={{display: 'grid', gridTemplateColumns: {xs: '1fr', sm: 'minmax(140px, 2fr) 3fr auto'}, gap: 1, alignItems: 'start', mb: 1.5}}
                               role="group" aria-label="Custom parameters">
                            {form.parameters.map((p, i) => <React.Fragment key={p.rowId}>
                                <TextField size="small" id={`param-key-${p.rowId}`} placeholder="Key" value={p.key}
                                           error={errors.has(i)} helperText={errors.get(i)}
                                           onChange={e => setRow(p.rowId, {key: e.target.value})}
                                           slotProps={{htmlInput: {'aria-label': `Key of parameter ${i + 1}`, spellCheck: false}}}/>
                                <TextField size="small" placeholder="Value" value={p.value}
                                           onChange={e => setRow(p.rowId, {value: e.target.value})}
                                           slotProps={{htmlInput: {'aria-label': `Value of ${p.key || `parameter ${i + 1}`}`, spellCheck: false}}}/>
                                <Tooltip title="Remove">
                                    <IconButton aria-label={`Remove ${p.key || `parameter ${i + 1}`}`} onClick={() => removeRow(p.rowId)}>
                                        <DeleteOutlineRounded fontSize="small"/>
                                    </IconButton>
                                </Tooltip>
                            </React.Fragment>)}
                        </Box>}
                    <Button size="small" startIcon={<AddRounded/>} onClick={addRow}>Add parameter</Button>
                </Section>

                <Typography variant="body2" color="text.secondary">
                    The name, system, type and connection are reported by the device and its node, so they can't be edited here.
                </Typography>
            </Stack>

            <Dialog open={leaving} onClose={() => setLeaving(false)} aria-labelledby="discard-title">
                <DialogTitle id="discard-title">Discard your changes?</DialogTitle>
                <DialogContent>
                    <DialogContentText>The alias and parameter changes on this page aren't saved yet.</DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setLeaving(false)}>Keep editing</Button>
                    <Button color="warning" onClick={() => navigate(detailsUrl)}>Discard</Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default DeviceEditPage;
