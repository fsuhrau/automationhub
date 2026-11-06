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
    Tab,
    Tabs,
    TextField,
    ToggleButton,
    ToggleButtonGroup,
    Tooltip,
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
import {DeviceVideoPlayer, VIDEO_TEXT_MAX_BYTES, VideoMeta, videoDecodingAvailable, VideoStatus, videoUrl} from '../../services/device.video';

const TAP_DISTANCE = 12; // css px; a longer drag is a swipe

// video: H.264 stream (Android), screenshots: MJPEG of screenshots, off: manual refresh
type ScreenMode = 'video' | 'screenshots' | 'off';
const VIDEO_OPTIONS = {maxSize: 1280, maxFps: 30};
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
    const canDecodeVideo = useMemo(videoDecodingAvailable, []);
    const [mode, setMode] = useState<ScreenMode>(canDecodeVideo ? 'video' : 'screenshots');
    const [videoStatus, setVideoStatus] = useState<VideoStatus | null>(null);
    const [videoMeta, setVideoMeta] = useState<VideoMeta | null>(null);
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

    const canvasRef = useRef<HTMLCanvasElement>(null);
    const pointerStart = useRef<{ x: number, y: number, t: number } | null>(null);
    // live input over the video socket (scrcpy control channel)
    const playerRef = useRef<DeviceVideoPlayer | null>(null);
    const activeTouch = useRef<{ pointerId: number, player: DeviceVideoPlayer } | null>(null);
    const pendingMove = useRef<{ x: number, y: number } | null>(null);
    const moveFrame = useRef(0);
    const leaseRef = useRef(lease);
    leaseRef.current = lease;
    const inputErrorRef = useRef<(message: string) => void>(() => undefined);
    const lastInputError = useRef(0);

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
        if (mode !== 'video' || !canvasRef.current) return;
        setVideoMeta(null);
        const player = new DeviceVideoPlayer(
            videoUrl(projectIdentifier, deviceId, VIDEO_OPTIONS),
            canvasRef.current,
            status => {
                setVideoStatus(status);
                if (status.state === 'unsupported') {
                    // e.g. iOS: fall back to the screenshot stream
                    setMode('screenshots');
                }
            },
            setVideoMeta,
            message => inputErrorRef.current(message),
        );
        playerRef.current = player;
        return () => {
            player.close();
            playerRef.current = null;
            activeTouch.current = null;
        };
    }, [mode, projectIdentifier, deviceId]);

    // the mouse wheel scrolls on the device; a native listener, since React's
    // wheel listener is passive and can't keep the page from scrolling
    useEffect(() => {
        const canvas = canvasRef.current;
        if (mode !== 'video' || !canvas) return;
        const onWheel = (e: WheelEvent) => {
            const player = playerRef.current;
            const token = leaseRef.current;
            if (!player?.canSendInput || !token) return;
            e.preventDefault();
            // pixels to lines; scrcpy scrolls up for positive values
            const scale = e.deltaMode === WheelEvent.DOM_DELTA_LINE ? 1 : 1 / 100;
            const clamp = (v: number) => Math.max(-16, Math.min(16, v));
            player.sendInput({
                kind: 'scroll', ...player.toVideo(e.clientX, e.clientY),
                scrollX: clamp(-e.deltaX * scale), scrollY: clamp(-e.deltaY * scale),
            }, token);
        };
        canvas.addEventListener('wheel', onWheel, {passive: false});
        return () => canvas.removeEventListener('wheel', onWheel);
    }, [mode]);

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

    // input errors of the video socket arrive per event; show one at a time
    inputErrorRef.current = (message: string) => {
        const now = Date.now();
        if (now - lastInputError.current < 3000) return;
        lastInputError.current = now;
        setError(message);
        refreshLock();
    };

    // the video player when input can go over the video socket, else null
    // (screenshots, iOS, not connected yet): then the REST API is used
    const liveInput = (): DeviceVideoPlayer | null =>
        mode === 'video' && lease && playerRef.current?.canSendInput ? playerRef.current : null;

    const sendKey = (key: string) => {
        const player = liveInput();
        if (player) {
            player.sendInput({kind: 'key', action: 'press', key}, lease);
            return;
        }
        run(() => pressKey(projectIdentifier, deviceId, lease!, key));
    };

    const sendText = () => {
        const player = liveInput();
        if (player && new TextEncoder().encode(text).length <= VIDEO_TEXT_MAX_BYTES) {
            player.sendInput({kind: 'text', text}, lease);
            setText('');
            return;
        }
        run(async () => {
            await inputText(projectIdentifier, deviceId, lease!, text);
            setText('');
        });
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

    // device pixels of the shown screen: screenshots are in device pixels; the
    // video may be scaled, so it uses the screen size from the stream meta,
    // turned to the orientation of the current video frame
    const deviceSize = (el: HTMLElement): { width: number, height: number } => {
        if (el instanceof HTMLImageElement) {
            return {width: el.naturalWidth, height: el.naturalHeight};
        }
        const canvas = el as HTMLCanvasElement;
        const sw = videoMeta?.screenWidth ?? 0;
        const sh = videoMeta?.screenHeight ?? 0;
        if (!sw || !sh) {
            return {width: canvas.width, height: canvas.height};
        }
        const landscape = canvas.width > canvas.height;
        return landscape === (sw > sh) ? {width: sw, height: sh} : {width: sh, height: sw};
    };

    // map a pointer position on the screen element to device pixels
    const toDevice = (el: HTMLElement, clientX: number, clientY: number) => {
        const rect = el.getBoundingClientRect();
        const size = deviceSize(el);
        const sx = (size.width || rect.width) / rect.width;
        const sy = (size.height || rect.height) / rect.height;
        return {x: Math.round((clientX - rect.left) * sx), y: Math.round((clientY - rect.top) * sy)};
    };

    const onPointerDown = (e: React.PointerEvent<HTMLElement>) => {
        e.preventDefault();
        const player = liveInput();
        if (player && !activeTouch.current) {
            // live touch: down, moves and up go to the device as they happen
            e.currentTarget.setPointerCapture(e.pointerId);
            activeTouch.current = {pointerId: e.pointerId, player};
            player.sendInput({kind: 'touch', action: 'down', pointerId: e.pointerId, ...player.toVideo(e.clientX, e.clientY)}, lease);
            return;
        }
        pointerStart.current = {x: e.clientX, y: e.clientY, t: Date.now()};
    };

    const onPointerMove = (e: React.PointerEvent<HTMLElement>) => {
        const touch = activeTouch.current;
        if (!touch || touch.pointerId !== e.pointerId) return;
        // at most one move per animation frame
        pendingMove.current = {x: e.clientX, y: e.clientY};
        if (moveFrame.current) return;
        moveFrame.current = requestAnimationFrame(() => {
            moveFrame.current = 0;
            const pos = pendingMove.current;
            const current = activeTouch.current;
            if (!pos || !current) return;
            current.player.sendInput({kind: 'touch', action: 'move', pointerId: current.pointerId, ...current.player.toVideo(pos.x, pos.y)}, leaseRef.current);
        });
    };

    // ends a live touch; returns false when the event isn't part of one
    const endTouch = (e: React.PointerEvent<HTMLElement>, action: 'up' | 'cancel'): boolean => {
        const touch = activeTouch.current;
        if (!touch || touch.pointerId !== e.pointerId) return false;
        activeTouch.current = null;
        if (moveFrame.current) {
            cancelAnimationFrame(moveFrame.current);
            moveFrame.current = 0;
        }
        touch.player.sendInput({kind: 'touch', action, pointerId: touch.pointerId, ...touch.player.toVideo(e.clientX, e.clientY)}, leaseRef.current);
        return true;
    };

    const onPointerCancel = (e: React.PointerEvent<HTMLElement>) => {
        if (!endTouch(e, 'cancel')) pointerStart.current = null;
    };

    const onPointerUp = (e: React.PointerEvent<HTMLElement>) => {
        if (endTouch(e, 'up')) return;
        const start = pointerStart.current;
        pointerStart.current = null;
        if (!start) return;
        if (!lease) {
            setError('Lock the device to control it');
            return;
        }
        const from = toDevice(e.currentTarget, start.x, start.y);
        const to = toDevice(e.currentTarget, e.clientX, e.clientY);
        const distance = Math.hypot(e.clientX - start.x, e.clientY - start.y);
        if (distance < TAP_DISTANCE) {
            run(() => tap(projectIdentifier, deviceId, lease, from.x, from.y));
        } else {
            const duration = Math.min(Math.max(Date.now() - start.t, 100), 2000);
            run(() => swipe(projectIdentifier, deviceId, lease, from.x, from.y, to.x, to.y, duration));
        }
        if (mode === 'off') {
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
    const imageSrc = mode === 'screenshots' ? streamUrl(projectIdentifier, deviceId, 2) : snapshot;
    const screenStyle: React.CSSProperties = {
        maxHeight: '75vh', maxWidth: '100%', cursor: lease ? 'crosshair' : 'not-allowed', userSelect: 'none', touchAction: 'none',
    };

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
                    <ToggleButtonGroup size="small" exclusive value={mode} onChange={(_, v: ScreenMode | null) => v && setMode(v)}>
                        <Tooltip title={canDecodeVideo
                            ? 'Live H.264 video (Android)'
                            : 'Video decoding needs a secure context: open the hub over HTTPS or localhost'}>
                            <span><ToggleButton value="video" disabled={!canDecodeVideo}>Video</ToggleButton></span>
                        </Tooltip>
                        <ToggleButton value="screenshots">Screenshots</ToggleButton>
                        <ToggleButton value="off">Off</ToggleButton>
                    </ToggleButtonGroup>
                    {mode === 'video' && videoStatus?.state === 'connecting' && <Chip size="small" label="Connecting video…"/>}
                    {mode === 'video' && videoStatus?.state === 'playing' && lease &&
                        <Tooltip title="Touch, drag, scroll and keys go through the video connection (scrcpy) instead of adb">
                            <Chip size="small" color="success" variant="outlined" label="Live input"/>
                        </Tooltip>}
                    {mode === 'video' && videoStatus?.state === 'error' && <Chip size="small" color="error" label={videoStatus.message ?? 'Video error'}/>}
                    {mode === 'off' && <Button onClick={() => setSnapshot(screenshotUrl(projectIdentifier, deviceId))}>Refresh</Button>}
                    {busy && <CircularProgress size={20}/>}
                    <Box sx={{flexGrow: 1}}/>
                    <Button onClick={() => navigate(`/project/${projectIdentifier}/device/${deviceId}`)}>Device details</Button>
                </Stack>
            </TitleCard>

            <Box sx={{display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'flex-start'}}>
                <Paper sx={{p: 1, flex: '0 1 auto', maxWidth: '100%'}}>
                    <Box sx={{position: 'relative', display: 'inline-block', lineHeight: 0}}>
                        {mode === 'video'
                            ? <canvas
                                ref={canvasRef}
                                width={360}
                                height={640}
                                onPointerDown={onPointerDown}
                                onPointerMove={onPointerMove}
                                onPointerUp={onPointerUp}
                                onPointerCancel={onPointerCancel}
                                style={{...screenStyle, backgroundColor: '#000'}}
                            />
                            : <img
                                src={imageSrc}
                                alt="device screen"
                                draggable={false}
                                onPointerDown={onPointerDown}
                                onPointerUp={onPointerUp}
                                style={screenStyle}
                            />}
                        {highlight && <Box sx={{
                            position: 'absolute', pointerEvents: 'none', border: '2px solid red',
                            backgroundColor: 'rgba(255,0,0,0.15)', ...highlight,
                        }}/>}
                    </Box>
                    {!lease && <Alert severity="info" sx={{mt: 1}}>Lock the device to tap (click), swipe (drag) and scroll (wheel) on the screen.</Alert>}
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
                                    onClick={sendText}>Type</Button>
                        </Stack>
                        <ButtonGroup variant="outlined" size="small" sx={{flexWrap: 'wrap'}}>
                            {KEYS.map(key => <Button key={key} disabled={!lease}
                                                     onClick={() => sendKey(key)}>{key}</Button>)}
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
