import React, {useCallback, useEffect, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    ButtonBase,
    Card,
    CardContent,
    Chip,
    CircularProgress,
    LinearProgress,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableRow,
    Tooltip,
    Typography,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import {AddRounded, CheckCircleRounded} from '@mui/icons-material';
import {Link as RouterLink, useNavigate} from 'react-router-dom';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {getOverview, IOverviewRun, IProjectOverview, OverviewMissing} from '../../services/overview.service';
import {formatDateTime} from '../../components/DateTime';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import {byteFormat} from '../tests/value_formatter';
import {ago} from '../apps/bundlesModel';
import {STATUS} from '../tests/testsListModel';
import {availability} from '../devices/deviceDetailsModel';
import {describeController} from '../../components/DeviceController';

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};

const Section: React.FC<{ title: string, action?: React.ReactNode, children: React.ReactNode }> = ({title, action, children}) =>
    <Card variant="outlined" sx={{height: '100%'}}>
        <CardContent>
            <Stack direction="row" alignItems="center" spacing={1} useFlexGap sx={{mb: 1.5, minHeight: 32}}>
                <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
                {action && <Box sx={{ml: 'auto'}}>{action}</Box>}
            </Stack>
            {children}
        </CardContent>
    </Card>;

// a number with what it counts, leading to where it's managed
const Tile: React.FC<{ label: string, value: React.ReactNode, note?: React.ReactNode, to?: string }> = ({label, value, note, to}) => {
    const content = <Box sx={{p: 1.5, textAlign: 'left', width: '100%'}}>
        <Typography variant="caption" color="text.secondary">{label}</Typography>
        <Typography variant="h5" component="div">{value}</Typography>
        {note && <Typography variant="caption" color="text.secondary" component="div">{note}</Typography>}
    </Box>;
    return <Card variant="outlined" sx={{height: '100%'}}>
        {to ? <ButtonBase component={RouterLink} to={to} sx={{display: 'block', height: '100%', width: '100%'}}>{content}</ButtonBase> : content}
    </Card>;
};

const RunRows: React.FC<{ runs: IOverviewRun[], runUrl: (r: IOverviewRun) => string, appName: (id: number) => React.ReactNode, now: number }> =
    ({runs, runUrl, appName, now}) => {
        const navigate = useNavigate();
        return <Table size="small">
            <TableBody>
                {runs.map(r => {
                    const s = STATUS[r.status];
                    const counts = [r.passed && `${r.passed} passed`, r.failed && `${r.failed} failed`, r.unstable && `${r.unstable} unstable`, r.open && `${r.open} open`]
                        .filter(Boolean).join(' · ');
                    return <TableRow key={r.id} hover sx={{cursor: 'pointer', '&:last-child td': {borderBottom: 0}}} onClick={() => navigate(runUrl(r))}>
                        <TableCell sx={{pl: 0, width: 120}}>
                            <Chip size="small" color={s.tone} label={s.label}
                                  icon={r.status === 'running' ? <CircularProgress size={12} color="inherit"/> : undefined}/>
                        </TableCell>
                        <TableCell>
                            <Typography variant="body2" sx={{fontWeight: 600}}>{r.testName}</Typography>
                            <Typography variant="caption" color="text.secondary">{appName(r.appId)} · run {r.id}{counts && ` · ${counts}`}</Typography>
                        </TableCell>
                        <TableCell align="right" sx={{pr: 0, whiteSpace: 'nowrap'}}>
                            <Tooltip title={formatDateTime(r.createdAt)}><Typography variant="body2" color="text.secondary">{ago(r.createdAt, now)}</Typography></Tooltip>
                        </TableCell>
                    </TableRow>;
                })}
            </TableBody>
        </Table>;
    };

// ProjectDashboard is the project's home: what needs attention, what runs,
// what ran, and whether devices are ready
const ProjectDashboard: React.FC = () => {
    const {project, projectIdentifier} = useProjectContext();
    const {state} = useHubState();
    const [overview, setOverview] = useState<IProjectOverview | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [outdated, setOutdated] = useState(false);
    const now = Date.now();
    const base = `/project/${projectIdentifier}`;

    const load = useCallback(() => {
        setError(null);
        setOutdated(false);
        getOverview(projectIdentifier).then(setOverview).catch(e => e instanceof OverviewMissing ? setOutdated(true) : setError(messageOf(e)));
    }, [projectIdentifier]);
    useEffect(load, [load]);
    // runs change while the page is open; refresh while something runs
    useEffect(() => {
        if (!overview || overview.running.length === 0) return;
        const timer = setInterval(load, 15000);
        return () => clearInterval(timer);
    }, [overview, load]);

    const apps = state.apps ?? [];
    const appName = (id: number) => {
        const app = apps.find(a => a.id === id);
        return app ? <><PlatformTypeIcon platformType={app.platform} sx={{fontSize: 12, verticalAlign: 'middle', mr: 0.5}}/>{app.name}</> : `app ${id}`;
    };
    const runUrl = (r: IOverviewRun) => `${base}/app:${r.appId}/test/${r.testId}/run/${r.id}`;

    const devices = state.devices ?? [];
    const states = devices.map(d => availability(d, owner => describeController(owner).label));
    const ready = states.filter(a => a.label === 'Ready').length;
    const inUse = states.filter(a => a.canUnlock).length;

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Typography component="h1" variant="h5">{project.name}</Typography>
            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>How the project's tests are doing, what runs now, and whether devices are ready.</Typography>

            {error && <Alert severity="error" sx={{mb: 2}} action={<Button color="inherit" size="small" onClick={load}>Retry</Button>}>{error}</Alert>}
            {outdated && <Alert severity="info" sx={{mb: 2}}>{new OverviewMissing().message}</Alert>}
            {!overview && !error && !outdated && <LinearProgress sx={{mb: 2}}/>}

            {/* devices and apps are known without the overview, so they show even when it fails */}
            <Grid container spacing={1.5} sx={{mb: 2}}>
                <Grid size={{xs: 6, md: 3}}>
                    <Tile label="Devices ready" value={devices.length ? `${ready} of ${devices.length}` : '0'}
                          note={inUse ? `${inUse} in use` : devices.length ? 'none in use' : 'no devices yet'} to={`${base}/devices`}/>
                </Grid>
                <Grid size={{xs: 6, md: 3}}>
                    <Tile label="Apps" value={overview?.apps ?? apps.length} note="platforms of the game" to={`${base}/settings?tab=apps`}/>
                </Grid>
                {overview && <>
                    <Grid size={{xs: 6, md: 3}}>
                        <Tile label="Tests" value={overview.tests} note={overview.failing.length ? `${overview.failing.length} need attention` : undefined}/>
                    </Grid>
                    <Grid size={{xs: 6, md: 3}}>
                        <Tile label="Bundles" value={overview.bundles} note={`${byteFormat(overview.bundleStorage)} stored`}/>
                    </Grid>
                </>}
            </Grid>

            {overview && <>

                {overview.apps === 0
                    ? <Section title="Get started">
                        <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>
                            Add the game's apps first (one per platform). Then upload a bundle and create a test.
                        </Typography>
                        <Button variant="contained" startIcon={<AddRounded/>} component={RouterLink} to={`${base}/settings?tab=apps`}>Add an app</Button>
                    </Section>
                    : overview.tests === 0
                        ? <Section title="Get started">
                            <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>The project has no tests yet. A test says what to run and on which devices.</Typography>
                            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                                {apps.map(a => <Button key={a.id} variant="outlined" startIcon={<AddRounded/>}
                                                       component={RouterLink} to={`${base}/app:${a.id}/test/new`}>Test for {a.name}</Button>)}
                            </Stack>
                        </Section>
                        : <Grid container spacing={2}>
                            <Grid size={{xs: 12, lg: 6}}>
                                <Stack spacing={2}>
                                    <Section title="Needs attention">
                                        {overview.failing.length === 0
                                            ? <Stack direction="row" spacing={1} alignItems="center">
                                                <CheckCircleRounded color="success" fontSize="small"/>
                                                <Typography variant="body2">No test failed in its latest run.</Typography>
                                            </Stack>
                                            : <>
                                                <Typography variant="caption" color="text.secondary">Tests whose latest run failed, was unstable or didn't run.</Typography>
                                                <RunRows runs={overview.failing} runUrl={runUrl} appName={appName} now={now}/>
                                            </>}
                                    </Section>
                                    {overview.running.length > 0 && <Section title="Running now">
                                        <RunRows runs={overview.running} runUrl={runUrl} appName={appName} now={now}/>
                                    </Section>}
                                </Stack>
                            </Grid>
                            <Grid size={{xs: 12, lg: 6}}>
                                <Section title="Latest runs">
                                    {overview.latest.length === 0
                                        ? <Typography variant="body2" color="text.secondary">Nothing ran yet. Start a test from Tests.</Typography>
                                        : <RunRows runs={overview.latest} runUrl={runUrl} appName={appName} now={now}/>}
                                </Section>
                            </Grid>
                        </Grid>}
            </>}
        </Box>
    );
};

export default ProjectDashboard;
