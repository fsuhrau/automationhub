import React, {useEffect, useMemo, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Chip,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    FormControlLabel,
    IconButton,
    InputLabel,
    Link,
    List,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    MenuItem,
    Radio,
    RadioGroup,
    Select,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {AddRounded, DeleteOutlineRounded, EditRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import {getPlatformName, getPlatformTypes, PlatformType} from '../../types/platform.type.enum';
import {AppParameter, IAppData} from '../../types/app';
import {addAppParameter, createApp, getAllApps, removeAppParameter, updateApp, updateAppParameter} from '../../services/app.service';
import {HubStateActions} from '../../application/HubState';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useError} from '../../ErrorProvider';
import {ConfirmDialog, CopyButton, messageOf, Section} from './parts';
import {parameterErrors, parameterForm, ParameterForm, parameterRequest, tidyOptions} from './settingsModel';

// platforms that install a bundle need its identifier
const needsIdentifier = (p: PlatformType) => p === PlatformType.Android || p === PlatformType.iOS || p === PlatformType.Mac;
const identifierLabel = (p: PlatformType) => p === PlatformType.Android ? 'Package name' : 'Bundle identifier';

const NewAppDialog: React.FC<{ open: boolean, onClose: () => void, onCreated: (app: IAppData) => void }> = ({open, onClose, onCreated}) => {
    const {projectIdentifier} = useProjectContext();
    const [name, setName] = useState('');
    const [platform, setPlatform] = useState<PlatformType>(PlatformType.Android);
    const [identifier, setIdentifier] = useState('');
    const [tried, setTried] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [busy, setBusy] = useState(false);
    useEffect(() => {
        if (!open) return;
        setName('');
        setIdentifier('');
        setTried(false);
        setError(null);
    }, [open]);

    const nameError = name.trim() === '' ? 'A name is needed.' : undefined;
    const idError = needsIdentifier(platform) && identifier.trim() === '' ? `The ${identifierLabel(platform).toLowerCase()} is needed to install bundles.` : undefined;
    const submit = () => {
        setTried(true);
        if (nameError || idError) return;
        setBusy(true);
        createApp(projectIdentifier, {name: name.trim(), platform, identifier: identifier.trim()} as IAppData)
            .then(app => {
                onCreated(app);
                onClose();
            }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    return <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm" aria-labelledby="new-app-title">
        <Box component="form" noValidate onSubmit={e => {
            e.preventDefault();
            submit();
        }}>
            <DialogTitle id="new-app-title">Add an app</DialogTitle>
            <DialogContent>
                <Stack spacing={2} sx={{pt: 1}}>
                    <TextField size="small" autoFocus label="Name" placeholder="e.g. Game Android" value={name}
                               onChange={e => setName(e.target.value)} error={tried && !!nameError} helperText={tried ? nameError : undefined}/>
                    <FormControl size="small">
                        <InputLabel id="new-app-platform">Platform</InputLabel>
                        <Select labelId="new-app-platform" label="Platform" value={platform}
                                onChange={e => setPlatform(Number(e.target.value) as PlatformType)}>
                            {getPlatformTypes().map(p => <MenuItem key={p.id} value={Number(p.id)}>
                                <Stack direction="row" spacing={1} alignItems="center">
                                    <PlatformTypeIcon platformType={Number(p.id) as PlatformType} fontSize="small"/>
                                    <span>{p.id === String(PlatformType.Editor) ? 'Unity editor' : p.name}</span>
                                </Stack>
                            </MenuItem>)}
                        </Select>
                    </FormControl>
                    {needsIdentifier(platform) || platform === PlatformType.Windows || platform === PlatformType.Linux
                        ? <TextField size="small" label={identifierLabel(platform)} placeholder="com.example.game" value={identifier}
                                     onChange={e => setIdentifier(e.target.value)} error={tried && !!idError}
                                     helperText={tried && idError ? idError : 'The hub starts and stops the app by it.'}/>
                        : <Typography variant="body2" color="text.secondary">
                            {platform === PlatformType.Editor ? 'Unity editor apps run in the editor and need no bundle.' : 'Web apps run in a browser and need no bundle.'}
                        </Typography>}
                    {error && <Alert severity="error">{error}</Alert>}
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button type="submit" variant="contained" disabled={busy}>Add app</Button>
            </DialogActions>
        </Box>
    </Dialog>;
};

const ParameterDialog: React.FC<{
    open: boolean,
    app: IAppData,
    parameter: AppParameter | null,
    onClose: () => void,
}> = ({open, app, parameter, onClose}) => {
    const {projectIdentifier} = useProjectContext();
    const {dispatch} = useHubState();
    const [form, setForm] = useState<ParameterForm>(parameterForm(null));
    const [option, setOption] = useState('');
    const [tried, setTried] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [busy, setBusy] = useState(false);
    useEffect(() => {
        if (!open) return;
        setForm(parameterForm(parameter as never));
        setOption('');
        setTried(false);
        setError(null);
    }, [open, parameter]);

    const errors = parameterErrors(form, (app.parameter ?? []) as never, parameter?.id ?? null);
    const set = (change: Partial<ParameterForm>) => setForm(f => ({...f, ...change}));
    const addOption = () => {
        const add = option.split(',');
        set({options: tidyOptions([...form.options, ...add])});
        setOption('');
    };
    const options = tidyOptions(form.options);

    const submit = () => {
        setTried(true);
        if (Object.keys(errors).length > 0) return;
        setBusy(true);
        const body = parameterRequest(form) as never;
        const done = parameter
            ? updateAppParameter(projectIdentifier, app.id, parameter.id, body).then(p => dispatch({type: HubStateActions.UpdateAppParameter, payload: {...p, appId: app.id}}))
            : addAppParameter(projectIdentifier, app.id, body).then(p => dispatch({type: HubStateActions.AddAppParameter, payload: {...p, appId: app.id}}));
        done.then(onClose).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    return <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm" aria-labelledby="param-title">
        <Box component="form" noValidate onSubmit={e => {
            e.preventDefault();
            submit();
        }}>
            <DialogTitle id="param-title">{parameter ? `Edit ${parameter.name}` : 'Add a parameter'}</DialogTitle>
            <DialogContent>
                <Stack spacing={2} sx={{pt: 1}}>
                    <TextField size="small" autoFocus label="Name" placeholder="e.g. MarketId" value={form.name}
                               onChange={e => set({name: e.target.value})} error={tried && !!errors.name}
                               helperText={tried && errors.name ? errors.name : 'Passed to the app as NAME=value when a test runs.'}/>
                    <RadioGroup row value={form.kind} onChange={e => set({kind: e.target.value as ParameterForm['kind']})}>
                        <FormControlLabel value="string" control={<Radio size="small"/>} label="Text"/>
                        <FormControlLabel value="option" control={<Radio size="small"/>} label="Choice of options"/>
                    </RadioGroup>
                    {form.kind === 'option' && <Box>
                        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap sx={{mb: 1}}>
                            {options.map(o => <Chip key={o} size="small" label={o}
                                                    onDelete={() => set({options: options.filter(x => x !== o), defaultValue: form.defaultValue === o ? '' : form.defaultValue})}/>)}
                        </Stack>
                        <Stack direction="row" spacing={1} alignItems="flex-start">
                            <TextField size="small" label="Option" value={option} onChange={e => setOption(e.target.value)}
                                       error={tried && !!errors.options} helperText={tried && errors.options ? errors.options : 'Enter adds it; commas separate several.'}
                                       onKeyDown={e => {
                                           if (e.key === 'Enter') {
                                               e.preventDefault();
                                               addOption();
                                           }
                                       }}/>
                            <Button variant="outlined" onClick={addOption} disabled={!option.trim()}>Add</Button>
                        </Stack>
                    </Box>}
                    {form.kind === 'option'
                        ? <FormControl size="small" error={tried && !!errors.defaultValue}>
                            <InputLabel id="param-default">Default</InputLabel>
                            <Select labelId="param-default" label="Default" value={options.includes(form.defaultValue) ? form.defaultValue : ''}
                                    onChange={e => set({defaultValue: e.target.value})}>
                                <MenuItem value=""><em>none</em></MenuItem>
                                {options.map(o => <MenuItem key={o} value={o}>{o}</MenuItem>)}
                            </Select>
                        </FormControl>
                        : <TextField size="small" label="Default" value={form.defaultValue} onChange={e => set({defaultValue: e.target.value})}
                                     helperText="Filled in when a test is started; can be changed there."/>}
                    {error && <Alert severity="error">{error}</Alert>}
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button type="submit" variant="contained" disabled={busy}>{parameter ? 'Save' : 'Add parameter'}</Button>
            </DialogActions>
        </Box>
    </Dialog>;
};

// AppsGroup lists the project's apps and edits one: its name and the
// parameters a test run passes to it
const AppsGroup: React.FC = () => {
    const {state, dispatch} = useHubState();
    const {projectIdentifier} = useProjectContext();
    const {setError} = useError();
    const apps = useMemo(() => [...(state.apps ?? [])].sort((a, b) => a.platform - b.platform || a.name.localeCompare(b.name)), [state.apps]);
    const [selectedId, setSelectedId] = useState<number | null>(null);
    const app = apps.find(a => a.id === selectedId) ?? apps[0] ?? null;

    const [creating, setCreating] = useState(false);
    const [name, setName] = useState('');
    const [paramDialog, setParamDialog] = useState<{ parameter: AppParameter | null } | null>(null);
    const [deleting, setDeleting] = useState<AppParameter | null>(null);
    const [busy, setBusy] = useState(false);

    useEffect(() => {
        getAllApps(projectIdentifier).then(list => dispatch({type: HubStateActions.AppsUpdate, payload: list})).catch(e => setError(messageOf(e)));
    }, [projectIdentifier]);
    useEffect(() => setName(app?.name ?? ''), [app?.id, app?.name]);

    const saveName = () => {
        if (!app || !name.trim() || name.trim() === app.name) return;
        updateApp(projectIdentifier, app.id, {...app, name: name.trim()}).then(() => {
            dispatch({type: HubStateActions.AppAttributeUpdate, payload: {appId: app.id, attribute: 'name', value: name.trim()}});
        }).catch(e => setError(messageOf(e)));
    };

    const deleteParameter = () => {
        if (!app || !deleting) return;
        setBusy(true);
        removeAppParameter(projectIdentifier, app.id, deleting.id).then(() => {
            dispatch({type: HubStateActions.DeleteAppParameter, payload: {appId: app.id, id: deleting.id}});
            setDeleting(null);
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const parameters = app?.parameter ?? [];

    return <Section title="Apps" description="The builds this project tests, one per platform, with the parameters runs pass to them."
                    action={<Button variant="contained" size="small" startIcon={<AddRounded/>} onClick={() => setCreating(true)}>Add app</Button>}>
        {apps.length === 0
            ? <Typography variant="body2" color="text.secondary">No apps yet. Add one to create tests and upload bundles.</Typography>
            : <Box sx={{display: 'grid', gridTemplateColumns: {xs: '1fr', md: '260px 1fr'}, gap: 2}}>
                <List dense sx={{border: 1, borderColor: 'divider', borderRadius: 1, p: 0.5, alignSelf: 'start'}} aria-label="Apps">
                    {apps.map(a => <ListItemButton key={a.id} selected={a.id === app?.id} onClick={() => setSelectedId(a.id)} sx={{borderRadius: 1}}>
                        <ListItemIcon sx={{minWidth: 36}}><PlatformTypeIcon platformType={a.platform} fontSize="small"/></ListItemIcon>
                        <ListItemText primary={a.name} secondary={a.identifier || getPlatformName(a.platform)}
                                      slotProps={{secondary: {noWrap: true}}}/>
                    </ListItemButton>)}
                </List>

                {app && <Stack spacing={2.5} sx={{minWidth: 0}}>
                    <Box component="form" onSubmit={e => {
                        e.preventDefault();
                        saveName();
                    }}>
                        <Stack direction="row" spacing={1} alignItems="flex-start">
                            <TextField size="small" fullWidth label="Name" value={name} onChange={e => setName(e.target.value)}
                                       error={!name.trim()} helperText={!name.trim() ? 'A name is needed.' : undefined}/>
                            {name.trim() && name.trim() !== app.name && <>
                                <Button type="submit" variant="contained">Save</Button>
                                <Button onClick={() => setName(app.name)}>Undo</Button>
                            </>}
                        </Stack>
                    </Box>
                    <Box sx={{display: 'grid', gridTemplateColumns: 'max-content 1fr', columnGap: 2, rowGap: 0.75, alignItems: 'center'}}>
                        <Typography variant="body2" color="text.secondary">Platform</Typography>
                        <Stack direction="row" spacing={1} alignItems="center">
                            <PlatformTypeIcon platformType={app.platform} fontSize="small"/>
                            <Typography variant="body2">{app.platform === PlatformType.Editor ? 'Unity editor' : getPlatformName(app.platform)}</Typography>
                        </Stack>
                        {app.identifier && <>
                            <Typography variant="body2" color="text.secondary">{identifierLabel(app.platform)}</Typography>
                            <Stack direction="row" spacing={0.5} alignItems="center">
                                <Typography variant="body2" sx={{fontFamily: 'monospace', wordBreak: 'break-all'}}>{app.identifier}</Typography>
                                <CopyButton text={app.identifier} label="Copy"/>
                            </Stack>
                        </>}
                        <Typography variant="body2" color="text.secondary">Hub id</Typography>
                        <Typography variant="body2">{app.id}</Typography>
                        <Typography variant="body2" color="text.secondary">Go to</Typography>
                        <Stack direction="row" spacing={2}>
                            <Link component={RouterLink} to={`/project/${projectIdentifier}/app:${app.id}/tests`} variant="body2">Tests</Link>
                            <Link component={RouterLink} to={`/project/${projectIdentifier}/app:${app.id}/bundles`} variant="body2">Bundles</Link>
                        </Stack>
                    </Box>

                    <Box>
                        <Stack direction="row" alignItems="center" sx={{mb: 1}}>
                            <Box>
                                <Typography variant="body2" sx={{fontWeight: 600}}>Parameters</Typography>
                                <Typography variant="caption" color="text.secondary">Asked for when a test of this app is started, and passed to the app.</Typography>
                            </Box>
                            <Button size="small" startIcon={<AddRounded/>} sx={{ml: 'auto'}} onClick={() => setParamDialog({parameter: null})}>Add parameter</Button>
                        </Stack>
                        {parameters.length === 0
                            ? <Typography variant="body2" color="text.secondary">None.</Typography>
                            : <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="Parameters">
                                <TableHead><TableRow>
                                    <TableCell>Name</TableCell><TableCell>Kind</TableCell><TableCell>Default</TableCell><TableCell/>
                                </TableRow></TableHead>
                                <TableBody>
                                    {parameters.map(p => {
                                        const opts = p.type.type === 'option' ? (p.type as { options?: string[] }).options ?? [] : [];
                                        return <TableRow key={p.id}>
                                            <TableCell sx={{fontFamily: 'monospace'}}>{p.name}</TableCell>
                                            <TableCell>
                                                {p.type.type === 'option'
                                                    ? <Tooltip title={opts.join(', ')}><span>Choice of {opts.length}</span></Tooltip>
                                                    : 'Text'}
                                            </TableCell>
                                            <TableCell sx={{wordBreak: 'break-all'}}>{p.type.defaultValue || <Typography variant="body2" color="text.secondary" component="span">none</Typography>}</TableCell>
                                            <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>
                                                <Tooltip title="Edit"><IconButton size="small" aria-label={`Edit ${p.name}`} onClick={() => setParamDialog({parameter: p})}><EditRounded fontSize="small"/></IconButton></Tooltip>
                                                <Tooltip title="Delete"><IconButton size="small" aria-label={`Delete ${p.name}`} onClick={() => setDeleting(p)}><DeleteOutlineRounded fontSize="small"/></IconButton></Tooltip>
                                            </TableCell>
                                        </TableRow>;
                                    })}
                                </TableBody>
                            </Table></Box>}
                    </Box>
                </Stack>}
            </Box>}

        <NewAppDialog open={creating} onClose={() => setCreating(false)} onCreated={created => {
            dispatch({type: HubStateActions.AppAdd, payload: created});
            setSelectedId(created.id);
        }}/>
        {app && <ParameterDialog open={!!paramDialog} app={app} parameter={paramDialog?.parameter ?? null} onClose={() => setParamDialog(null)}/>}
        <ConfirmDialog open={!!deleting} title={`Delete ${deleting?.name}?`} confirm="Delete parameter" busy={busy}
                       onConfirm={deleteParameter} onClose={() => setDeleting(null)}>
            New runs of {app?.name} won't ask for it or pass it. Runs that already used it keep their values.
        </ConfirmDialog>
    </Section>;
};

export default AppsGroup;
