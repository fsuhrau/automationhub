import React, {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    IconButton,
    InputAdornment,
    LinearProgress,
    ListItemIcon,
    ListItemText,
    Menu,
    MenuItem,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    CloudUploadRounded,
    ContentCopyRounded,
    DeleteOutlineRounded,
    DownloadRounded,
    EditRounded,
    MoreVertRounded,
    SearchRounded,
} from '@mui/icons-material';
import {useDropzone} from 'react-dropzone';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {useError} from '../../ErrorProvider';
import {IAppBinaryData} from '../../types/app';
import {getPlatformName, PlatformType} from '../../types/platform.type.enum';
import {deleteAppBundle, getAppBundles, updateBundleTags, uploadBundle} from '../../services/app.service';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import {formatDateTime} from '../../components/DateTime';
import {byteFormat} from '../tests/value_formatter';
import {acceptFor, accepts, ago, joinTags, matches, parseTags} from './bundlesModel';

type UploadState =
    | { phase: 'sending', sent: number }
    | { phase: 'reading' }
    | { phase: 'done', bundle: IAppBinaryData }
    | { phase: 'exists' }
    | { phase: 'failed', message: string };

interface Upload {
    id: number,
    file: File,
    state: UploadState,
}

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};

const statusOf = (e: unknown): number | undefined => (e as { response?: { status?: number } }).response?.status;

let nextUpload = 0;

const UploadRow: React.FC<{ upload: Upload, onDismiss: () => void }> = ({upload, onDismiss}) => {
    const {file, state} = upload;
    const severity = state.phase === 'failed' ? 'error' : state.phase === 'exists' ? 'warning' : state.phase === 'done' ? 'success' : 'info';
    const text = state.phase === 'sending' ? `Uploading ${Math.round(state.sent * 100)} %`
        : state.phase === 'reading' ? 'Reading the bundle on the master…'
            : state.phase === 'done' ? `Uploaded as version ${state.bundle.version || state.bundle.name} (#${state.bundle.id})`
                : state.phase === 'exists' ? 'This exact file was uploaded before; the list already has it.'
                    : state.message;
    const busy = state.phase === 'sending' || state.phase === 'reading';
    return <Alert severity={severity} onClose={busy ? undefined : onDismiss} sx={{'& .MuiAlert-message': {flex: 1}}}>
        <Typography variant="body2" sx={{fontWeight: 600, wordBreak: 'break-all'}}>
            {file.name} <Typography component="span" variant="caption" color="text.secondary">{byteFormat(file.size)}</Typography>
        </Typography>
        <Typography variant="body2">{text}</Typography>
        {busy && <LinearProgress sx={{mt: 1, height: 6}} variant={state.phase === 'sending' ? 'determinate' : 'indeterminate'}
                                 value={state.phase === 'sending' ? state.sent * 100 : undefined}/>}
    </Alert>;
};

// AppBundlesPage lists the builds of the app picked in the menu and takes
// new ones, by button or by dropping files anywhere on the page.
const AppBundlesPage: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const {setError} = useError();
    const app = state.apps?.find(a => a.id === appId);
    const accept = acceptFor(app?.platform);

    const [bundles, setBundles] = useState<IAppBinaryData[] | null>(null);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [search, setSearch] = useState('');
    const [uploads, setUploads] = useState<Upload[]>([]);
    const [fresh, setFresh] = useState<number | null>(null);
    const [menu, setMenu] = useState<{ anchor: HTMLElement, bundle: IAppBinaryData } | null>(null);
    const [deleting, setDeleting] = useState<IAppBinaryData | null>(null);
    const [tagging, setTagging] = useState<{ bundle: IAppBinaryData, text: string } | null>(null);
    const [busy, setBusy] = useState(false);
    const now = Date.now();

    const load = useCallback(() => {
        if (appId === null) return;
        setLoadError(null);
        getAppBundles(projectIdentifier, appId).then(list => setBundles([...list].sort((a, b) => b.id - a.id)))
            .catch(e => setLoadError(messageOf(e)));
    }, [projectIdentifier, appId]);
    useEffect(load, [load]);

    const setUpload = (id: number, s: UploadState) => setUploads(list => list.map(u => u.id === id ? {...u, state: s} : u));

    // files are uploaded one after the other; the master reads each one
    const queue = useRef<Promise<void>>(Promise.resolve());
    const upload = (files: File[]) => {
        if (appId === null) return;
        files.forEach(file => {
            const id = nextUpload++;
            if (!accepts(accept, file.name)) {
                setUploads(list => [...list, {id, file, state: {phase: 'failed', message: `${app?.name ?? 'This app'} takes ${accept.label} files.`}}]);
                return;
            }
            setUploads(list => [...list, {id, file, state: {phase: 'sending', sent: 0}}]);
            queue.current = queue.current.then(() => uploadBundle(file, projectIdentifier, appId, sent => {
                setUpload(id, sent < 1 ? {phase: 'sending', sent} : {phase: 'reading'});
            }).then(bundle => {
                setUpload(id, {phase: 'done', bundle});
                setFresh(bundle.id);
                load();
            }).catch(e => {
                setUpload(id, statusOf(e) === 409 ? {phase: 'exists'} : {phase: 'failed', message: messageOf(e)});
            }));
        });
    };

    const {getRootProps, getInputProps, isDragActive, open} = useDropzone({
        onDrop: upload, noClick: true, noKeyboard: true, multiple: true,
    });

    const visible = useMemo(() => (bundles ?? []).filter(b => matches(b, search)), [bundles, search]);

    const onDelete = () => {
        if (!deleting || appId === null) return;
        setBusy(true);
        deleteAppBundle(projectIdentifier, appId, deleting.id).then(() => {
            setBundles(list => list ? list.filter(b => b.id !== deleting.id) : list);
            setDeleting(null);
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const onSaveTags = () => {
        if (!tagging || appId === null) return;
        setBusy(true);
        updateBundleTags(projectIdentifier, appId, tagging.bundle.id, joinTags(tagging.text.split(','))).then(saved => {
            setBundles(list => list ? list.map(b => b.id === saved.id ? {...b, tags: saved.tags} : b) : list);
            setTagging(null);
        }).catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const busyUploads = uploads.some(u => u.state.phase === 'sending' || u.state.phase === 'reading');

    return (
        <Box {...getRootProps()} sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2, position: 'relative', outline: 'none'}}>
            <input {...getInputProps()} aria-label="Bundle files"/>

            {isDragActive && <Box sx={{
                position: 'absolute', inset: -8, zIndex: 2, border: '2px dashed', borderColor: 'primary.main', borderRadius: 2,
                bgcolor: 'background.default', opacity: 0.94, display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}>
                <Stack alignItems="center" spacing={1}>
                    <CloudUploadRounded color="primary" sx={{fontSize: 48}}/>
                    <Typography variant="h6">Drop to upload to {app?.name ?? 'this app'}</Typography>
                    <Typography variant="body2" color="text.secondary">{accept.label}</Typography>
                </Stack>
            </Box>}

            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Stack direction="row" spacing={1.5} alignItems="center" sx={{minWidth: 0}}>
                    {app && <Box sx={{display: 'flex', p: 1, borderRadius: 1, border: 1, borderColor: 'divider'}}>
                        <PlatformTypeIcon platformType={app.platform}/>
                    </Box>}
                    <Box sx={{minWidth: 0}}>
                        <Typography component="h1" variant="h5">Bundles</Typography>
                        <Typography variant="body2" color="text.secondary" sx={{wordBreak: 'break-word'}}>
                            Builds of {app?.name ?? 'this app'}
                            {app?.identifier && <> (<Box component="span" sx={{fontFamily: 'monospace'}}>{app.identifier}</Box>)</>}
                            {' '}to install on devices and run tests with.
                        </Typography>
                    </Box>
                </Stack>
                <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {md: 'auto'}}}>
                    <Typography variant="caption" color="text.secondary" sx={{display: {xs: 'none', lg: 'block'}}}>
                        or drop {accept.label} files on the page
                    </Typography>
                    <Button variant="contained" startIcon={<CloudUploadRounded/>} onClick={open}>Upload bundle</Button>
                </Stack>
            </Stack>

            {app && (app.platform === PlatformType.Editor || app.platform === PlatformType.Web) &&
                <Alert severity="info" sx={{mb: 2}}>
                    {getPlatformName(app.platform)} apps don't need a bundle: tests run in the {app.platform === PlatformType.Editor ? 'Unity editor' : 'browser'}.
                </Alert>}

            {uploads.length > 0 && <Stack spacing={1} sx={{mb: 2}} aria-live="polite">
                {uploads.map(u => <UploadRow key={u.id} upload={u} onDismiss={() => setUploads(list => list.filter(x => x.id !== u.id))}/>)}
                {!busyUploads && uploads.length > 1 && <Box>
                    <Button size="small" onClick={() => setUploads([])}>Clear all</Button>
                </Box>}
            </Stack>}

            <Card variant="outlined">
                <CardContent>
                    {loadError
                        ? <Alert severity="error" action={<Button color="inherit" size="small" onClick={load}>Retry</Button>}>{loadError}</Alert>
                        : bundles === null
                            ? <LinearProgress/>
                            : bundles.length === 0
                                ? <Stack alignItems="center" spacing={1.5} sx={{py: 6, textAlign: 'center'}}>
                                    <CloudUploadRounded color="disabled" sx={{fontSize: 48}}/>
                                    <Typography variant="subtitle1">No bundles yet</Typography>
                                    <Typography variant="body2" color="text.secondary">
                                        Drop {accept.label} files here, or choose them. Test runs and live control install a bundle on the device.
                                    </Typography>
                                    <Button variant="outlined" onClick={open}>Choose files</Button>
                                </Stack>
                                : <>
                                    <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 1.5}}>
                                        <TextField size="small" placeholder="Find by version, tag or #id" value={search}
                                                   onChange={e => setSearch(e.target.value)} sx={{width: {xs: '100%', sm: 300}}}
                                                   slotProps={{
                                                       input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                                                       htmlInput: {'aria-label': 'Find a bundle'},
                                                   }}/>
                                        <Typography variant="body2" color="text.secondary" sx={{ml: 'auto'}}>
                                            {visible.length === bundles.length ? `${bundles.length} ${bundles.length === 1 ? 'bundle' : 'bundles'}` : `${visible.length} of ${bundles.length}`}
                                        </Typography>
                                    </Stack>
                                    {visible.length === 0
                                        ? <Typography variant="body2" color="text.secondary" sx={{py: 2}}>No bundle matches.</Typography>
                                        : <TableContainer sx={{mx: -2, width: 'auto'}}>
                                            <Table size="small" aria-label="Bundles">
                                                <TableHead>
                                                    <TableRow>
                                                        <TableCell sx={{pl: 2}}>Version</TableCell>
                                                        <TableCell>Tags</TableCell>
                                                        <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>Uploaded</TableCell>
                                                        <TableCell align="right" sx={{display: {xs: 'none', sm: 'table-cell'}}}>Size</TableCell>
                                                        <TableCell align="right" sx={{display: {xs: 'none', md: 'table-cell'}}}>Runs</TableCell>
                                                        <TableCell/>
                                                    </TableRow>
                                                </TableHead>
                                                <TableBody>
                                                    {visible.map(b => {
                                                        const tags = parseTags(b.tags);
                                                        return <TableRow key={b.id} selected={b.id === fresh}>
                                                            <TableCell sx={{pl: 2, maxWidth: 420}}>
                                                                <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                                                                    <Typography variant="body2" sx={{fontWeight: 600, wordBreak: 'break-all'}}>{b.version || '–'}</Typography>
                                                                    {b.id === bundles[0].id && <Chip size="small" variant="outlined" label="newest"/>}
                                                                    {b.id === fresh && <Chip size="small" color="success" label="just uploaded"/>}
                                                                </Stack>
                                                                <Typography variant="caption" color="text.secondary">{b.name} · #{b.id}</Typography>
                                                            </TableCell>
                                                            <TableCell>
                                                                <Stack direction="row" spacing={0.5} alignItems="center" flexWrap="wrap" useFlexGap>
                                                                    {tags.map(t => <Chip key={t} size="small" label={t} onClick={() => setSearch(t)}/>)}
                                                                    {tags.length > 0
                                                                        ? <Tooltip title="Edit tags">
                                                                            <IconButton size="small" aria-label={`Edit tags of #${b.id}`}
                                                                                        onClick={() => setTagging({bundle: b, text: tags.join(', ')})}>
                                                                                <EditRounded fontSize="inherit"/>
                                                                            </IconButton>
                                                                        </Tooltip>
                                                                        : <Button size="small" variant="text" aria-label={`Add tags to #${b.id}`}
                                                                                  onClick={() => setTagging({bundle: b, text: ''})}>Add tags</Button>}
                                                                </Stack>
                                                            </TableCell>
                                                            <TableCell sx={{display: {xs: 'none', md: 'table-cell'}, whiteSpace: 'nowrap'}}>
                                                                <Tooltip title={formatDateTime(b.createdAt)}><span>{ago(b.createdAt, now)}</span></Tooltip>
                                                            </TableCell>
                                                            <TableCell align="right" sx={{display: {xs: 'none', sm: 'table-cell'}, whiteSpace: 'nowrap'}}>
                                                                {b.size ? byteFormat(b.size) : '–'}
                                                            </TableCell>
                                                            <TableCell align="right" sx={{display: {xs: 'none', md: 'table-cell'}}}>
                                                                {b.runCount === undefined ? '–' : b.runCount}
                                                            </TableCell>
                                                            <TableCell align="right" sx={{whiteSpace: 'nowrap', pr: 1}}>
                                                                <Tooltip title="Download">
                                                                    <span><IconButton size="small" aria-label={`Download #${b.id}`} disabled={!b.appPath}
                                                                                      href={`/upload/${encodeURIComponent(b.appPath)}`} download>
                                                                        <DownloadRounded fontSize="small"/>
                                                                    </IconButton></span>
                                                                </Tooltip>
                                                                <IconButton size="small" aria-label={`More for #${b.id}`}
                                                                            onClick={e => setMenu({anchor: e.currentTarget, bundle: b})}>
                                                                    <MoreVertRounded fontSize="small"/>
                                                                </IconButton>
                                                            </TableCell>
                                                        </TableRow>;
                                                    })}
                                                </TableBody>
                                            </Table>
                                        </TableContainer>}
                                </>}
                </CardContent>
            </Card>

            <Menu anchorEl={menu?.anchor} open={!!menu} onClose={() => setMenu(null)}>
                <MenuItem disabled={!menu?.bundle.hash} onClick={() => {
                    navigator.clipboard.writeText(menu!.bundle.hash).catch(() => undefined);
                    setMenu(null);
                }}>
                    <ListItemIcon><ContentCopyRounded fontSize="small"/></ListItemIcon>
                    <ListItemText primary="Copy hash" secondary={menu?.bundle.hash ? `${menu.bundle.hash.slice(0, 12)}…` : undefined}/>
                </MenuItem>
                <MenuItem onClick={() => {
                    setDeleting(menu!.bundle);
                    setMenu(null);
                }} sx={{color: 'error.main'}}>
                    <ListItemIcon><DeleteOutlineRounded fontSize="small" color="error"/></ListItemIcon>
                    <ListItemText primary="Delete bundle"/>
                </MenuItem>
            </Menu>

            <Dialog open={!!deleting} onClose={() => setDeleting(null)} aria-labelledby="delete-bundle-title">
                <DialogTitle id="delete-bundle-title">Delete version {deleting?.version || deleting?.name}?</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        The file{deleting?.size ? ` (${byteFormat(deleting.size)})` : ''} is removed from the master.
                        {deleting?.runCount
                            ? ` ${deleting.runCount} test ${deleting.runCount === 1 ? 'run used' : 'runs used'} it; their results stay, but they can't be run again with this bundle.`
                            // unknown on older masters: say nothing rather than "none"
                            : deleting?.runCount === 0 ? ' No test run used it.' : ''}
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setDeleting(null)}>Keep it</Button>
                    <Button color="error" variant="contained" disabled={busy} onClick={onDelete}>Delete bundle</Button>
                </DialogActions>
            </Dialog>

            <Dialog open={!!tagging} onClose={() => setTagging(null)} fullWidth maxWidth="xs" aria-labelledby="tags-title">
                <Box component="form" onSubmit={e => {
                    e.preventDefault();
                    onSaveTags();
                }}>
                    <DialogTitle id="tags-title">Tags of version {tagging?.bundle.version || tagging?.bundle.name}</DialogTitle>
                    <DialogContent>
                        <DialogContentText variant="body2" sx={{mb: 2}}>
                            Labels to find a bundle by, like a branch or "release candidate". Separate them with commas.
                        </DialogContentText>
                        <TextField autoFocus fullWidth size="small" label="Tags" placeholder="nightly, rc1" value={tagging?.text ?? ''}
                                   onChange={e => setTagging(t => t ? {...t, text: e.target.value} : t)}/>
                        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap sx={{mt: 1}}>
                            {parseTags(tagging?.text).map(t => <Chip key={t} size="small" label={t}/>)}
                        </Stack>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={() => setTagging(null)}>Cancel</Button>
                        <Button type="submit" variant="contained" disabled={busy}>Save</Button>
                    </DialogActions>
                </Box>
            </Dialog>
        </Box>
    );
};

export default AppBundlesPage;
