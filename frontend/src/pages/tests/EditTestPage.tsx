import React, {useEffect, useMemo, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Checkbox,
    Chip,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    FormControlLabel,
    IconButton,
    InputAdornment,
    Radio,
    RadioGroup,
    Stack,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {CloseRounded, RefreshRounded, SearchRounded} from '@mui/icons-material';
import {useNavigate} from 'react-router-dom';
import ITestData from '../../types/test';
import IAppFunctionData from '../../types/app.function';
import IDeviceData from '../../types/device';
import {TestType} from '../../types/test.type.enum';
import {TestExecutionType} from '../../types/test.execution.type.enum';
import {UnityTestCategory} from '../../types/unity.test.category.type.enum';
import {updateTest} from '../../services/test.service';
import {getAllDevices} from '../../services/device.service';
import {getTestFunctions} from '../../services/unity.service';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {useError} from '../../ErrorProvider';
import DeviceStatus from '../../components/DeviceStatus';
import {
    addFunctions,
    formOf,
    functionKey,
    groupByClass,
    isDirty,
    matchesFunction,
    parsePastedTests,
    shortClass,
    shortMethod,
    splitCategories,
    TestForm,
    toRequest,
    validate,
} from './testEditModel';

interface TestContentProps {
    test: ITestData
}

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};

const Section: React.FC<{ title: string, description?: React.ReactNode, children: React.ReactNode }> = ({title, description, children}) =>
    <Card variant="outlined">
        <CardContent>
            <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
            {description && <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>{description}</Typography>}
            {children}
        </CardContent>
    </Card>;

// a radio choice with a line that says what it means
const Choice: React.FC<{ value: string, label: string, hint: string }> = ({value, label, hint}) =>
    <FormControlLabel value={value} control={<Radio size="small"/>} sx={{alignItems: 'flex-start', mb: 0.5, '& .MuiRadio-root': {pt: 0.25}}}
                      label={<Box>
                          <Typography variant="body2" sx={{fontWeight: 600}}>{label}</Typography>
                          <Typography variant="caption" color="text.secondary">{hint}</Typography>
                      </Box>}/>;

// TestPicker shows the picked Unity tests and adds more from the Unity
// editor on this computer, or from a pasted list
const TestPicker: React.FC<{ picked: IAppFunctionData[], onChange: (f: IAppFunctionData[]) => void, error?: string }> = ({picked, onChange, error}) => {
    const [available, setAvailable] = useState<IAppFunctionData[] | null>(null);
    const [loading, setLoading] = useState(false);
    const [unreachable, setUnreachable] = useState(false);
    const [search, setSearch] = useState('');
    const [checked, setChecked] = useState<Set<string>>(new Set());
    const [paste, setPaste] = useState('');
    const [pasteError, setPasteError] = useState<string | null>(null);

    const fetchFromUnity = () => {
        setLoading(true);
        setUnreachable(false);
        getTestFunctions().then(setAvailable).catch(() => setUnreachable(true)).finally(() => setLoading(false));
    };
    useEffect(fetchFromUnity, []);

    const pickedKeys = useMemo(() => new Set(picked.map(functionKey)), [picked]);
    const choices = (available ?? []).filter(f => !pickedKeys.has(functionKey(f)) && matchesFunction(f, search));
    const add = (list: IAppFunctionData[]) => {
        onChange(addFunctions(picked, list));
        setChecked(new Set());
    };
    const usePaste = () => {
        const result = parsePastedTests(paste);
        if ('error' in result) {
            setPasteError(result.error);
            return;
        }
        setPasteError(null);
        setAvailable(result.functions);
        setPaste('');
    };

    return <Stack spacing={2}>
        <Box>
            <Typography variant="body2" sx={{fontWeight: 600, mb: 0.5}}>Picked tests ({picked.length})</Typography>
            {error && <Typography variant="body2" color="error" sx={{mb: 1}}>{error}</Typography>}
            {picked.length === 0
                ? <Typography variant="body2" color="text.secondary">None yet; add them from the list below.</Typography>
                : <Box sx={{maxHeight: 280, overflowY: 'auto', border: 1, borderColor: 'divider', borderRadius: 1, p: 1}}>
                    {groupByClass(picked).map(([cls, fns]) => <Box key={cls} sx={{mb: 1}}>
                        <Typography variant="caption" color="text.secondary" title={cls}>{shortClass(cls)}</Typography>
                        {fns.map(f => <Stack key={functionKey(f)} direction="row" alignItems="center" sx={{pl: 1}}>
                            <Typography variant="body2" sx={{fontFamily: 'monospace', fontSize: 13, flex: 1, wordBreak: 'break-all'}}>{shortMethod(f.method)}</Typography>
                            <Tooltip title="Remove">
                                <IconButton size="small" aria-label={`Remove ${shortMethod(f.method)}`}
                                            onClick={() => onChange(picked.filter(p => functionKey(p) !== functionKey(f)))}>
                                    <CloseRounded fontSize="small"/>
                                </IconButton>
                            </Tooltip>
                        </Stack>)}
                    </Box>)}
                </Box>}
        </Box>

        <Box>
            <Stack direction="row" alignItems="center" spacing={1} sx={{mb: 1}}>
                <Typography variant="body2" sx={{fontWeight: 600}}>Add tests</Typography>
                <Tooltip title="Ask the Unity editor again">
                    <span><IconButton size="small" aria-label="Load the tests from Unity again" onClick={fetchFromUnity} disabled={loading}>
                        <RefreshRounded fontSize="small"/>
                    </IconButton></span>
                </Tooltip>
            </Stack>
            {loading
                ? <Stack direction="row" spacing={1} alignItems="center"><CircularProgress size={16}/>
                    <Typography variant="body2" color="text.secondary">Asking the Unity editor on this computer for its tests…</Typography></Stack>
                : available === null || (unreachable && available.length === 0)
                    ? <Stack spacing={1}>
                        <Alert severity="info">
                            The Unity editor on this computer didn't answer. Open the project in Unity with the Automation Hub plugin
                            (Tools › Automation Hub › Settings) and try again, or copy the test list there and paste it below.
                        </Alert>
                        <TextField multiline minRows={3} maxRows={8} fullWidth size="small" placeholder='[{"assembly": "…", "class": "…", "method": "…"}]'
                                   value={paste} onChange={e => setPaste(e.target.value)} error={!!pasteError} helperText={pasteError ?? undefined}
                                   slotProps={{htmlInput: {'aria-label': 'Pasted test list', spellCheck: false}}}/>
                        <Box><Button size="small" variant="outlined" disabled={!paste.trim()} onClick={usePaste}>Use this list</Button></Box>
                    </Stack>
                    : <Stack spacing={1}>
                        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                            <TextField size="small" placeholder="Find a test or class" value={search} onChange={e => setSearch(e.target.value)}
                                       sx={{flex: 1, minWidth: 200}}
                                       slotProps={{
                                           input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                                           htmlInput: {'aria-label': 'Find a test to add'},
                                       }}/>
                            <Button size="small" variant="contained" disabled={checked.size === 0}
                                    onClick={() => add(choices.filter(f => checked.has(functionKey(f))))}>Add {checked.size || ''} picked</Button>
                            <Button size="small" disabled={choices.length === 0} onClick={() => add(choices)}>
                                Add all {choices.length}{search ? ' found' : ''}
                            </Button>
                        </Stack>
                        {choices.length === 0
                            ? <Typography variant="body2" color="text.secondary">
                                {(available ?? []).length === 0 ? 'Unity reported no tests.' : search ? 'No test matches.' : 'Every test is picked.'}
                            </Typography>
                            : <Box sx={{maxHeight: 320, overflowY: 'auto', border: 1, borderColor: 'divider', borderRadius: 1, p: 1}}
                                   role="group" aria-label="Tests to add">
                                {groupByClass(choices).map(([cls, fns]) => <Box key={cls} sx={{mb: 1}}>
                                    <Typography variant="caption" color="text.secondary" title={cls}>{shortClass(cls)}</Typography>
                                    {fns.map(f => <FormControlLabel key={functionKey(f)} sx={{display: 'flex', pl: 0.5, mr: 0}}
                                                                    control={<Checkbox size="small" checked={checked.has(functionKey(f))}
                                                                                       onChange={(_, on) => setChecked(prev => {
                                                                                           const next = new Set(prev);
                                                                                           if (on) next.add(functionKey(f)); else next.delete(functionKey(f));
                                                                                           return next;
                                                                                       })}/>}
                                                                    label={<Typography variant="body2" sx={{fontFamily: 'monospace', fontSize: 13}}>{shortMethod(f.method)}</Typography>}/>)}
                                </Box>)}
                            </Box>}
                    </Stack>}
        </Box>
    </Stack>;
};

// EditTestPage changes a test: its name, what it runs and where
const EditTestPage: React.FC<TestContentProps> = ({test}) => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const {setError} = useError();
    const navigate = useNavigate();
    const app = state.apps?.find(a => a.id === appId);
    const unity = test.testConfig.type === TestType.Unity && !!test.testConfig.unity;
    const testUrl = `/project/${projectIdentifier}/app:${appId}/test/${test.id}`;

    const initial = useMemo(() => formOf(test), [test]);
    const [form, setForm] = useState<TestForm>(initial);
    const [category, setCategory] = useState('');
    const [devices, setDevices] = useState<IDeviceData[] | null>(null);
    const [saving, setSaving] = useState(false);
    const [leaving, setLeaving] = useState(false);
    const [tried, setTried] = useState(false);

    const errors = validate(form, unity);
    const valid = Object.keys(errors).length === 0;
    const dirty = isDirty(initial, form);
    const set = (change: Partial<TestForm>) => setForm(f => ({...f, ...change}));

    useEffect(() => {
        getAllDevices(projectIdentifier, app?.platform).then(setDevices).catch(e => setError(messageOf(e)));
    }, [projectIdentifier, app?.platform]);

    useEffect(() => {
        if (!dirty) return;
        const warn = (e: BeforeUnloadEvent) => e.preventDefault();
        window.addEventListener('beforeunload', warn);
        return () => window.removeEventListener('beforeunload', warn);
    }, [dirty]);

    const addCategory = () => {
        const add = splitCategories(category);
        if (add.length === 0) return;
        set({categories: splitCategories([...form.categories, ...add].join(','))});
        setCategory('');
    };

    const save = () => {
        setTried(true);
        if (!valid || !dirty || appId === null) return;
        setSaving(true);
        updateTest(projectIdentifier, appId, test.id, toRequest(form)).then(() => navigate(testUrl))
            .catch(e => {
                setError(messageOf(e));
                setSaving(false);
            });
    };

    // devices picked earlier that the list doesn't show (other platform, gone)
    const missingDevices = form.devices.filter(id => devices && !devices.some(d => d.id === id));

    return (
        <Box component="form" noValidate sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2, '& > *': {maxWidth: 960}}}
             onSubmit={e => {
                 e.preventDefault();
                 save();
             }}
             onKeyDown={e => {
                 if ((e.metaKey || e.ctrlKey) && e.key === 's') {
                     e.preventDefault();
                     save();
                 }
             }}>
            <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} useFlexGap alignItems={{sm: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>Edit {test.name}</Typography>
                    <Typography variant="body2" color="text.secondary">
                        {TestType[test.testConfig.type]} test · {dirty ? 'unsaved changes' : 'no changes yet'}
                    </Typography>
                </Box>
                <Stack direction="row" spacing={1} sx={{ml: {sm: 'auto'}}}>
                    <Button onClick={() => dirty ? setLeaving(true) : navigate(testUrl)} disabled={saving}>Cancel</Button>
                    <Tooltip title={!dirty ? 'Nothing to save yet' : !valid ? 'Fix the marked fields first' : 'Save (⌘S / Ctrl+S)'}>
                        <span><Button type="submit" variant="contained" disabled={!dirty || saving || (tried && !valid)}
                                      startIcon={saving ? <CircularProgress size={16} color="inherit"/> : undefined}>Save</Button></span>
                    </Tooltip>
                </Stack>
            </Stack>

            {tried && !valid && <Alert severity="error" sx={{mb: 2}}>
                {Object.values(errors).join(' ')}
            </Alert>}

            <Stack spacing={2}>
                <Section title="Name">
                    <TextField size="small" fullWidth required label="Name" value={form.name} onChange={e => set({name: e.target.value})}
                               error={tried && !!errors.name} helperText={tried ? errors.name : undefined}/>
                </Section>

                {unity && <Section title="What it runs" description="Which Unity tests of the app the test starts.">
                    <RadioGroup value={String(form.mode)} onChange={e => set({mode: Number(e.target.value) as UnityTestCategory})}>
                        <Choice value={String(UnityTestCategory.RunAllTests)} label="All tests" hint="Every Unity test the app has, including new ones."/>
                        <Choice value={String(UnityTestCategory.RunAllOfCategory)} label="Tests of categories" hint="The Unity tests with one of the named categories."/>
                        <Choice value={String(UnityTestCategory.RunSelectedTestsOnly)} label="Picked tests" hint="Only the tests picked here."/>
                    </RadioGroup>

                    {form.mode === UnityTestCategory.RunAllOfCategory && <Box sx={{mt: 1.5}}>
                        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap sx={{mb: 1}}>
                            {form.categories.map(c => <Chip key={c} size="small" label={c}
                                                            onDelete={() => set({categories: form.categories.filter(x => x !== c)})}/>)}
                        </Stack>
                        <Stack direction="row" spacing={1}>
                            <TextField size="small" label="Category" value={category} onChange={e => setCategory(e.target.value)}
                                       error={tried && !!errors.categories} helperText={tried ? errors.categories : 'Enter adds it; commas separate several.'}
                                       onKeyDown={e => {
                                           if (e.key === 'Enter') {
                                               e.preventDefault();
                                               addCategory();
                                           }
                                       }}/>
                            <Box><Button variant="outlined" onClick={addCategory} disabled={!category.trim()}>Add</Button></Box>
                        </Stack>
                    </Box>}

                    {form.mode === UnityTestCategory.RunSelectedTestsOnly && <Box sx={{mt: 1.5}}>
                        <TestPicker picked={form.functions} onChange={functions => set({functions})} error={tried ? errors.tests : undefined}/>
                    </Box>}
                </Section>}

                <Section title="Where it runs">
                    <RadioGroup value={form.allDevices ? 'all' : 'picked'} onChange={e => set({allDevices: e.target.value === 'all'})}>
                        <Choice value="all" label="All devices" hint="Every device in the device pool."/>
                        <Choice value="picked" label="Picked devices" hint="Only the devices picked below."/>
                    </RadioGroup>
                    {!form.allDevices && <Box sx={{mt: 1, mb: 2}}>
                        {tried && errors.devices && <Typography variant="body2" color="error" sx={{mb: 1}}>{errors.devices}</Typography>}
                        {devices === null
                            ? <CircularProgress size={18}/>
                            : devices.length === 0
                                ? <Alert severity="warning">No {app ? `${app.name} ` : ''}devices are known to the hub yet.</Alert>
                                : <Box sx={{border: 1, borderColor: 'divider', borderRadius: 1, p: 1, maxHeight: 320, overflowY: 'auto'}}
                                       role="group" aria-label="Devices">
                                    {devices.map(d => <FormControlLabel key={d.id} sx={{display: 'flex', mr: 0}}
                                                                        control={<Checkbox size="small" checked={form.devices.includes(d.id)}
                                                                                           onChange={(_, on) => set({devices: on ? [...form.devices, d.id] : form.devices.filter(x => x !== d.id)})}/>}
                                                                        label={<Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                                                                            <Typography variant="body2" sx={{fontWeight: 600}}>{d.alias || d.name}</Typography>
                                                                            <Typography variant="caption" color="text.secondary">{d.os} {d.osVersion}</Typography>
                                                                            <Typography variant="caption" component="span" color="text.secondary">
                                                                                <DeviceStatus state={d.status} platform={d.platformType}/>
                                                                            </Typography>
                                                                        </Stack>}/>)}
                                </Box>}
                        {missingDevices.length > 0 && <Alert severity="info" sx={{mt: 1}}
                                                             action={<Button color="inherit" size="small"
                                                                             onClick={() => set({devices: form.devices.filter(id => !missingDevices.includes(id))})}>Remove</Button>}>
                            {missingDevices.length === 1 ? 'One picked device isn\'t' : `${missingDevices.length} picked devices aren't`} in the list (another platform, or deleted).
                        </Alert>}
                    </Box>}

                    <Typography variant="body2" sx={{fontWeight: 600, mt: 1}}>With several devices</Typography>
                    <RadioGroup value={String(form.executionType)} onChange={e => set({executionType: Number(e.target.value) as TestExecutionType})}>
                        <Choice value={String(TestExecutionType.Concurrent)} label="Spread the tests"
                                hint="Each test runs on the next free device. Faster results."/>
                        <Choice value={String(TestExecutionType.Simultaneously)} label="Every test on every device"
                                hint="Each test runs on all devices. Shows differences between devices."/>
                    </RadioGroup>
                </Section>
            </Stack>

            <Dialog open={leaving} onClose={() => setLeaving(false)} aria-labelledby="discard-test-title">
                <DialogTitle id="discard-test-title">Discard your changes?</DialogTitle>
                <DialogContent>
                    <DialogContentText>The changes to {test.name} aren't saved yet.</DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setLeaving(false)}>Keep editing</Button>
                    <Button color="warning" onClick={() => navigate(testUrl)}>Discard</Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default EditTestPage;
