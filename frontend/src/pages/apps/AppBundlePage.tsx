import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {
    Alert,
    Avatar,
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
    Grid,
    IconButton,
    LinearProgress,
    Link,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    ArrowBackRounded,
    CheckCircleRounded,
    ContentCopyRounded,
    DeleteOutlineRounded,
    DownloadRounded,
    ErrorOutlineRounded,
    RefreshRounded,
    WarningAmberRounded,
} from '@mui/icons-material';
import {Link as RouterLink, useNavigate, useParams} from 'react-router-dom';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {useError} from '../../ErrorProvider';
import {IAppBundleDetail} from '../../types/app';
import {AndroidInfo, AppleInfo, BundleInfo, Certificate, Contents, ProvisioningProfile} from '../../types/bundle.info';
import IDeviceData from '../../types/device';
import {bundleIconUrl, deleteAppBundle, getAppBundle, readAppBundleAgain} from '../../services/app.service';
import {getAllDevices} from '../../services/device.service';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import {formatDateTime} from '../../components/DateTime';
import {byteFormat} from '../tests/value_formatter';
import {STATUS} from '../tests/testsListModel';
import {RunSummaryStatus} from '../../types/test';
import {ago, parseTags} from './bundlesModel';
import {
    apiLevel,
    commonName,
    deviceFit,
    entitlementValue,
    expiry,
    fileName,
    isDangerous,
    orientationName,
    platformsOf,
    PROFILE_KINDS,
    shortPermission,
    usageName,
} from './bundleDetailModel';

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};

const mono = {fontFamily: 'monospace', fontSize: '0.8125rem', wordBreak: 'break-all'} as const;

const copy = (text: string) => navigator.clipboard.writeText(text).catch(() => undefined);

// Fact is one label and value; values that are ids can be copied
interface FactProps {
    label: string,
    value?: React.ReactNode,
    copyText?: string,
    monospace?: boolean,
}

const Fact: React.FC<FactProps> = ({label, value, copyText, monospace}) => {
    if (value === undefined || value === null || value === '' || (Array.isArray(value) && value.length === 0)) return null;
    return <Box component="div" sx={{display: 'grid', gridTemplateColumns: {xs: '1fr', sm: '180px 1fr'}, gap: {xs: 0, sm: 2}, py: 0.75,
        borderBottom: 1, borderColor: 'divider', '&:last-child': {borderBottom: 0}}}>
        <Typography variant="body2" color="text.secondary">{label}</Typography>
        <Stack direction="row" spacing={0.5} alignItems="center" sx={{minWidth: 0}}>
            <Typography component="div" variant="body2" sx={monospace ? mono : {wordBreak: 'break-word'}}>{value}</Typography>
            {copyText && <Tooltip title="Copy">
                <IconButton size="small" aria-label={`Copy ${label}`} onClick={() => copy(copyText)}>
                    <ContentCopyRounded sx={{fontSize: 14}}/>
                </IconButton>
            </Tooltip>}
        </Stack>
    </Box>;
};

const Section: React.FC<{ title: string, subtitle?: React.ReactNode, action?: React.ReactNode, children: React.ReactNode }> =
    ({title, subtitle, action, children}) =>
        <Card variant="outlined" sx={{height: '100%'}}>
            <CardContent>
                <Stack direction="row" alignItems="flex-start" spacing={1} sx={{mb: 1}}>
                    <Box sx={{flex: 1, minWidth: 0}}>
                        <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
                        {subtitle && <Typography variant="body2" color="text.secondary">{subtitle}</Typography>}
                    </Box>
                    {action}
                </Stack>
                {children}
            </CardContent>
        </Card>;

const Chips: React.FC<{ items?: string[], max?: number }> = ({items, max = 40}) => {
    const [all, setAll] = useState(false);
    if (!items || items.length === 0) return null;
    const shown = all ? items : items.slice(0, max);
    return <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap>
        {shown.map(i => <Chip key={i} size="small" variant="outlined" label={i}/>)}
        {items.length > shown.length && <Chip size="small" label={`+${items.length - shown.length} more`} onClick={() => setAll(true)}/>}
    </Stack>;
};

const YesNo: React.FC<{ value: boolean, yes: string, warn?: boolean }> = ({value, yes, warn}) =>
    value ? <Chip size="small" color={warn ? 'warning' : 'default'} label={yes}/> : null;

const ExpiryText: React.FC<{ date: string, now: number }> = ({date, now}) => {
    const e = expiry(date, now);
    return <Stack direction="row" spacing={0.75} alignItems="center" component="span">
        <span>{formatDateTime(date)}</span>
        {e === 'expired' && <Chip size="small" color="error" icon={<ErrorOutlineRounded/>} label="expired"/>}
        {e === 'soon' && <Chip size="small" color="warning" icon={<WarningAmberRounded/>}
                               label={`${Math.ceil((new Date(date).valueOf() - now) / 86400000)} days left`}/>}
    </Stack>;
};

const AndroidSection: React.FC<{ android: AndroidInfo }> = ({android}) =>
    <Section title="Android" subtitle="From the manifest, as aapt reads it">
        <Fact label="Package" value={android.packageName} copyText={android.packageName} monospace/>
        <Fact label="Version name" value={android.versionName}/>
        <Fact label="Version code" value={android.versionCode}/>
        <Fact label="Min SDK" value={apiLevel(android.minSdk)}/>
        <Fact label="Target SDK" value={apiLevel(android.targetSdk)}/>
        {android.maxSdk && <Fact label="Max SDK" value={apiLevel(android.maxSdk)}/>}
        <Fact label="Compile SDK" value={android.compileSdk && apiLevel(android.compileSdk)}/>
        <Fact label="Launch activity" value={android.launchActivity} copyText={android.launchActivity} monospace/>
        <Fact label="ABIs" value={android.abis?.length ? <Chips items={android.abis}/> : undefined}/>
        <Fact label="OpenGL ES" value={android.gles}/>
        <Fact label="Screens" value={android.screens?.join(', ')}/>
        <Fact label="Densities" value={android.densities?.map(d => d === '65534' ? 'any' : d === '65535' ? 'none' : d).join(', ')}/>
        <Fact label={`Locales (${android.locales?.length ?? 0})`} value={android.locales?.length ? <Chips items={android.locales} max={24}/> : undefined}/>
        <Fact label="Flags" value={(android.debuggable || android.testOnly || android.leanback) ? <Stack direction="row" spacing={0.5}>
            <YesNo value={android.debuggable} yes="debuggable" warn/>
            <YesNo value={android.testOnly} yes="test only" warn/>
            <YesNo value={android.leanback} yes="Android TV"/>
        </Stack> : undefined}/>
    </Section>;

const PermissionsSection: React.FC<{ android: AndroidInfo }> = ({android}) => {
    const permissions = android.permissions ?? [];
    const features = android.features ?? [];
    const dangerous = permissions.filter(p => isDangerous(p.name)).length;
    return <Section title={`Permissions (${permissions.length})`}
                    subtitle={dangerous > 0 ? `${dangerous} need the player's consent at runtime` : undefined}>
        {permissions.length === 0
            ? <Typography variant="body2" color="text.secondary">None requested.</Typography>
            : <Stack spacing={0.25} sx={{mb: 2}}>
                {permissions.map(p => <Stack key={p.name} direction="row" spacing={1} alignItems="center">
                    <Typography variant="body2" sx={mono}>{shortPermission(p.name)}</Typography>
                    {isDangerous(p.name) && <Chip size="small" variant="outlined" color="warning" label="runtime"/>}
                    {p.maxSdk && <Typography variant="caption" color="text.secondary">up to API {p.maxSdk}</Typography>}
                </Stack>)}
            </Stack>}
        {features.length > 0 && <>
            <Typography variant="subtitle2" sx={{mt: 1, mb: 0.5}}>Hardware features ({features.length})</Typography>
            <Stack spacing={0.25}>
                {features.map(f => <Stack key={f.name} direction="row" spacing={1} alignItems="center">
                    <Typography variant="body2" sx={mono}>{f.name.replace(/^android\.hardware\./, '')}</Typography>
                    <Typography variant="caption" color="text.secondary">{f.required ? 'required' : 'optional'}</Typography>
                </Stack>)}
            </Stack>
        </>}
    </Section>;
};

const AppleSection: React.FC<{ apple: AppleInfo, info: BundleInfo }> = ({apple, info}) =>
    <Section title={apple.platform === 'macosx' ? 'macOS' : 'iOS'} subtitle="From the Info.plist">
        <Fact label="Bundle identifier" value={info.identifier} copyText={info.identifier} monospace/>
        <Fact label="Version" value={info.version}/>
        <Fact label="Build" value={info.build}/>
        <Fact label="Bundle name" value={apple.bundleName}/>
        <Fact label="Executable" value={apple.executable} monospace/>
        <Fact label="Platform" value={apple.platform}/>
        <Fact label="Minimum OS" value={apple.minOs}/>
        <Fact label="SDK" value={apple.sdk}/>
        <Fact label="Xcode" value={apple.xcode && `${apple.xcode}${apple.xcodeBuild ? ` (${apple.xcodeBuild})` : ''}`}/>
        <Fact label="Built on macOS" value={apple.buildMachineOs}/>
        <Fact label="Devices" value={apple.deviceFamily?.join(', ')}/>
        <Fact label="Orientations" value={apple.orientations?.map(orientationName).join(', ')}/>
        <Fact label="Requires" value={apple.requiredCapabilities?.length ? <Chips items={apple.requiredCapabilities}/> : undefined}/>
        <Fact label="Background modes" value={apple.backgroundModes?.length ? <Chips items={apple.backgroundModes}/> : undefined}/>
        <Fact label="URL schemes" value={apple.urlSchemes?.length ? <Chips items={apple.urlSchemes}/> : undefined}/>
        <Fact label="File sharing" value={apple.fileSharing ? 'on' : undefined}/>
        <Fact label={`Extensions (${apple.extensions?.length ?? 0})`} value={apple.extensions?.length ? <Chips items={apple.extensions}/> : undefined}/>
        <Fact label={`Frameworks (${apple.frameworks?.length ?? 0})`} value={apple.frameworks?.length ? <Chips items={apple.frameworks} max={16}/> : undefined}/>
    </Section>;

const UsageSection: React.FC<{ apple: AppleInfo }> = ({apple}) => {
    const usage = apple.usageDescriptions ?? [];
    return <Section title={`Permissions (${usage.length})`} subtitle="What the app tells the player when it asks for access">
        {usage.length === 0
            ? <Typography variant="body2" color="text.secondary">The app asks for no protected access.</Typography>
            : <Stack spacing={0}>{usage.map(u => <Fact key={u.key} label={usageName(u.key)} value={u.text}/>)}</Stack>}
    </Section>;
};

const ProfileSection: React.FC<{ profile: ProvisioningProfile, devices: IDeviceData[], now: number }> = ({profile, devices, now}) => {
    const [showAll, setShowAll] = useState(false);
    const [showEntitlements, setShowEntitlements] = useState(false);
    const known = new Map(devices.map(d => [d.deviceIdentifier?.toLowerCase(), d]));
    const udids = profile.devices ?? [];
    const inPool = udids.filter(u => known.has(u.toLowerCase()));
    const shown = showAll ? udids : inPool;
    const entitlements = Object.entries(profile.entitlements ?? {}).sort(([a], [b]) => a.localeCompare(b));
    return <Section title="Provisioning profile" subtitle={PROFILE_KINDS[profile.kind] ?? profile.kind}>
        <Fact label="Name" value={profile.name}/>
        <Fact label="Team" value={profile.teamName && `${profile.teamName}${profile.teamId ? ` (${profile.teamId})` : ''}`}/>
        <Fact label="App ID" value={profile.appIdName}/>
        <Fact label="UUID" value={profile.uuid} copyText={profile.uuid} monospace/>
        <Fact label="Created" value={formatDateTime(profile.createdAt)}/>
        <Fact label="Expires" value={<ExpiryText date={profile.expiresAt} now={now}/>}/>
        <Fact label="Devices" value={profile.allDevices ? 'every device (enterprise)'
            : udids.length === 0 ? 'none (installs from the App Store or TestFlight)'
                : `${udids.length} devices, ${inPool.length} of them in the device pool`}/>
        {udids.length > 0 && <Box sx={{mt: 1}}>
            {shown.map(u => {
                const d = known.get(u.toLowerCase());
                return <Stack key={u} direction="row" spacing={1} alignItems="center" sx={{py: 0.25}}>
                    <Typography variant="body2" sx={{...mono, flex: '0 1 auto'}}>{u}</Typography>
                    {d && <Chip size="small" color="primary" variant="outlined" label={d.alias || d.name}/>}
                </Stack>;
            })}
            {udids.length > inPool.length && <Button size="small" onClick={() => setShowAll(s => !s)}>
                {showAll ? 'Show only devices in the pool' : `Show all ${udids.length} devices`}
            </Button>}
        </Box>}
        {entitlements.length > 0 && <Box sx={{mt: 1}}>
            <Button size="small" onClick={() => setShowEntitlements(s => !s)}>
                {showEntitlements ? 'Hide entitlements' : `Show ${entitlements.length} entitlements`}
            </Button>
            {showEntitlements && entitlements.map(([k, v]) => <Fact key={k} label={k} value={entitlementValue(v)} monospace/>)}
        </Box>}
    </Section>;
};

const CertificatesSection: React.FC<{ certificates: Certificate[], schemes?: string[], apple: boolean, now: number }> =
    ({certificates, schemes, apple, now}) =>
        <Section title="Signing"
                 subtitle={apple ? 'Certificates the provisioning profile allows to sign the app'
                     : schemes?.length ? `APK signature scheme ${schemes.join(', ')}` : 'No signature found'}>
            {certificates.length === 0 && <Alert severity="warning">The bundle is not signed; devices won't install it.</Alert>}
            {certificates.map(c => <Box key={c.sha256} sx={{mb: 1.5, '&:last-child': {mb: 0}}}>
                <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 0.5}}>
                    <Typography variant="subtitle2">{commonName(c.subject)}</Typography>
                    {c.debug && <Chip size="small" color="warning" label="debug key"/>}
                </Stack>
                <Fact label="Subject" value={c.subject}/>
                <Fact label="Issuer" value={c.issuer === c.subject ? 'self-signed' : commonName(c.issuer)}/>
                <Fact label="Valid from" value={formatDateTime(c.notBefore)}/>
                <Fact label="Valid until" value={<ExpiryText date={c.notAfter} now={now}/>}/>
                <Fact label="SHA-1" value={c.sha1} copyText={c.sha1} monospace/>
                <Fact label="SHA-256" value={c.sha256} copyText={c.sha256} monospace/>
            </Box>)}
        </Section>;

// SizeBars shows where the size of a bundle goes: one bar per part, longest first
const SizeBars: React.FC<{ contents: Contents }> = ({contents}) => {
    const groups = contents.groups ?? [];
    const max = Math.max(1, ...groups.map(g => g.size));
    return <Box component="ul" aria-label="Size by part" sx={{listStyle: 'none', p: 0, m: 0}}>
        {groups.map(g => {
            const share = contents.size > 0 ? g.size / contents.size : 0;
            return <Tooltip key={g.path} placement="top-start" title={<>
                <div>{g.path}</div>
                <div>{byteFormat(g.size)} installed · {byteFormat(g.compressed)} in the file</div>
                <div>{g.files} {g.files === 1 ? 'file' : 'files'} · {(share * 100).toFixed(1)} %</div>
            </>}>
                <Box component="li" sx={{display: 'grid', gridTemplateColumns: {xs: '1fr 80px', sm: 'minmax(120px, 280px) 1fr 80px'}, gap: 1.5,
                    alignItems: 'center', py: 0.5, '&:hover': {bgcolor: 'action.hover'}, borderRadius: 1, px: 0.5}}>
                    <Typography variant="body2" sx={{...mono, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', wordBreak: 'normal'}}>{g.path}</Typography>
                    <Box sx={{display: {xs: 'none', sm: 'block'}, height: 10}}>
                        <Box sx={{height: '100%', width: `${Math.max(0.5, g.size / max * 100)}%`, bgcolor: 'primary.main', borderRadius: '0 4px 4px 0'}}/>
                    </Box>
                    <Typography variant="body2" color="text.secondary" align="right" sx={{whiteSpace: 'nowrap'}}>{byteFormat(g.size)}</Typography>
                </Box>
            </Tooltip>;
        })}
    </Box>;
};

const ContentsSection: React.FC<{ contents: Contents }> = ({contents}) =>
    <Section title="Contents"
             subtitle={`${contents.files.toLocaleString()} files · ${byteFormat(contents.size)} installed · ${byteFormat(contents.compressed)} compressed`}>
        <Grid container spacing={3}>
            <Grid size={{xs: 12, lg: 6}}>
                <Typography variant="subtitle2" sx={{mb: 0.5}}>Size by part</Typography>
                <SizeBars contents={contents}/>
            </Grid>
            <Grid size={{xs: 12, lg: 6}}>
                <Typography variant="subtitle2" sx={{mb: 0.5}}>Largest files</Typography>
                <TableContainer>
                    <Table size="small" aria-label="Largest files">
                        <TableHead>
                            <TableRow>
                                <TableCell sx={{pl: 0}}>File</TableCell>
                                <TableCell align="right">Installed</TableCell>
                                <TableCell align="right" sx={{display: {xs: 'none', sm: 'table-cell'}}}>In the file</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {(contents.largest ?? []).map(f => <TableRow key={f.path}>
                                <TableCell sx={{pl: 0, ...mono}}>{f.path}</TableCell>
                                <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>{byteFormat(f.size)}</TableCell>
                                <TableCell align="right" sx={{whiteSpace: 'nowrap', display: {xs: 'none', sm: 'table-cell'}}}>{byteFormat(f.compressed)}</TableCell>
                            </TableRow>)}
                        </TableBody>
                    </Table>
                </TableContainer>
            </Grid>
        </Grid>
    </Section>;

const DevicesSection: React.FC<{ info: BundleInfo, devices: IDeviceData[] | null, base: string }> = ({info, devices, base}) =>
    <Section title="Devices" subtitle="Which devices of the pool can install this bundle">
        {devices === null ? <LinearProgress/>
            : devices.length === 0 ? <Typography variant="body2" color="text.secondary">The pool has no device for this platform.</Typography>
                : <Stack spacing={0.5}>
                    {devices.map(d => {
                        const fit = deviceFit(info, d);
                        return <Stack key={d.id} direction="row" spacing={1} alignItems="flex-start">
                            {fit.ok ? <CheckCircleRounded color="success" fontSize="small" aria-label="can install"/>
                                : <ErrorOutlineRounded color="error" fontSize="small" aria-label="can't install"/>}
                            <Box sx={{minWidth: 0}}>
                                <Link component={RouterLink} to={`${base}/device/${d.id}`} variant="body2">{d.alias || d.name}</Link>
                                <Typography component="span" variant="body2" color="text.secondary"> · {d.os} {d.osVersion}</Typography>
                                {!fit.ok && <Typography variant="caption" color="error" sx={{display: 'block'}}>{fit.reasons.join('; ')}</Typography>}
                            </Box>
                        </Stack>;
                    })}
                </Stack>}
    </Section>;

const RunsSection: React.FC<{ detail: IAppBundleDetail, appBase: string, now: number }> = ({detail, appBase, now}) =>
    <Section title={`Test runs (${detail.runCount})`}
             subtitle={detail.runCount > detail.runs.length ? `The latest ${detail.runs.length}` : undefined}>
        {detail.runs.length === 0
            ? <Typography variant="body2" color="text.secondary">No test ran with this bundle yet.</Typography>
            : <TableContainer>
                <Table size="small" aria-label="Test runs with this bundle">
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{pl: 0}}>Run</TableCell>
                            <TableCell>Test</TableCell>
                            <TableCell>Result</TableCell>
                            <TableCell sx={{display: {xs: 'none', sm: 'table-cell'}}}>Started</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {detail.runs.map(r => {
                            const s = STATUS[r.status as RunSummaryStatus] ?? {label: r.status, tone: 'default'};
                            return <TableRow key={r.id} hover>
                                <TableCell sx={{pl: 0}}>
                                    <Link component={RouterLink} to={`${appBase}/test/${r.testId}/run/${r.id}`}>#{r.id}</Link>
                                </TableCell>
                                <TableCell>{r.testName || `Test ${r.testId}`}</TableCell>
                                <TableCell>
                                    <Chip size="small" color={s.tone} variant={s.tone === 'default' ? 'outlined' : 'filled'} label={s.label}/>
                                    {(r.passed + r.failed + r.unstable) > 0 && <Typography component="span" variant="caption" color="text.secondary" sx={{ml: 1}}>
                                        {r.passed} passed{r.failed ? `, ${r.failed} failed` : ''}{r.unstable ? `, ${r.unstable} unstable` : ''}
                                    </Typography>}
                                </TableCell>
                                <TableCell sx={{display: {xs: 'none', sm: 'table-cell'}, whiteSpace: 'nowrap'}}>
                                    <Tooltip title={formatDateTime(r.createdAt)}><span>{ago(r.createdAt, now)}</span></Tooltip>
                                </TableCell>
                            </TableRow>;
                        })}
                    </TableBody>
                </Table>
            </TableContainer>}
    </Section>;

const BundleIcon: React.FC<{ src?: string, platform?: number }> = ({src, platform}) => {
    const [failed, setFailed] = useState(false);
    if (src && !failed) {
        return <Avatar variant="rounded" src={src} alt="App icon" imgProps={{onError: () => setFailed(true)}}
                       sx={{width: 64, height: 64, border: 1, borderColor: 'divider', borderRadius: 3}}/>;
    }
    return <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', width: 64, height: 64, borderRadius: 3, border: 1, borderColor: 'divider'}}>
        {platform !== undefined && <PlatformTypeIcon platformType={platform}/>}
    </Box>;
};

// AppBundlePage shows everything known about one bundle: what the file says
// about the app, how it is signed, where its size goes, which devices can
// install it and which test runs used it.
const AppBundlePage: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const {setError} = useError();
    const navigate = useNavigate();
    const {bundleId} = useParams<{ bundleId: string }>();
    const id = Number(bundleId);
    const app = state.apps?.find(a => a.id === appId);
    const base = `/project/${projectIdentifier}`;
    const appBase = `${base}/app:${appId}`;

    const [detail, setDetail] = useState<IAppBundleDetail | null>(null);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [devices, setDevices] = useState<IDeviceData[] | null>(null);
    const [reading, setReading] = useState(false);
    const [deleting, setDeleting] = useState(false);
    const [busy, setBusy] = useState(false);
    const now = Date.now();

    const load = useCallback(() => {
        if (appId === null || isNaN(id)) return;
        setLoadError(null);
        getAppBundle(projectIdentifier, appId, id).then(setDetail).catch(e => setLoadError(messageOf(e)));
    }, [projectIdentifier, appId, id]);
    useEffect(load, [load]);

    useEffect(() => {
        const platforms = platformsOf(app?.platform);
        if (platforms.length === 0) {
            setDevices([]);
            return;
        }
        getAllDevices(projectIdentifier)
            .then(list => setDevices(list.filter(d => platforms.includes(d.platformType))))
            .catch(() => setDevices([]));
    }, [projectIdentifier, app?.platform]);

    const readAgain = () => {
        if (appId === null) return;
        setReading(true);
        readAppBundleAgain(projectIdentifier, appId, id).then(setDetail)
            .catch(e => setError(messageOf(e))).finally(() => setReading(false));
    };

    const onDelete = () => {
        if (appId === null) return;
        setBusy(true);
        deleteAppBundle(projectIdentifier, appId, id).then(() => navigate(`${appBase}/bundles`))
            .catch(e => setError(messageOf(e))).finally(() => setBusy(false));
    };

    const info = detail?.info;
    const iconUrl = useMemo(() => info?.icon && appId !== null ? bundleIconUrl(projectIdentifier, appId, id) : undefined,
        [info?.icon, projectIdentifier, appId, id]);

    if (loadError) {
        return <Box sx={{mt: 2}}>
            <Alert severity="error" action={<Button color="inherit" size="small" onClick={load}>Retry</Button>}>{loadError}</Alert>
            <Button sx={{mt: 2}} startIcon={<ArrowBackRounded/>} component={RouterLink} to={`${appBase}/bundles`}>All bundles</Button>
        </Box>;
    }
    if (!detail) return <LinearProgress sx={{mt: 2}}/>;

    const tags = parseTags(detail.tags);
    const title = info?.displayName || detail.name || detail.identifier || `Bundle #${detail.id}`;
    const version = info?.version || detail.version;
    const build = info?.build && info.build !== version ? info.build : undefined;
    const android = info?.android;
    const apple = info?.apple;
    const certificates = info?.signing?.certificates ?? [];

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Stack direction="row" spacing={2} alignItems="center" sx={{minWidth: 0}}>
                    <BundleIcon src={iconUrl} platform={app?.platform}/>
                    <Box sx={{minWidth: 0}}>
                        <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{title}</Typography>
                        <Typography variant="body2" color="text.secondary">
                            Version {version || '–'}{build && ` (build ${build})`} · #{detail.id}
                        </Typography>
                        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap sx={{mt: 0.5}}>
                            {info?.format && <Chip size="small" variant="outlined" label={`.${info.format}`}/>}
                            {info?.engine && <Chip size="small" variant="outlined"
                                                   label={[info.engine.name, info.engine.version, info.engine.backend].filter(Boolean).join(' · ')}/>}
                            {android?.debuggable && <Chip size="small" color="warning" label="debuggable"/>}
                            {certificates.some(c => c.debug) && <Chip size="small" color="warning" label="debug signed"/>}
                            {apple?.profile && <Chip size="small" variant="outlined" label={PROFILE_KINDS[apple.profile.kind] ?? apple.profile.kind}/>}
                            {tags.map(t => <Chip key={t} size="small" label={t}/>)}
                        </Stack>
                    </Box>
                </Stack>
                <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{ml: {md: 'auto'}}}>
                    <Button startIcon={<ArrowBackRounded/>} component={RouterLink} to={`${appBase}/bundles`}>All bundles</Button>
                    <Button variant="outlined" startIcon={<RefreshRounded/>} disabled={reading || !detail.appPath} onClick={readAgain}>
                        {reading ? 'Reading…' : 'Read again'}
                    </Button>
                    <Button variant="contained" startIcon={<DownloadRounded/>} disabled={!detail.appPath}
                            href={`/upload/${encodeURIComponent(detail.appPath)}`} download>Download</Button>
                    <Tooltip title="Delete bundle">
                        <IconButton color="error" aria-label="Delete bundle" onClick={() => setDeleting(true)}><DeleteOutlineRounded/></IconButton>
                    </Tooltip>
                </Stack>
            </Stack>

            {reading && <LinearProgress sx={{mb: 2}}/>}
            {!info && <Alert severity="info" sx={{mb: 2}} action={detail.appPath &&
                <Button color="inherit" size="small" disabled={reading} onClick={readAgain}>Read now</Button>}>
                This bundle was uploaded before the master read all details from the file. Read it again to see them.
            </Alert>}
            {info?.warnings?.map(w => <Alert key={w} severity="warning" sx={{mb: 2}}>{w}</Alert>)}

            <Grid container spacing={2}>
                <Grid size={{xs: 12, md: 6}}>
                    <Section title="File">
                        <Fact label="File name" value={fileName(detail.appPath, detail.hash)} monospace/>
                        <Fact label="Size" value={detail.size ? byteFormat(detail.size) : undefined}/>
                        <Fact label="Installed size" value={info?.contents ? byteFormat(info.contents.size) : undefined}/>
                        <Fact label="SHA-1" value={detail.hash} copyText={detail.hash} monospace/>
                        <Fact label="Uploaded" value={`${formatDateTime(detail.createdAt)} (${ago(detail.createdAt, now)})`}/>
                        <Fact label="Details read" value={info?.readAt ? formatDateTime(info.readAt) : undefined}/>
                        <Fact label="Platform" value={detail.platform}/>
                        <Fact label="Identifier" value={detail.identifier} copyText={detail.identifier} monospace/>
                        {!android && detail.android?.LaunchActivity &&
                            <Fact label="Launch activity" value={detail.android.LaunchActivity} monospace/>}
                        {!apple && detail.executable?.Executable && <Fact label="Executable" value={detail.executable.Executable} monospace/>}
                        <Fact label="Engine" value={info?.engine && [info.engine.name, info.engine.version, info.engine.backend].filter(Boolean).join(' · ')}/>
                    </Section>
                </Grid>
                {info && <Grid size={{xs: 12, md: 6}}>
                    <DevicesSection info={info} devices={devices} base={base}/>
                </Grid>}

                {android && <>
                    <Grid size={{xs: 12, md: 6}}><AndroidSection android={android}/></Grid>
                    <Grid size={{xs: 12, md: 6}}><PermissionsSection android={android}/></Grid>
                </>}
                {apple && info && <>
                    <Grid size={{xs: 12, md: 6}}><AppleSection apple={apple} info={info}/></Grid>
                    <Grid size={{xs: 12, md: 6}}><UsageSection apple={apple}/></Grid>
                    {apple.profile && <Grid size={{xs: 12, md: 6}}><ProfileSection profile={apple.profile} devices={devices ?? []} now={now}/></Grid>}
                </>}
                {info && (info.signing || android) && <Grid size={{xs: 12, md: 6}}>
                    <CertificatesSection certificates={certificates} schemes={info.signing?.schemes} apple={!!apple} now={now}/>
                </Grid>}

                {info?.contents && <Grid size={12}><ContentsSection contents={info.contents}/></Grid>}
                <Grid size={12}><RunsSection detail={detail} appBase={appBase} now={now}/></Grid>
            </Grid>

            <Dialog open={deleting} onClose={() => setDeleting(false)} aria-labelledby="delete-bundle-title">
                <DialogTitle id="delete-bundle-title">Delete version {version || detail.name}?</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        The file{detail.size ? ` (${byteFormat(detail.size)})` : ''} is removed from the master.
                        {detail.runCount > 0
                            ? ` ${detail.runCount} test ${detail.runCount === 1 ? 'run used' : 'runs used'} it; their results stay, but they can't be run again with this bundle.`
                            : ' No test run used it.'}
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setDeleting(false)}>Keep it</Button>
                    <Button color="error" variant="contained" disabled={busy} onClick={onDelete}>Delete bundle</Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default AppBundlePage;
