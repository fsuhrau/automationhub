import React, {useEffect, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    ButtonBase,
    Card,
    CardContent,
    Chip,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    IconButton,
    Link,
    ListItemIcon,
    ListItemText,
    Menu,
    MenuItem,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableRow,
    Tooltip,
    Typography,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import {
    BrokenImageRounded,
    CancelRounded,
    CheckCircleRounded,
    ContentCopyRounded,
    DeleteOutlineRounded,
    EditRounded,
    ErrorRounded,
    HourglassEmptyRounded,
    LockOpenRounded,
    MoreVertRounded,
    RefreshRounded,
    TouchAppRounded,
} from '@mui/icons-material';
import {Link as RouterLink, useNavigate} from 'react-router-dom';
import IDeviceData from '../../types/device';
import ITestProtocolData from '../../types/test.protocol';
import {deleteDevice, getDeviceProtocols, postUnlockDevice} from '../../services/device.service';
import {screenshotUrl} from '../../services/device.control.service';
import {DeviceType} from '../../types/device.type.enum';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {getPlatformName} from '../../types/platform.type.enum';
import {TestResultState} from '../../types/test.result.state.enum';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {useError} from '../../ErrorProvider';
import {HubStateActions} from '../../application/HubState';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import DeviceStatus from '../../components/DeviceStatus';
import {describeController} from '../../components/DeviceController';
import {formatDateTime} from '../../components/DateTime';
import {between, formatDuration, splitTestName} from '../tests/testRunModel';
import {availability} from './deviceDetailsModel';
import {useAuth} from '../../hooks/AuthProvider';

interface DeviceShowPageProps {
    device: IDeviceData
}

const DEVICE_TYPES: Record<DeviceType, string> = {
    [DeviceType.Phone]: 'Phone',
    [DeviceType.Tablet]: 'Tablet',
    [DeviceType.Desktop]: 'Desktop',
    [DeviceType.UnityEditor]: 'Unity editor',
    [DeviceType.Browser]: 'Browser',
};

const CONNECTIONS: Record<DeviceConnectionType, string> = {
    [DeviceConnectionType.USB]: 'USB',
    [DeviceConnectionType.Remote]: 'Network',
    [DeviceConnectionType.HubNode]: 'Through a node',
};

const ResultIcon: React.FC<{ result: TestResultState }> = ({result}) => {
    switch (result) {
        case TestResultState.TestResultFailed:
            return <Tooltip title="Failed"><CancelRounded color="error" fontSize="small"/></Tooltip>;
        case TestResultState.TestResultUnstable:
            return <Tooltip title="Unstable"><ErrorRounded color="warning" fontSize="small"/></Tooltip>;
        case TestResultState.TestResultSuccess:
            return <Tooltip title="Passed"><CheckCircleRounded color="success" fontSize="small"/></Tooltip>;
        default:
            return <Tooltip title="Running or not finished"><HourglassEmptyRounded color="disabled" fontSize="small"/></Tooltip>;
    }
};

const Section: React.FC<{ title: string, action?: React.ReactNode, children: React.ReactNode }> = ({title, action, children}) =>
    <Card variant="outlined">
        <CardContent>
            <Stack direction="row" alignItems="center" spacing={1} useFlexGap sx={{mb: 1.5, minHeight: 32}}>
                <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
                {action && <Box sx={{ml: 'auto'}}>{action}</Box>}
            </Stack>
            {children}
        </CardContent>
    </Card>;

const KeyValues: React.FC<{ rows: [string, React.ReactNode][] }> = ({rows}) =>
    <Box sx={{display: 'grid', gridTemplateColumns: 'minmax(110px, max-content) 1fr', columnGap: 2, rowGap: 0.75}}>
        {rows.map(([k, v]) => <React.Fragment key={k}>
            <Typography variant="body2" color="text.secondary">{k}</Typography>
            <Typography variant="body2" component="div" sx={{wordBreak: 'break-word', minWidth: 0}}>{v}</Typography>
        </React.Fragment>)}
    </Box>;

const copy = (text: string) => navigator.clipboard.writeText(text).catch(() => undefined);

// DeviceShowPage answers whether the device can be used now, what it shows,
// what ran on it lately, and how it is set up.
const DeviceShowPage: React.FC<DeviceShowPageProps> = (props: DeviceShowPageProps) => {
    const {projectIdentifier} = useProjectContext();
    const {state, dispatch} = useHubState();
    const navigate = useNavigate();
    const {setError} = useError();

    // the device list follows live updates; the loaded device has the details
    const live = state.devices?.find(d => d.id === props.device.id);
    const device: IDeviceData = live ? {...props.device, status: live.status, isLocked: live.isLocked, lockedBy: live.lockedBy} : props.device;
    const node = state.nodes?.find(n => n.id === device.nodeId) ?? device.node;
    const {user} = useAuth();
    const avail = availability(device, owner => describeController(owner).label, user?.name);
    const name = device.alias || device.name;
    const base = `/project/${projectIdentifier}`;

    const [menu, setMenu] = useState<HTMLElement | null>(null);
    const [confirm, setConfirm] = useState<'delete' | 'unlock' | null>(null);
    const [busy, setBusy] = useState(false);

    const [shot, setShot] = useState(() => screenshotUrl(projectIdentifier, device.id));
    const [shotBroken, setShotBroken] = useState(false);
    const refreshShot = () => {
        setShotBroken(false);
        setShot(screenshotUrl(projectIdentifier, device.id));
    };

    const [recent, setRecent] = useState<ITestProtocolData[] | null>(null);
    const [recentError, setRecentError] = useState(false);
    useEffect(() => {
        getDeviceProtocols(projectIdentifier, device.id, 10).then(setRecent).catch(() => setRecentError(true));
    }, [projectIdentifier, device.id]);

    const onDelete = () => {
        setBusy(true);
        deleteDevice(projectIdentifier, device.id).then(() => navigate(`${base}/devices`))
            .catch(e => setError(e)).finally(() => {
            setBusy(false);
            setConfirm(null);
        });
    };

    const onUnlock = () => {
        setBusy(true);
        postUnlockDevice(projectIdentifier, device.id).then(() => {
            dispatch({type: HubStateActions.UnlockDevice, payload: {deviceId: device.id}});
        }).catch(e => setError(e)).finally(() => {
            setBusy(false);
            setConfirm(null);
        });
    };

    const deviceParameter = device.deviceParameter ?? [];
    const customParameter = device.customParameter ?? [];
    const connection = device.connectionParameter;

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            {/* which device, whether it can be used, and what to do with it */}
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Stack direction="row" spacing={1.5} alignItems="center" sx={{minWidth: 0}}>
                    <Box sx={{display: 'flex', p: 1, borderRadius: 1, border: 1, borderColor: 'divider'}}>
                        <PlatformTypeIcon platformType={device.platformType}/>
                    </Box>
                    <Box sx={{minWidth: 0}}>
                        <Stack direction="row" spacing={1.5} alignItems="center" flexWrap="wrap" useFlexGap>
                            <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{name}</Typography>
                            <Tooltip title={avail.detail}><Chip size="small" color={avail.tone} label={avail.label}/></Tooltip>
                        </Stack>
                        <Typography variant="body2" color="text.secondary">
                            {DEVICE_TYPES[device.deviceType] ?? 'Device'} · {device.os} {device.osVersion}
                            {node && <> · on {node.name}</>}
                        </Typography>
                    </Box>
                </Stack>
                <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {md: 'auto'}}}>
                    <Tooltip title={avail.canControl ? 'See the screen and use the device' : `Not possible now: ${avail.label.toLowerCase()}`}>
                        <span><Button variant="contained" startIcon={<TouchAppRounded/>} disabled={!avail.canControl}
                                      component={RouterLink} to="control">Live control</Button></span>
                    </Tooltip>
                    <Button variant="outlined" startIcon={<EditRounded/>} component={RouterLink} to="edit">Edit</Button>
                    <IconButton aria-label="More actions" onClick={e => setMenu(e.currentTarget)}><MoreVertRounded/></IconButton>
                    <Menu anchorEl={menu} open={!!menu} onClose={() => setMenu(null)}>
                        <MenuItem disabled={!avail.canUnlock} onClick={() => {
                            setMenu(null);
                            setConfirm('unlock');
                        }}>
                            <ListItemIcon><LockOpenRounded fontSize="small"/></ListItemIcon>
                            <ListItemText primary="Unlock" secondary="End the test run or control session holding it"/>
                        </MenuItem>
                        <MenuItem onClick={() => {
                            setMenu(null);
                            setConfirm('delete');
                        }} sx={{color: 'error.main'}}>
                            <ListItemIcon><DeleteOutlineRounded fontSize="small" color="error"/></ListItemIcon>
                            <ListItemText primary="Delete device"/>
                        </MenuItem>
                    </Menu>
                </Stack>
            </Stack>

            {avail.label === 'Controlled by you' && <Alert severity="info" sx={{mb: 2}}
                                                          action={<Button color="inherit" size="small" component={RouterLink} to="control">Live control</Button>}>
                {avail.detail}
            </Alert>}
            {avail.tone !== 'success' && <Alert severity={avail.tone === 'error' ? 'error' : avail.tone === 'warning' ? 'warning' : 'info'}
                                                sx={{mb: 2}}
                                                action={avail.canUnlock && <Button color="inherit" size="small" onClick={() => setConfirm('unlock')}>Unlock</Button>}>
                <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                    <span>{avail.detail}</span>
                    {!avail.canUnlock && <DeviceStatus state={device.status} platform={device.platformType} label="How to fix"/>}
                </Stack>
            </Alert>}

            <Grid container spacing={2}>
                <Grid size={{xs: 12, md: 5, lg: 4}}>
                    <Section title="Screen" action={avail.canControl && <Tooltip title="Take a new screenshot">
                        <IconButton size="small" aria-label="Take a new screenshot" onClick={refreshShot}><RefreshRounded fontSize="small"/></IconButton>
                    </Tooltip>}>
                        {avail.canControl && !shotBroken
                            ? <ButtonBase component={RouterLink} to="control" aria-label="Open live control"
                                          sx={{display: 'block', borderRadius: 1, overflow: 'hidden', width: '100%'}}>
                                <Box component="img" src={shot} alt={`Screen of ${name}`} onError={() => setShotBroken(true)}
                                     // a screenshot takes a moment; keep the space meanwhile
                                     sx={{display: 'block', width: '100%', minHeight: 200, maxHeight: 440, objectFit: 'contain', bgcolor: 'action.hover'}}/>
                            </ButtonBase>
                            : <Stack alignItems="center" justifyContent="center" spacing={1}
                                     sx={{height: 200, bgcolor: 'action.hover', borderRadius: 1, color: 'text.secondary', p: 2, textAlign: 'center'}}>
                                <BrokenImageRounded/>
                                <Typography variant="body2">
                                    {avail.canControl ? 'The device returned no screenshot.' : `No screen while the device is ${avail.label.toLowerCase()}.`}
                                </Typography>
                            </Stack>}
                        {avail.canControl && <Typography variant="caption" color="text.secondary">Click it for live control.</Typography>}
                    </Section>
                </Grid>

                <Grid size={{xs: 12, md: 7, lg: 8}}>
                    <Stack spacing={2}>
                        <Section title="Recent tests">
                            {recentError
                                ? <Typography variant="body2" color="text.secondary">The test history couldn't be loaded.</Typography>
                                : recent === null
                                    ? <Typography variant="body2" color="text.secondary">Loading…</Typography>
                                    : recent.length === 0
                                        ? <Typography variant="body2" color="text.secondary">No test of this project ran on this device yet.</Typography>
                                        : <Table size="small" aria-label="Recent tests on this device">
                                            <TableBody>
                                                {recent.map(p => {
                                                    const {name: testName, suite} = splitTestName(p.testName);
                                                    const test = p.testRun?.test;
                                                    const ms = between(p.startedAt, p.endedAt);
                                                    const url = test ? `${base}/app:${test.appId}/test/${p.testRun.testId}/run/${p.testRunId}/${p.id}` : null;
                                                    return <TableRow key={p.id as number} hover={!!url} sx={{cursor: url ? 'pointer' : undefined}}
                                                                     onClick={() => url && navigate(url)}>
                                                        <TableCell sx={{width: 36, pl: 0}}><ResultIcon result={p.testResult}/></TableCell>
                                                        <TableCell>
                                                            <Typography variant="body2" sx={{fontWeight: 600, wordBreak: 'break-word'}}>{testName}</Typography>
                                                            <Typography variant="caption" color="text.secondary">
                                                                {test ? `${test.name} · run ${p.testRunId}` : suite}
                                                            </Typography>
                                                        </TableCell>
                                                        <TableCell align="right" sx={{whiteSpace: 'nowrap', pr: 0}}>
                                                            <Typography variant="body2">{formatDateTime(p.startedAt, 'YYYY/MM/DD HH:mm')}</Typography>
                                                            <Typography variant="caption" color="text.secondary">{ms === null ? 'not finished' : formatDuration(ms)}</Typography>
                                                        </TableCell>
                                                    </TableRow>;
                                                })}
                                            </TableBody>
                                        </Table>}
                        </Section>

                        <Grid container spacing={2}>
                            <Grid size={{xs: 12, lg: 6}}>
                                <Section title="Device">
                                    <KeyValues rows={[
                                        ['Name', device.name],
                                        ...(device.alias ? [['Alias', device.alias] as [string, React.ReactNode]] : []),
                                        ['Identifier', <Stack direction="row" alignItems="center" spacing={0.5} component="span">
                                            <Box component="span" sx={{fontFamily: 'monospace', wordBreak: 'break-all'}}>{device.deviceIdentifier}</Box>
                                            <Tooltip title="Copy"><IconButton size="small" aria-label="Copy identifier"
                                                                              onClick={() => copy(device.deviceIdentifier)}><ContentCopyRounded fontSize="inherit"/></IconButton></Tooltip>
                                        </Stack>],
                                        ['Platform', getPlatformName(device.platformType)],
                                        ['System', `${device.os} ${device.osVersion}`],
                                        ['Type', DEVICE_TYPES[device.deviceType] ?? '–'],
                                        ['Status', <DeviceStatus state={device.status} platform={device.platformType}/>],
                                        ['Hub id', String(device.id)],
                                        ...(device.createdAt ? [['Added', formatDateTime(device.createdAt, 'YYYY/MM/DD')] as [string, React.ReactNode]] : []),
                                    ]}/>
                                </Section>
                            </Grid>
                            <Grid size={{xs: 12, lg: 6}}>
                                <Section title="Connection">
                                    <KeyValues rows={[
                                        // devices found by a node have no connection settings of their own
                                        ...(connection ? [['Connected', CONNECTIONS[connection.connectionType] ?? '–'] as [string, React.ReactNode]] : []),
                                        ['Node', node
                                            ? <>{node.name} <Typography component="span" variant="caption" color="text.secondary">
                                                ({node.status === 1 ? 'connected' : 'offline'})
                                            </Typography></>
                                            : 'the master itself'],
                                        ...(connection?.connectionType === DeviceConnectionType.Remote && connection.ip
                                            ? [['Address', `${connection.ip}${connection.port ? `:${connection.port}` : ''}`] as [string, React.ReactNode]]
                                            : []),
                                        ...(device.connection?.version ? [['App connected', `version ${device.connection.version}`] as [string, React.ReactNode]] : []),
                                    ]}/>
                                </Section>
                            </Grid>
                        </Grid>

                        <Grid container spacing={2}>
                            <Grid size={{xs: 12, lg: 6}}>
                                <Section title="Reported by the device">
                                    {deviceParameter.length === 0
                                        ? <Typography variant="body2" color="text.secondary">The device reported no further details.</Typography>
                                        : <KeyValues rows={deviceParameter.map(p => [p.key, p.value] as [string, React.ReactNode])}/>}
                                </Section>
                            </Grid>
                            <Grid size={{xs: 12, lg: 6}}>
                                <Section title="Custom parameters" action={<Button size="small" component={RouterLink} to="edit">Edit</Button>}>
                                    {customParameter.length === 0
                                        ? <Typography variant="body2" color="text.secondary">
                                            None. Custom parameters are passed to tests on this device; add them under <Link component={RouterLink} to="edit">Edit</Link>.
                                        </Typography>
                                        : <KeyValues rows={customParameter.map(p => [p.key, p.value] as [string, React.ReactNode])}/>}
                                </Section>
                            </Grid>
                        </Grid>
                    </Stack>
                </Grid>
            </Grid>

            <Dialog open={confirm === 'delete'} onClose={() => setConfirm(null)} aria-labelledby="delete-device-title">
                <DialogTitle id="delete-device-title">Delete {name}?</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        The device leaves the device pool, with its alias and custom parameters. Its test results stay.
                        A device that is still connected can be detected again.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setConfirm(null)}>Keep it</Button>
                    <Button color="error" variant="contained" onClick={onDelete} disabled={busy}>Delete device</Button>
                </DialogActions>
            </Dialog>

            <Dialog open={confirm === 'unlock'} onClose={() => setConfirm(null)} aria-labelledby="unlock-device-title">
                <DialogTitle id="unlock-device-title">Unlock {name}?</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        {device.lockedBy
                            ? `This ends the live control of ${describeController(device.lockedBy).label}.`
                            : 'A test run still on this device loses it and fails. Unlock only when the run is stuck.'}
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setConfirm(null)}>Cancel</Button>
                    <Button color="warning" variant="contained" onClick={onUnlock} disabled={busy}>Unlock</Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default DeviceShowPage;
