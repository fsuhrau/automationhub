import React, {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {Link as RouterLink, useParams} from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Checkbox,
    Chip,
    CircularProgress,
    Divider,
    FormControl,
    FormControlLabel,
    IconButton,
    InputAdornment,
    InputLabel,
    Link,
    ListItemText,
    Menu,
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
import {
    ArrowBackIosNewRounded,
    CameraAltRounded,
    CircleOutlined,
    CropSquareRounded,
    KeyboardRounded,
    LockOpenRounded,
    MoreHorizRounded,
    RefreshRounded,
    SearchRounded,
    SendRounded,
    TouchAppRounded,
    VolumeDownRounded,
    VolumeUpRounded,
} from '@mui/icons-material';
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
import {describeController} from "../../components/DeviceController";
import {DeviceVideoPlayer, VIDEO_TEXT_MAX_BYTES, VideoMeta, videoDecodingAvailable, VideoStatus, videoUrl} from '../../services/device.video';
import {useSelectedApp} from '../../hooks/useSelectedApp';
import {formatDateTime} from '../../components/DateTime';
import {defaultApp, elementAt, filterLogLines, keyForEvent, leaseNote, logLevel, runsOn, SceneElement} from './deviceControlModel';
import {PlatformType} from '../../types/platform.type.enum';
import {red} from '../../shared-theme/themePrimitives';

const TAP_DISTANCE = 12; // css px; a longer drag is a swipe

// video: H.264 stream (Android), screenshots: MJPEG of screenshots, off: manual refresh
type ScreenMode = 'video' | 'screenshots' | 'off';
const VIDEO_OPTIONS = {maxSize: 1280, maxFps: 30};
// the keys of a phone's navigation bar and side, then the rarer ones
const NAV_KEYS = [
    {key: 'back', label: 'Back', icon: <ArrowBackIosNewRounded fontSize="small"/>},
    {key: 'home', label: 'Home', icon: <CircleOutlined fontSize="small"/>},
    {key: 'app_switch', label: 'Recent apps', icon: <CropSquareRounded fontSize="small"/>},
];
const MORE_KEYS = [
    {key: 'enter', label: 'Enter'},
    {key: 'delete', label: 'Delete (backspace)'},
    {key: 'tab', label: 'Tab'},
    {key: 'menu', label: 'Menu'},
    {key: 'search', label: 'Search'},
    {key: 'power', label: 'Power (turns the screen off or on)'},
];
// key presses are sent as text once typing pauses this long
const TYPE_BATCH_MS = 60;

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
    root: Element,
    byId: Map<string, Element>,
    idOf: Map<Element, string>,
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
    const idOf = new Map<Element, string>();
    let counter = 0;
    const build = (e: Element): SceneNode => {
        const id = `n${counter++}`;
        byId.set(id, e);
        idOf.set(e, id);
        return {id, label: nodeLabel(e), element: e, children: Array.from(e.children).map(build)};
    };
    return {
        items: [build(root)],
        root,
        byId,
        idOf,
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

// the ids of an element's ancestors, to expand the tree down to it
const ancestorIds = (scene: SceneGraph, e: Element): string[] => {
    const ids: string[] = [];
    let current = e.parentElement;
    while (current) {
        const id = scene.idOf.get(current);
        if (id) ids.unshift(id);
        current = current.parentElement;
    }
    return ids;
};

const errorMessage = (error: unknown): string => {
    const e = error as { response?: { data?: unknown }, message?: string };
    let data = e.response?.data;
    // text requests (the scene graph) get the error body unparsed
    if (typeof data === 'string') {
        try {
            data = JSON.parse(data);
        } catch {
            // not JSON
        }
    }
    const message = (data as { message?: unknown } | undefined)?.message;
    return typeof message === 'string' && message ? message : e.message ?? String(error);
};

const saveBlob = (blob: Blob, name: string) => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = name;
    a.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
};

const DeviceControlPage: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {deviceId: deviceIdParam} = useParams();
    const deviceId = Number(deviceIdParam);
    const {setError} = useError();
    const {app: menuApp} = useSelectedApp();

    const [device, setDevice] = useState<IDeviceData | null>(null);
    const [lockStatus, setLockStatus] = useState<DeviceLockStatus | null>(null);
    const [lease, setLease] = useState<string | null>(() => loadLease(deviceId));
    const [busy, setBusy] = useState(false);
    const canDecodeVideo = useMemo(videoDecodingAvailable, []);
    const [mode, setMode] = useState<ScreenMode>(canDecodeVideo ? 'video' : 'screenshots');
    const [videoStatus, setVideoStatus] = useState<VideoStatus | null>(null);
    const [videoMeta, setVideoMeta] = useState<VideoMeta | null>(null);
    const [snapshot, setSnapshot] = useState(() => screenshotUrl(projectIdentifier, deviceId));
    const [tab, setTab] = useState<'app' | 'inspect' | 'logs'>('app');
    const [text, setText] = useState('');
    const [screenFocused, setScreenFocused] = useState(false);
    const [moreKeys, setMoreKeys] = useState<HTMLElement | null>(null);
    const [now, setNow] = useState(Date.now());

    const [scene, setScene] = useState<SceneGraph | null>(null);
    const [selected, setSelected] = useState<Element | null>(null);
    const [expanded, setExpanded] = useState<string[]>([]);
    const [picking, setPicking] = useState(false);
    const [sceneError, setSceneError] = useState<string | null>(null);

    const [logs, setLogs] = useState<string[] | null>(null);
    const [logLines, setLogLines] = useState(300);
    const [logFilter, setLogFilter] = useState('');
    const [problemsOnly, setProblemsOnly] = useState(false);

    const [apps, setApps] = useState<IAppData[]>([]);
    const [appId, setAppId] = useState<number | ''>('');
    const [bundles, setBundles] = useState<IAppBinaryData[]>([]);
    const [bundleId, setBundleId] = useState<number | ''>('');
    const [clearData, setClearData] = useState(false);

    const canvasRef = useRef<HTMLCanvasElement>(null);
    const imageRef = useRef<HTMLImageElement>(null);
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
    const typed = useRef('');
    const logBox = useRef<HTMLDivElement>(null);
    const typeTimer = useRef(0);

    const fail = useCallback((error: unknown) => setError(errorMessage(error)), [setError]);

    const refreshLock = useCallback(() => {
        getLockStatus(projectIdentifier, deviceId).then(status => {
            setLockStatus(status);
            setNow(Date.now());
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

    // offer the app picked in the menu when it runs on this device
    useEffect(() => {
        if (appId !== '' || apps.length === 0 || !device) return;
        const app = defaultApp(apps, device.platformType, menuApp ? menuApp.id as number : null);
        if (app) setAppId(app.id as number);
    }, [apps, device, menuApp, appId]);

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
        setBundleId('');
        getAppBundles(projectIdentifier, appId).then(list => {
            const sorted = [...list].sort((a, b) => b.id - a.id);
            setBundles(sorted);
            // the newest bundle is the usual choice
            if (sorted.length > 0) setBundleId(sorted[0].id);
        }).catch(fail);
    }, [projectIdentifier, appId, fail]);

    useEffect(() => () => window.clearTimeout(typeTimer.current), []);

    // the newest lines are at the end; start there after loading
    useEffect(() => {
        if (logBox.current) logBox.current.scrollTop = logBox.current.scrollHeight;
    }, [logs]);

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

    const typeText = (value: string, done?: () => void) => {
        if (!value) return;
        const player = liveInput();
        if (player && new TextEncoder().encode(value).length <= VIDEO_TEXT_MAX_BYTES) {
            player.sendInput({kind: 'text', text: value}, lease);
            done?.();
            return;
        }
        run(async () => {
            await inputText(projectIdentifier, deviceId, lease!, value);
            done?.();
        });
    };

    const flushTyped = () => {
        window.clearTimeout(typeTimer.current);
        const value = typed.current;
        typed.current = '';
        typeText(value);
    };

    // typing while the screen has focus goes to the device: characters as
    // text (batched, so a fast typist sends words), special keys by name
    const onScreenKey = (e: React.KeyboardEvent) => {
        if (!lease) return;
        const key = keyForEvent(e.nativeEvent);
        if (!key) return;
        e.preventDefault();
        if ('text' in key) {
            typed.current += key.text;
            window.clearTimeout(typeTimer.current);
            typeTimer.current = window.setTimeout(flushTyped, TYPE_BATCH_MS);
            return;
        }
        flushTyped();
        sendKey(key.key);
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

    const select = (e: Element | null) => {
        setSelected(e);
        if (e && scene) setExpanded(prev => Array.from(new Set([...prev, ...ancestorIds(scene, e)])));
    };

    // in pick mode a click selects the element under it instead of tapping
    const pickAt = (el: HTMLElement, clientX: number, clientY: number) => {
        if (!scene || !scene.width || !scene.height) return;
        const rect = el.getBoundingClientRect();
        const x = ((clientX - rect.left) / rect.width) * scene.width;
        const y = ((clientY - rect.top) / rect.height) * scene.height;
        select(elementAt(scene.root as unknown as SceneElement, x, y) as unknown as Element | null);
        setTab('inspect');
    };

    const onPointerDown = (e: React.PointerEvent<HTMLElement>) => {
        e.preventDefault();
        (e.currentTarget.parentElement as HTMLElement | null)?.focus();
        if (picking) {
            pickAt(e.currentTarget, e.clientX, e.clientY);
            return;
        }
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
        if (!start || !lease) return;
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

    // a missing app is an expected answer here, so it shows in the tab
    const loadHierarchy = async () => {
        setBusy(true);
        setSceneError(null);
        try {
            const graph = parseSceneGraph(await getHierarchy(projectIdentifier, deviceId));
            setScene(graph);
            setSelected(null);
            setExpanded(graph.items.map(i => i.id));
        } catch (error) {
            setSceneError(errorMessage(error));
        } finally {
            setBusy(false);
        }
    };

    const loadLogs = () => run(async () => setLogs(await getDeviceLogs(projectIdentifier, deviceId, logLines)));

    const onAppAction = (action: AppAction) => {
        if (!lease || bundleId === '') return;
        run(() => appAction(projectIdentifier, deviceId, lease, action, bundleId, action === 'stop' && clearData));
    };

    const deviceName = device ? (device.alias || device.name) : `Device ${deviceId}`;

    const saveScreenshot = () => run(async () => {
        const name = `${deviceName}-${formatDateTime(Date.now(), 'YYYYMMDD-HHmmss')}.png`.replace(/[^\w.-]+/g, '_');
        const canvas = canvasRef.current;
        if (mode === 'video' && canvas && videoStatus?.state === 'playing') {
            const blob = await new Promise<Blob | null>(resolve => canvas.toBlob(resolve, 'image/png'));
            if (blob) return saveBlob(blob, name);
        }
        const response = await fetch(screenshotUrl(projectIdentifier, deviceId), {credentials: 'include'});
        if (!response.ok) throw new Error(`screenshot failed: ${response.status}`);
        saveBlob(await response.blob(), name);
    });

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
    const usedByTest = !holder && !!lockStatus?.lockedByTest;
    const imageSrc = mode === 'screenshots' ? streamUrl(projectIdentifier, deviceId, 2) : snapshot;
    const interactive = !!lease || picking;
    const screenStyle: React.CSSProperties = {
        maxHeight: '72vh', maxWidth: '100%', display: 'block', userSelect: 'none', touchAction: 'none',
        cursor: picking ? 'help' : lease ? 'crosshair' : 'default',
    };
    const screenHandlers = interactive ? {onPointerDown, onPointerMove, onPointerUp, onPointerCancel} : {};
    const live = mode === 'video' && videoStatus?.state === 'playing' && !!lease;
    const visibleLogs = useMemo(() => logs ? filterLogLines(logs, logFilter, problemsOnly) : [], [logs, logFilter, problemsOnly]);
    const fittingApps = device ? apps.filter(a => runsOn(a.platform, device.platformType)) : apps;
    const otherApps = apps.filter(a => !fittingApps.includes(a));

    const status = ownLease
        ? {color: 'success' as const, label: 'You control this device', note: leaseNote(lockStatus?.lease?.expiresAt, now)}
        : holder
            ? {color: 'warning' as const, label: `Controlled by ${describeController(holder).label}`,
                note: lockStatus?.lease?.acquiredAt ? `since ${formatDateTime(lockStatus.lease.acquiredAt, 'HH:mm')}` : null}
            : usedByTest
                ? {color: 'warning' as const, label: 'A test run is using this device', note: 'Control is possible when it ends.'}
                : {color: 'default' as const, label: 'Free', note: 'Take control to use the device.'};

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            {/* which device, who controls it, and taking or giving back control */}
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{deviceName}</Typography>
                    <Typography variant="body2" color="text.secondary">
                        {device ? <>{device.os} {device.osVersion}
                            {device.alias && device.alias !== device.name && <> · {device.name}</>}
                            {device.node && <> · on {device.node.name}</>}</> : ' '}
                        {' · '}<Link component={RouterLink} to={`/project/${projectIdentifier}/device/${deviceId}`}
                                     underline="hover" color="inherit">Device details</Link>
                    </Typography>
                </Box>
                <Stack direction="row" spacing={1.5} alignItems="center" sx={{ml: {md: 'auto'}}} useFlexGap flexWrap="wrap">
                    <Box sx={{textAlign: {md: 'right'}}}>
                        <Chip size="small" color={status.color} label={status.label}/>
                        {status.note && <Typography variant="caption" color="text.secondary" component="div">{status.note}</Typography>}
                    </Box>
                    {ownLease
                        ? <Tooltip title="Give the device back so tests and others can use it">
                            <Button variant="outlined" startIcon={<LockOpenRounded/>} onClick={onRelease} disabled={busy}>Release</Button>
                        </Tooltip>
                        : <Button variant="contained" startIcon={<TouchAppRounded/>} onClick={onLock}
                                  disabled={busy || lockStatus?.locked}>Take control</Button>}
                </Stack>
            </Stack>

            <Box sx={{display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'flex-start'}}>
                {/* the screen with what can be done on it */}
                <Card variant="outlined" sx={{flex: '0 1 auto', maxWidth: '100%'}}>
                    <CardContent sx={{p: 1.5, '&:last-child': {pb: 1.5}}}>
                        <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap" sx={{mb: 1}}>
                            <ToggleButtonGroup size="small" exclusive value={mode} aria-label="Screen"
                                               onChange={(_, v: ScreenMode | null) => v && setMode(v)}
                                               // as tall as the icon buttons beside it; the theme's toggles are form-sized
                                               sx={{'& .MuiToggleButton-root': {py: 0.5, px: 1.25, height: 36, fontSize: 12, lineHeight: 1}}}>
                                <Tooltip title={canDecodeVideo
                                    ? 'Live video with fast input (Android)'
                                    : 'Video needs a secure context: open the hub over HTTPS or localhost'}>
                                    <span><ToggleButton value="video" disabled={!canDecodeVideo}>Video</ToggleButton></span>
                                </Tooltip>
                                <Tooltip title="A screenshot every half second; works on every device">
                                    <ToggleButton value="screenshots">Screenshots</ToggleButton>
                                </Tooltip>
                                <Tooltip title="No updates; refresh by hand">
                                    <ToggleButton value="off">Paused</ToggleButton>
                                </Tooltip>
                            </ToggleButtonGroup>
                            {mode === 'off' && <Tooltip title="Refresh the screen">
                                <IconButton size="small" aria-label="Refresh the screen"
                                            onClick={() => setSnapshot(screenshotUrl(projectIdentifier, deviceId))}><RefreshRounded/></IconButton>
                            </Tooltip>}
                            <Tooltip title="Save a screenshot">
                                <IconButton size="small" aria-label="Save a screenshot" onClick={saveScreenshot} disabled={busy}>
                                    <CameraAltRounded/>
                                </IconButton>
                            </Tooltip>
                            <Box sx={{flexGrow: 1}}/>
                            {busy && <CircularProgress size={18}/>}
                            {mode === 'video' && videoStatus?.state === 'connecting' && <Chip size="small" label="Connecting…"/>}
                            {mode === 'video' && videoStatus?.state === 'error' && <Chip size="small" color="error" label={videoStatus.message ?? 'Video error'}/>}
                            {live && <Tooltip title="Touch, drag, scroll and keys go straight through the video connection">
                                <Chip size="small" color="success" variant="outlined" label="Live"/>
                            </Tooltip>}
                        </Stack>

                        {/* the keys where a phone has them in portrait: the
                            navigation bar below the screen, volume on the right edge */}
                        <Box sx={{display: 'flex', alignItems: 'flex-start', gap: 1, maxWidth: '100%'}}>
                            <Box sx={{display: 'flex', flexDirection: 'column', alignItems: 'center', minWidth: 0}}>
                                <Box tabIndex={lease ? 0 : -1} onKeyDown={onScreenKey}
                                     onFocus={() => setScreenFocused(true)} onBlur={() => {
                                         setScreenFocused(false);
                                         flushTyped();
                                     }}
                                     aria-label="Device screen"
                                     sx={{
                                         // clip the video, screenshot and overlay to the rounded corners
                                         position: 'relative', display: 'inline-block', lineHeight: 0, borderRadius: 1, outline: 'none', overflow: 'hidden',
                                         boxShadow: screenFocused ? (t => `0 0 0 2px ${(t.vars ?? t).palette.primary.main}`) : undefined,
                                     }}>
                                    {mode === 'video'
                                        ? <canvas ref={canvasRef} width={360} height={640} {...screenHandlers}
                                                  style={{...screenStyle, backgroundColor: '#000'}}/>
                                        : <img ref={imageRef} src={imageSrc} alt="device screen" draggable={false}
                                               {...screenHandlers} style={screenStyle}/>}
                                    {highlight && <Box sx={{
                                        position: 'absolute', pointerEvents: 'none', border: '2px solid', borderColor: 'error.main',
                                        backgroundColor: 'rgba(255,0,0,0.15)', ...highlight,
                                    }}/>}
                                    {!lease && !picking && <Box sx={{
                                        position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center',
                                        bgcolor: 'rgba(0,0,0,0.45)', lineHeight: 1.5, p: 2,
                                    }}>
                                        <Paper sx={{p: 2, maxWidth: 260, textAlign: 'center'}}>
                                            <Typography variant="subtitle2" gutterBottom>{status.label}</Typography>
                                            <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>
                                                {holder || usedByTest
                                                    ? 'You can watch the screen. Control is possible once the device is free.'
                                                    : 'Take control to tap, swipe, scroll and type on the device.'}
                                            </Typography>
                                            {!holder && !usedByTest && <Button variant="contained" size="small" startIcon={<TouchAppRounded/>}
                                                                               onClick={onLock} disabled={busy || lockStatus?.locked}>
                                                Take control
                                            </Button>}
                                        </Paper>
                                    </Box>}
                                </Box>
                                <Stack direction="row" alignItems="center" justifyContent="center" spacing={0.5} sx={{mt: 1}}
                                       role="group" aria-label="Navigation keys">
                                    {NAV_KEYS.map(k => <Tooltip key={k.key} title={k.label}><span>
                                        <IconButton aria-label={k.label} disabled={!lease} onClick={() => sendKey(k.key)} sx={{mx: 1}}>{k.icon}</IconButton>
                                    </span></Tooltip>)}
                                    <Tooltip title="More keys"><span><IconButton size="small" aria-label="More keys" disabled={!lease}
                                                                                 onClick={e => setMoreKeys(e.currentTarget)}><MoreHorizRounded fontSize="small"/></IconButton></span></Tooltip>
                                    <Menu anchorEl={moreKeys} open={!!moreKeys} onClose={() => setMoreKeys(null)}>
                                        {MORE_KEYS.map(k => <MenuItem key={k.key} onClick={() => {
                                            setMoreKeys(null);
                                            sendKey(k.key);
                                        }}>{k.label}</MenuItem>)}
                                    </Menu>
                                </Stack>
                            </Box>
                            <Stack spacing={0.5} sx={{mt: 10}} role="group" aria-label="Volume keys">
                                <Tooltip title="Volume up" placement="right"><span><IconButton size="small" aria-label="Volume up" disabled={!lease}
                                                                                               onClick={() => sendKey('volume_up')}><VolumeUpRounded fontSize="small"/></IconButton></span></Tooltip>
                                <Tooltip title="Volume down" placement="right"><span><IconButton size="small" aria-label="Volume down" disabled={!lease}
                                                                                                 onClick={() => sendKey('volume_down')}><VolumeDownRounded fontSize="small"/></IconButton></span></Tooltip>
                            </Stack>
                        </Box>
                        <TextField size="small" fullWidth placeholder="Type or paste text, then Enter" value={text} disabled={!lease}
                                   onChange={e => setText(e.target.value)}
                                   onKeyDown={e => {
                                       if (e.key === 'Enter' && text) {
                                           e.preventDefault();
                                           typeText(text, () => setText(''));
                                       }
                                   }}
                                   sx={{mt: 1}}
                                   slotProps={{
                                       htmlInput: {'aria-label': 'Text to type on the device'},
                                       input: {
                                           endAdornment: <InputAdornment position="end">
                                               <IconButton size="small" edge="end" aria-label="Type on the device" disabled={!lease || !text}
                                                           onClick={() => typeText(text, () => setText(''))}><SendRounded fontSize="small"/></IconButton>
                                           </InputAdornment>,
                                       },
                                   }}/>
                        {lease && <Typography variant="caption" color="text.secondary" component="div" sx={{mt: 0.5, display: 'flex', gap: 0.5, alignItems: 'center'}}>
                            <KeyboardRounded fontSize="inherit"/>
                            {screenFocused ? 'Your keyboard types on the device; Esc is Back.' : 'Click the screen to type on it with your keyboard.'}
                        </Typography>}
                    </CardContent>
                </Card>

                {/* what can be done beyond the screen */}
                <Card variant="outlined" sx={{flex: '1 1 420px', minWidth: 320}}>
                    <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{px: 1, borderBottom: 1, borderColor: 'divider'}}>
                        <Tab value="app" label="App"/>
                        <Tab value="inspect" label="Inspect"/>
                        <Tab value="logs" label="System log"/>
                    </Tabs>
                    <CardContent>
                        {tab === 'app' && <Stack spacing={2}>
                            <Typography variant="body2" color="text.secondary">
                                Install, start or stop a bundle on this device.
                            </Typography>
                            <FormControl size="small" fullWidth>
                                <InputLabel id="control-app">App</InputLabel>
                                <Select labelId="control-app" label="App" value={appId}
                                        onChange={e => setAppId(e.target.value === '' ? '' : Number(e.target.value))}>
                                    {fittingApps.map(a => <MenuItem key={a.id} value={a.id}>{a.name}</MenuItem>)}
                                    {otherApps.length > 0 && <Divider/>}
                                    {otherApps.map(a => <MenuItem key={a.id} value={a.id}>
                                        <ListItemText primary={a.name} secondary="other platform"/>
                                    </MenuItem>)}
                                </Select>
                            </FormControl>
                            <FormControl size="small" fullWidth disabled={appId === '' || bundles.length === 0}>
                                <InputLabel id="control-bundle">Bundle</InputLabel>
                                <Select labelId="control-bundle" label="Bundle" value={bundleId}
                                        renderValue={id => {
                                            const b = bundles.find(x => x.id === id);
                                            return b ? `${b.version || b.name} (#${b.id})` : '';
                                        }}
                                        onChange={e => setBundleId(e.target.value === '' ? '' : Number(e.target.value))}>
                                    {bundles.map((b, i) => <MenuItem key={b.id} value={b.id}>
                                        <ListItemText primary={`${b.version || b.name}${i === 0 ? ' (newest)' : ''}`}
                                                      secondary={`#${b.id} · uploaded ${formatDateTime(b.createdAt, 'YYYY/MM/DD HH:mm')}`}/>
                                    </MenuItem>)}
                                </Select>
                            </FormControl>
                            {appId !== '' && bundles.length === 0 && <Typography variant="body2" color="text.secondary">
                                This app has no bundles. <Link component={RouterLink} to={`/project/${projectIdentifier}/app:${appId}/bundles`}>Bundles</Link>
                            </Typography>}
                            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                                <Button variant="outlined" disabled={!lease || bundleId === '' || busy} onClick={() => onAppAction('install')}>Install</Button>
                                <Button variant="contained" disabled={!lease || bundleId === '' || busy} onClick={() => onAppAction('start')}>Start</Button>
                                <Button variant="outlined" color="warning" disabled={!lease || bundleId === '' || busy} onClick={() => onAppAction('stop')}>Stop</Button>
                            </Stack>
                            <FormControlLabel control={<Checkbox size="small" checked={clearData} onChange={(_, v) => setClearData(v)}/>}
                                              label={<Typography variant="body2">Clear the app's data when stopping</Typography>}/>
                            {!lease && <Typography variant="caption" color="text.secondary">Take control to use these.</Typography>}
                        </Stack>}

                        {tab === 'inspect' && <Box>
                            <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>
                                The scene graph of the running app, for finding elements and their XPath.
                                Needs the app connected to the hub (start it from the App tab).
                            </Typography>
                            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{mb: 1.5}}>
                                <Button variant="outlined" startIcon={<RefreshRounded/>} onClick={loadHierarchy} disabled={busy}>
                                    {scene ? 'Reload' : 'Load scene graph'}
                                </Button>
                                <Tooltip title="Click an element on the screen to find it in the tree; the click doesn't reach the device">
                                    <span><ToggleButton size="small" value="pick" selected={picking} disabled={!scene}
                                                        onChange={() => setPicking(p => !p)}>Pick on screen</ToggleButton></span>
                                </Tooltip>
                            </Stack>
                            {sceneError && <Alert severity="warning" sx={{mb: 1.5}}
                                                  action={<Button color="inherit" size="small" onClick={() => setTab('app')}>App tab</Button>}>
                                {sceneError}
                            </Alert>}
                            {selected && <Paper variant="outlined" sx={{p: 1.5, mb: 1.5}}>
                                <Stack direction="row" spacing={1} alignItems="center" useFlexGap>
                                    <Typography variant="subtitle2" sx={{wordBreak: 'break-word'}}>{nodeLabel(selected)}</Typography>
                                    <Button size="small" variant="contained" sx={{ml: 'auto'}}
                                            disabled={!lease || !selected.getAttribute('ID')}
                                            onClick={() => run(() => touchElement(projectIdentifier, deviceId, lease!, {elementId: selected.getAttribute('ID')!}))}>
                                        Tap it
                                    </Button>
                                </Stack>
                                <Stack direction="row" spacing={1} alignItems="center" sx={{mt: 1}}>
                                    <Typography variant="body2" sx={{fontFamily: 'monospace', wordBreak: 'break-all', flex: 1}}>{xpathFor(selected)}</Typography>
                                    <Button size="small" onClick={() => navigator.clipboard.writeText(xpathFor(selected)).catch(fail)}>Copy XPath</Button>
                                </Stack>
                                <Box sx={{mt: 1, display: 'grid', gridTemplateColumns: 'max-content 1fr', columnGap: 2, rowGap: 0.25}}>
                                    {Array.from(selected.attributes).map(a => <React.Fragment key={a.name}>
                                        <Typography variant="caption" color="text.secondary">{a.name}</Typography>
                                        <Typography variant="caption" sx={{fontFamily: 'monospace', wordBreak: 'break-all'}}>{a.value}</Typography>
                                    </React.Fragment>)}
                                </Box>
                            </Paper>}
                            {scene && <Box sx={{maxHeight: '50vh', overflow: 'auto'}}>
                                <RichTreeView items={scene.items}
                                              selectedItems={selected ? scene.idOf.get(selected) ?? null : null}
                                              expandedItems={expanded}
                                              onExpandedItemsChange={(_, ids) => setExpanded(ids)}
                                              onSelectedItemsChange={(_, id) => select(id ? scene.byId.get(id as string) ?? null : null)}/>
                            </Box>}
                        </Box>}

                        {tab === 'logs' && <Box>
                            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{mb: 1.5}}>
                                <Button variant="outlined" startIcon={<RefreshRounded/>} onClick={loadLogs} disabled={busy}>
                                    {logs ? 'Refresh' : 'Load'}
                                </Button>
                                <Select size="small" value={logLines} onChange={e => setLogLines(Number(e.target.value))}
                                        inputProps={{'aria-label': 'Lines'}}>
                                    {[300, 1000, 3000].map(n => <MenuItem key={n} value={n}>last {n} lines</MenuItem>)}
                                </Select>
                                {logs && <>
                                    <TextField size="small" placeholder="Find" value={logFilter} onChange={e => setLogFilter(e.target.value)}
                                               sx={{width: 180}}
                                               slotProps={{
                                                   input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                                                   htmlInput: {'aria-label': 'Find in the system log'},
                                               }}/>
                                    <Chip size="small" label="Warnings and errors" variant={problemsOnly ? 'filled' : 'outlined'}
                                          color={problemsOnly ? 'warning' : 'default'} onClick={() => setProblemsOnly(v => !v)}/>
                                </>}
                            </Stack>
                            {logs === null
                                ? <Typography variant="body2" color="text.secondary">The device's system log (logcat on Android), newest last.</Typography>
                                : <>
                                    <Typography variant="caption" color="text.secondary">
                                        {visibleLogs.length === logs.length ? `${logs.length} lines` : `${visibleLogs.length} of ${logs.length} lines`}
                                    </Typography>
                                    <Box ref={logBox} sx={{
                                        mt: 0.5, p: 1, maxHeight: '50vh', overflow: 'auto', fontFamily: 'monospace', fontSize: 12,
                                        bgcolor: 'background.default', border: 1, borderColor: 'divider', borderRadius: 1,
                                    }} role="log" aria-label="System log">
                                        {visibleLogs.map((line, i) => {
                                            const level = logLevel(line);
                                            const error = level === 'E' || level === 'F';
                                            return <Box key={i} sx={t => ({
                                                whiteSpace: 'pre-wrap', wordBreak: 'break-all',
                                                color: error ? 'error.main' : level === 'W' ? 'warning.main' : 'text.primary',
                                                // the lighter red of the ramp reads on the dark surface
                                                ...(error ? t.applyStyles('dark', {color: red[300]}) : {}),
                                            })}>{line}</Box>;
                                        })}
                                    </Box>
                                </>}
                        </Box>}
                    </CardContent>
                </Card>
            </Box>

            {lease && mode !== 'video' && device?.platformType === PlatformType.Android && canDecodeVideo &&
                <Alert severity="info" sx={{mt: 2}}>Video mode is faster and sends touches and keys live on Android.</Alert>}
        </Box>
    );
};

export default DeviceControlPage;
