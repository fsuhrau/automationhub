import React, {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {useNavigate, useParams} from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    ButtonGroup,
    Checkbox,
    Chip,
    CircularProgress,
    FormControl,
    FormControlLabel,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Stack,
    Switch,
    Tab,
    Tabs,
    TextField,
    Typography,
} from '@mui/material';
import {RichTreeView} from '@mui/x-tree-view/RichTreeView';
import {TreeViewBaseItem} from '@mui/x-tree-view/models';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useError} from '../../ErrorProvider';
import IDeviceData from '../../types/device';
import {IAppBinaryData, IAppData} from '../../types/app';
import {getDevice} from '../../services/device.service';
import {getAllApps, getAppBundles} from '../../services/app.service';
import {
    appAction,
    AppAction,
    DeviceLockStatus,
    getDeviceLogs,
    getHierarchy,
    getLockStatus,
    inputText,
    lockDevice,
    pressKey,
    releaseDevice,
    screenshotUrl,
    streamUrl,
    swipe,
    tap,
    touchElement,
} from '../../services/device.control.service';
import {TitleCard} from '../../components/title.card.component';
import {describeController} from "../../components/DeviceController";

const TAP_DISTANCE = 12; // css px; a longer drag is a swipe
const KEYS = ['back', 'home', 'app_switch', 'enter', 'delete', 'volume_up', 'volume_down'];

const leaseStorageKey = (deviceId: number) => `device-lease-${deviceId}`;
const loadLease = (deviceId: number): string | null => {
    try {
        return sessionStorage.getItem(leaseStorageKey(deviceId));
    } catch {
        return null;
    }
};
const storeLease = (deviceId: number, token: string | null) => {
    try {
        if (token) {
            sessionStorage.setItem(leaseStorageKey(deviceId), token);
        } else {
            sessionStorage.removeItem(leaseStorageKey(deviceId));
        }
    } catch {
        // storage unavailable; the lease only lives in memory
    }
};

// --- scene graph ---

interface SceneNode extends TreeViewBaseItem {
    element: Element,
}

interface SceneGraph {
    items: SceneNode[],
    byId: Map<string, Element>,
    width: number,
    height: number,
}

const nodeLabel = (e: Element) => {
    const name = e.getAttribute('Name');
    const text = e.getAttribute('LabelText');
    return `${e.nodeName}${name ? ` "${name}"` : ''}${text ? ` – ${text}` : ''}`;
};

const parseSceneGraph = (xml: string): SceneGraph => {
    const doc = new DOMParser().parseFromString(xml, 'text/xml');
    const root = doc.documentElement;
    if (!root || root.nodeName === 'parsererror') {
        throw new Error('scene graph is not valid XML');
    }
    const byId = new Map<string, Element>();
    let counter = 0;
    const build = (e: Element): SceneNode => {
        const id = `n${counter++}`;
        byId.set(id, e);
        return {id, label: nodeLabel(e), element: e, children: Array.from(e.children).map(build)};
    };
    return {
        items: [build(root)],
        byId,
        width: Number(root.getAttribute('RectangleX')) || 0,
        height: Number(root.getAttribute('RectangleY')) || 0,
    };
};

const xpathFor = (e: Element): string => {
    const name = e.getAttribute('Name');
    if (name) {
        return `//${e.nodeName}[@Name='${name.replace(/'/g, "\\'")}']`;
    }
    const parts: string[] = [];
    let current: Element | null = e;
    while (current && current.parentElement) {
        const tag: string = current.nodeName;
        const index = Array.from(current.parentElement.children).filter(c => c.nodeName === tag).indexOf(current) + 1;
        parts.unshift(`${tag}[${index}]`);
        current = current.parentElement;
    }
    return '/' + (current ? current.nodeName : '') + (parts.length ? '/' + parts.join('/') : '');
};

const errorMessage = (error: unknown): string => {
    const e = error as { response?: { data?: { message?: string } }, message?: string };
    return e.response?.data?.message ?? e.message ?? String(error);
};

const DeviceControlPage: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {deviceId: deviceIdParam} = useParams();
    const deviceId = Number(deviceIdParam);
    const navigate = useNavigate();
    const {setError} = useError();

    const [device, setDevice] = useState<IDeviceData | null>(null);
    const [lockStatus, setLockStatus] = useState<DeviceLockStatus | null>(null);
    const [lease, setLease] = useState<string | null>(() => loadLease(deviceId));
    const [busy, setBusy] = useState(false);
    const [live, setLive] = useState(true);
    const [snapshot, setSnapshot] = useState(() => screenshotUrl(projectIdentifier, deviceId));
    const [tab, setTab] = useState(0);
    const [text, setText] = useState('');

    const [scene, setScene] = useState<SceneGraph | null>(null);
    const [selected, setSelected] = useState<Element | null>(null);
    const [logs, setLogs] = useState<string[]>([]);

    const [apps, setApps] = useState<IAppData[]>([]);
    const [appId, setAppId] = useState<number | ''>('');
    const [bundles, setBundles] = useState<IAppBinaryData[]>([]);
    const [bundleId, setBundleId] = useState<number | ''>('');
    const [clearData, setClearData] = useState(false);

    const imgRef = useRef<HTMLImageElement>(null);
    const pointerStart = useRef<{ x: number, y: number, t: number } | null>(null);

    const fail = useCallback((error: unknown) => setError(errorMessage(error)), [setError]);

    const refreshLock = useCallback(() => {
        getLockStatus(projectIdentifier, deviceId).then(status => {
            setLockStatus(status);
            if (!status.lease && lease) {
                // our lease expired or was released elsewhere
                setLease(null);
                storeLease(deviceId, null);
            }
        }).catch(() => undefined);
    }, [projectIdentifier, deviceId, lease]);

    useEffect(() => {
        getDevice(projectIdentifier, String(deviceId)).then(d => setDevice(d ?? null)).catch(fail);
        getAllApps(projectIdentifier).then(setApps).catch(fail);
    }, [projectIdentifier, deviceId, fail]);

    useEffect(() => {
        refreshLock();
        const timer = setInterval(refreshLock, 5000);
        return () => clearInterval(timer);
    }, [refreshLock]);

    useEffect(() => {
        if (appId === '') {
            setBundles([]);
            return;
        }
        getAppBundles(projectIdentifier, appId).then(list => setBundles([...list].sort((a, b) => b.id - a.id))).catch(fail);
        setBundleId('');
    }, [projectIdentifier, appId, fail]);

    const run = async (action: () => Promise<unknown>) => {
        setBusy(true);
        try {
            await action();
        } catch (error) {
            fail(error);
            refreshLock();
        } finally {
            setBusy(false);
        }
    };

    const onLock = () => run(async () => {
        const l = await lockDevice(projectIdentifier, deviceId);
        setLease(l.token ?? null);
        storeLease(deviceId, l.token ?? null);
        refreshLock();
    });

    const onRelease = () => run(async () => {
        if (lease) {
            await releaseDevice(projectIdentifier, deviceId, lease);
        }
        setLease(null);
        storeLease(deviceId, null);
        refreshLock();
    });

    // map a pointer position on the image to device pixels
    const toDevice = (clientX: number, clientY: number) => {
        const img = imgRef.current!;
        const rect = img.getBoundingClientRect();
        const sx = (img.naturalWidth || rect.width) / rect.width;
        const sy = (img.naturalHeight || rect.height) / rect.height;
        return {x: Math.round((clientX - rect.left) * sx), y: Math.round((clientY - rect.top) * sy)};
    };

    const onPointerDown = (e: React.PointerEvent<HTMLImageElement>) => {
        e.preventDefault();
        pointerStart.current = {x: e.clientX, y: e.clientY, t: Date.now()};
    };

    const onPointerUp = (e: React.PointerEvent<HTMLImageElement>) => {
        const start = pointerStart.current;
        pointerStart.current = null;
        if (!start) return;
        if (!lease) {
            setError('Lock the device to control it');
            return;
        }
        const from = toDevice(start.x, start.y);
        const to = toDevice(e.clientX, e.clientY);
        const distance = Math.hypot(e.clientX - start.x, e.clientY - start.y);
        if (distance < TAP_DISTANCE) {
            run(() => tap(projectIdentifier, deviceId, lease, from.x, from.y));
        } else {
            const duration = Math.min(Math.max(Date.now() - start.t, 100), 2000);
            run(() => swipe(projectIdentifier, deviceId, lease, from.x, from.y, to.x, to.y, duration));
        }
        if (!live) {
            setTimeout(() => setSnapshot(screenshotUrl(projectIdentifier, deviceId)), 700);
        }
    };

    const loadHierarchy = () => run(async () => {
        const xml = await getHierarchy(projectIdentifier, deviceId);
        setScene(parseSceneGraph(xml));
        setSelected(null);
    });

    const loadLogs = () => run(async () => setLogs(await getDeviceLogs(projectIdentifier, deviceId, 300)));

    const onAppAction = (action: AppAction) => {
        if (!lease || bundleId === '') return;
        run(() => appAction(projectIdentifier, deviceId, lease, action, bundleId, action === 'stop' && clearData));
    };

    // highlight of the selected scene graph element, in percent of the image
    const highlight = useMemo(() => {
        if (!selected || !scene || !scene.width || !scene.height) return null;
        const n = (attr: string) => Number(selected.getAttribute(attr)) || 0;
        return {
            left: `${(n('X') / scene.width) * 100}%`,
            top: `${(n('Y') / scene.height) * 100}%`,
            width: `${(n('RectangleX') / scene.width) * 100}%`,
            height: `${(n('RectangleY') / scene.height) * 100}%`,
        };
    }, [selected, scene]);

    const holder = lockStatus?.lease?.owner;
    const ownLease = lease !== null && lockStatus?.lease !== undefined;
    const imageSrc = live ? streamUrl(projectIdentifier, deviceId, 2) : snapshot;

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={`Control: ${device ? (device.alias || device.name) : deviceId}`}>
                <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap" useFlexGap sx={{p: 1}}>
                    {ownLease
                        ? <Chip color="success" label={`Controlled by you (${describeController(holder!).label})`}/>
                        : holder
                            ? <Chip color="warning" label={`Controlled by ${describeController(holder).label}`}/>
                            : lockStatus?.lockedByTest
                                ? <Chip color="warning" label="In use by a test run"/>
                                : <Chip label="Free"/>}
                    {ownLease
                        ? <Button variant="outlined" onClick={onRelease} disabled={busy}>Release</Button>
                        : <Button variant="contained" onClick={onLock} disabled={busy || lockStatus?.locked}>Lock for control</Button>}
                    <FormControlLabel control={<Switch checked={live} onChange={(_, v) => setLive(v)}/>} label="Live stream"/>
                    {!live && <Button onClick={() => setSnapshot(screenshotUrl(projectIdentifier, deviceId))}>Refresh</Button>}
                    {busy && <CircularProgress size={20}/>}
                    <Box sx={{flexGrow: 1}}/>
                    <Button onClick={() => navigate(`/project/${projectIdentifier}/device/${deviceId}`)}>Device details</Button>
                </Stack>
            </TitleCard>

            <Box sx={{display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'flex-start'}}>
                <Paper sx={{p: 1, flex: '0 1 auto', maxWidth: '100%'}}>
                    <Box sx={{position: 'relative', display: 'inline-block', lineHeight: 0}}>
                        <img
                            ref={imgRef}
                            src={imageSrc}
                            alt="device screen"
                            draggable={false}
                            onPointerDown={onPointerDown}
                            onPointerUp={onPointerUp}
                            style={{maxHeight: '75vh', maxWidth: '100%', cursor: lease ? 'crosshair' : 'not-allowed', userSelect: 'none', touchAction: 'none'}}
                        />
                        {highlight && <Box sx={{
                            position: 'absolute', pointerEvents: 'none', border: '2px solid red',
                            backgroundColor: 'rgba(255,0,0,0.15)', ...highlight,
                        }}/>}
                    </Box>
                    {!lease && <Alert severity="info" sx={{mt: 1}}>Lock the device to tap (click) and swipe (drag) on the screen.</Alert>}
                </Paper>

                <Paper sx={{p: 1, flex: '1 1 420px', minWidth: 320}}>
                    <Tabs value={tab} onChange={(_, v) => setTab(v)}>
                        <Tab label="Input"/>
                        <Tab label="App"/>
                        <Tab label="Hierarchy"/>
                        <Tab label="Logs"/>
                    </Tabs>

                    {tab === 0 && <Stack spacing={2} sx={{p: 1}}>
                        <Stack direction="row" spacing={1}>
                            <TextField size="small" fullWidth label="Text" value={text} onChange={e => setText(e.target.value)}/>
                            <Button variant="contained" disabled={!lease || !text}
                                    onClick={() => run(async () => {
                                        await inputText(projectIdentifier, deviceId, lease!, text);
                                        setText('');
                                    })}>Type</Button>
                        </Stack>
                        <ButtonGroup variant="outlined" size="small" sx={{flexWrap: 'wrap'}}>
                            {KEYS.map(key => <Button key={key} disabled={!lease}
                                                     onClick={() => run(() => pressKey(projectIdentifier, deviceId, lease!, key))}>{key}</Button>)}
                        </ButtonGroup>
                    </Stack>}

                    {tab === 1 && <Stack spacing={2} sx={{p: 1}}>
                        <FormControl size="small" fullWidth>
                            <InputLabel id="control-app">App</InputLabel>
                            <Select labelId="control-app" label="App" value={appId}
                                    onChange={e => setAppId(e.target.value === '' ? '' : Number(e.target.value))}>
                                {apps.map(a => <MenuItem key={a.id} value={a.id}>{a.name}</MenuItem>)}
                            </Select>
                        </FormControl>
                        <FormControl size="small" fullWidth disabled={appId === ''}>
                            <InputLabel id="control-bundle">Bundle</InputLabel>
                            <Select labelId="control-bundle" label="Bundle" value={bundleId}
                                    onChange={e => setBundleId(e.target.value === '' ? '' : Number(e.target.value))}>
                                {bundles.map(b => <MenuItem key={b.id} value={b.id}>{b.name} {b.version} (#{b.id})</MenuItem>)}
                            </Select>
                        </FormControl>
                        <Stack direction="row" spacing={1} alignItems="center">
                            <Button variant="contained" disabled={!lease || bundleId === ''} onClick={() => onAppAction('install')}>Install</Button>
                            <Button variant="contained" disabled={!lease || bundleId === ''} onClick={() => onAppAction('start')}>Start</Button>
                            <Button variant="outlined" color="warning" disabled={!lease || bundleId === ''} onClick={() => onAppAction('stop')}>Stop</Button>
                            <FormControlLabel control={<Checkbox checked={clearData} onChange={(_, v) => setClearData(v)}/>} label="clear data on stop"/>
                        </Stack>
                    </Stack>}

                    {tab === 2 && <Box sx={{p: 1}}>
                        <Stack direction="row" spacing={1} sx={{mb: 1}}>
                            <Button variant="outlined" onClick={loadHierarchy} disabled={busy}>Load scene graph</Button>
                            {selected && <Button variant="contained" disabled={!lease || !selected.getAttribute('ID')}
                                                 onClick={() => run(() => touchElement(projectIdentifier, deviceId, lease!, {elementId: selected.getAttribute('ID')!}))}>
                                Touch element</Button>}
                        </Stack>
                        {selected && <Box sx={{mb: 1}}>
                            <Stack direction="row" spacing={1} alignItems="center">
                                <Typography variant="body2" sx={{fontFamily: 'monospace', wordBreak: 'break-all'}}>{xpathFor(selected)}</Typography>
                                <Button size="small" onClick={() => navigator.clipboard.writeText(xpathFor(selected)).catch(fail)}>Copy</Button>
                            </Stack>
                            <Typography variant="caption" sx={{fontFamily: 'monospace'}}>
                                {Array.from(selected.attributes).map(a => `${a.name}=${a.value}`).join('  ')}
                            </Typography>
                        </Box>}
                        {scene
                            ? <Box sx={{maxHeight: '55vh', overflow: 'auto'}}>
                                <RichTreeView items={scene.items}
                                              onSelectedItemsChange={(_, id) => setSelected(id ? scene.byId.get(id as string) ?? null : null)}/>
                            </Box>
                            : <Typography variant="body2">Needs a connected app (start it from the App tab).</Typography>}
                    </Box>}

                    {tab === 3 && <Box sx={{p: 1}}>
                        <Button variant="outlined" onClick={loadLogs} disabled={busy} sx={{mb: 1}}>Load last 300 lines</Button>
                        <Box component="pre" sx={{
                            m: 0, p: 1, maxHeight: '55vh', overflow: 'auto', fontSize: 12,
                            bgcolor: 'background.default', border: 1, borderColor: 'divider',
                        }}>{logs.join('\n')}</Box>
                    </Box>}
                </Paper>
            </Box>
        </Box>
    );
};

export default DeviceControlPage;
