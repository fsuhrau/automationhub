import React, {useEffect, useRef, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    InputLabel,
    LinearProgress,
    ListItemText,
    MenuItem,
    Select,
    Stack,
    TextField,
    Typography,
} from '@mui/material';
import {CloudUploadRounded, PlayArrowRounded} from '@mui/icons-material';
import {useNavigate} from 'react-router-dom';
import {AppParameterOption, IAppBinaryData, IAppData} from '../types/app';
import {PlatformType} from '../types/platform.type.enum';
import {getAppBundles, uploadBundle} from '../services/app.service';
import {executeTest} from '../services/test.service';
import {formatDateTime} from './DateTime';
import {acceptFor, accepts} from '../pages/apps/bundlesModel';

interface RunTestDialogProps {
    projectIdentifier: string,
    app: IAppData,
    test: { id: number, name: string } | null,
    onClose: () => void,
}

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};

// RunTestDialog starts a test: it picks the bundle to install (the newest by
// default, or a new upload), takes the app's parameters and, for web apps,
// the start URL, then opens the new run.
const RunTestDialog: React.FC<RunTestDialogProps> = ({projectIdentifier, app, test, onClose}) => {
    const navigate = useNavigate();
    const needsBundle = app.platform !== PlatformType.Editor && app.platform !== PlatformType.Web;
    const needsUrl = app.platform === PlatformType.Web;
    const accept = acceptFor(app.platform);

    const [bundles, setBundles] = useState<IAppBinaryData[] | null>(null);
    const [bundleId, setBundleId] = useState<number | ''>('');
    const [params, setParams] = useState<Record<string, string>>({});
    const [startUrl, setStartUrl] = useState('');
    const [uploading, setUploading] = useState<number | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [starting, setStarting] = useState(false);
    const fileInput = useRef<HTMLInputElement>(null);

    // a fresh form each time the dialog opens for a test
    useEffect(() => {
        if (!test) return;
        setError(null);
        setStarting(false);
        setParams(Object.fromEntries(app.parameter.map(p => [p.name, p.type.defaultValue ?? ''])));
        if (!needsBundle) return;
        setBundles(null);
        getAppBundles(projectIdentifier, app.id).then(list => {
            const sorted = [...list].sort((a, b) => b.id - a.id);
            setBundles(sorted);
            setBundleId(sorted.length > 0 ? sorted[0].id : '');
        }).catch(e => setError(messageOf(e)));
    }, [test?.id, app.id, projectIdentifier]);

    const onFile = (file: File | undefined) => {
        if (!file) return;
        if (!accepts(accept, file.name)) {
            setError(`${app.name} takes ${accept.label} files.`);
            return;
        }
        setError(null);
        setUploading(0);
        uploadBundle(file, projectIdentifier, app.id, sent => setUploading(sent)).then(bundle => {
            setBundles(list => [bundle, ...(list ?? []).filter(b => b.id !== bundle.id)]);
            setBundleId(bundle.id);
        }).catch(e => {
            const status = (e as { response?: { status?: number } }).response?.status;
            setError(status === 409 ? 'This exact file was uploaded before; pick it in the list.' : messageOf(e));
        }).finally(() => setUploading(null));
    };

    const start = () => {
        if (!test) return;
        setStarting(true);
        setError(null);
        executeTest(projectIdentifier, app.id, test.id, {
            appBinaryId: bundleId === '' ? null : bundleId,
            params: Object.entries(params).filter(([, v]) => v.length > 0).map(([k, v]) => `${k}=${v}`).join(';'),
            startUrl: needsUrl ? startUrl : '',
        }).then(run => {
            onClose();
            navigate(`/project/${projectIdentifier}/app:${app.id}/test/${test.id}/run/${run.id}`);
        }).catch(e => {
            setError(messageOf(e));
            setStarting(false);
        });
    };

    const ready = (!needsBundle || bundleId !== '') && (!needsUrl || startUrl.trim() !== '') && uploading === null && !starting;

    return (
        <Dialog open={!!test} onClose={starting ? undefined : onClose} fullWidth maxWidth="sm" aria-labelledby="run-test-title">
            <Box component="form" onSubmit={e => {
                e.preventDefault();
                if (ready) start();
            }}>
                <DialogTitle id="run-test-title">Run {test?.name}</DialogTitle>
                <DialogContent>
                    <Stack spacing={2.5} sx={{pt: 1}}>
                        {needsBundle && <Box>
                            <Stack direction="row" spacing={1} alignItems="flex-start">
                                <FormControl size="small" fullWidth disabled={bundles === null || bundles.length === 0}>
                                    <InputLabel id="run-bundle">Bundle</InputLabel>
                                    <Select labelId="run-bundle" label="Bundle" value={bundleId}
                                            onChange={e => setBundleId(e.target.value === '' ? '' : Number(e.target.value))}
                                            renderValue={id => {
                                                const b = bundles?.find(x => x.id === id);
                                                return b ? `${b.version || b.name} (#${b.id})` : '';
                                            }}>
                                        {(bundles ?? []).map((b, i) => <MenuItem key={b.id} value={b.id}>
                                            <ListItemText primary={`${b.version || b.name}${i === 0 ? ' (newest)' : ''}`}
                                                          secondary={`#${b.id} · uploaded ${formatDateTime(b.createdAt, 'YYYY/MM/DD HH:mm')}${b.tags ? ` · ${b.tags}` : ''}`}/>
                                        </MenuItem>)}
                                    </Select>
                                </FormControl>
                                <Button variant="outlined" startIcon={<CloudUploadRounded/>} sx={{whiteSpace: 'nowrap', flexShrink: 0}}
                                        disabled={uploading !== null} onClick={() => fileInput.current?.click()}>
                                    Upload
                                </Button>
                                <input ref={fileInput} type="file" hidden accept={accept.extensions.join(',')}
                                       onChange={e => {
                                           onFile(e.target.files?.[0]);
                                           e.target.value = '';
                                       }}/>
                            </Stack>
                            <Typography variant="caption" color="text.secondary">
                                {bundles === null ? 'Loading bundles…'
                                    : bundles.length === 0 ? `No bundles yet; upload ${accept.label} to run the test.`
                                        : 'The test installs this build on its devices first.'}
                            </Typography>
                            {uploading !== null && <Box sx={{mt: 1}}>
                                <Typography variant="caption">{uploading < 1 ? `Uploading ${Math.round(uploading * 100)} %` : 'Reading the bundle on the master…'}</Typography>
                                <LinearProgress variant={uploading < 1 ? 'determinate' : 'indeterminate'} value={uploading * 100}/>
                            </Box>}
                        </Box>}

                        {needsUrl && <TextField size="small" required label="Start URL" placeholder="https://" value={startUrl}
                                                onChange={e => setStartUrl(e.target.value)} helperText="The page the browser opens for the test."/>}

                        {app.parameter.length > 0 && <Box>
                            <Typography variant="subtitle2" sx={{mb: 1}}>Parameters</Typography>
                            <Stack spacing={1.5}>
                                {app.parameter.map(p => p.type.type === 'option'
                                    ? <FormControl key={p.id} size="small" fullWidth>
                                        <InputLabel id={`run-param-${p.id}`}>{p.name}</InputLabel>
                                        <Select labelId={`run-param-${p.id}`} label={p.name} value={params[p.name] ?? ''}
                                                onChange={e => setParams(v => ({...v, [p.name]: e.target.value}))}>
                                            {(p.type as AppParameterOption).options.map(o => <MenuItem key={o} value={o}>
                                                {o}{o === p.type.defaultValue ? ' (default)' : ''}
                                            </MenuItem>)}
                                        </Select>
                                    </FormControl>
                                    : <TextField key={p.id} size="small" fullWidth label={p.name} value={params[p.name] ?? ''}
                                                 placeholder={p.type.defaultValue || undefined}
                                                 onChange={e => setParams(v => ({...v, [p.name]: e.target.value}))}/>)}
                            </Stack>
                            <Typography variant="caption" color="text.secondary">Empty parameters aren't passed.</Typography>
                        </Box>}

                        {error && <Alert severity="error">{error}</Alert>}
                    </Stack>
                </DialogContent>
                <DialogActions>
                    <Button onClick={onClose} disabled={starting}>Cancel</Button>
                    <Button type="submit" variant="contained" startIcon={<PlayArrowRounded/>} disabled={!ready}>
                        {starting ? 'Starting…' : 'Start run'}
                    </Button>
                </DialogActions>
            </Box>
        </Dialog>
    );
};

export default RunTestDialog;
