import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    CircularProgress,
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
import Grid from '@mui/material/Grid';
import {EditRounded, PlayArrowRounded} from '@mui/icons-material';
import {Link as RouterLink, useNavigate} from 'react-router-dom';
import ITestData from '../../types/test';
import {TestType} from '../../types/test.type.enum';
import {TestExecutionType} from '../../types/test.execution.type.enum';
import {UnityTestCategory} from '../../types/unity.test.category.type.enum';
import {getRunsPage, ITestRunListEntry} from '../../services/test.run.service';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {formatDateTime} from '../../components/DateTime';
import RunTestDialog from '../../components/RunTestDialog';
import {ago} from '../apps/bundlesModel';
import {between, formatDuration, parseParameters} from './testRunModel';
import {describeConfig, runStats, STATUS} from './testsListModel';

interface TestContentProps {
    test: ITestData
}

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
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

const Stat: React.FC<{ label: string, value: React.ReactNode, note?: React.ReactNode }> = ({label, value, note}) =>
    <Box sx={{p: 1.5, border: 1, borderColor: 'divider', borderRadius: 1, height: '100%'}}>
        <Typography variant="caption" color="text.secondary">{label}</Typography>
        <Typography variant="h6" component="div">{value}</Typography>
        {note && <Typography variant="caption" color="text.secondary" component="div">{note}</Typography>}
    </Box>;

const results = (r: ITestRunListEntry): string =>
    [r.passed && `${r.passed} passed`, r.failed && `${r.failed} failed`, r.unstable && `${r.unstable} unstable`, r.open && `${r.open} open`]
        .filter(Boolean).join(' · ');

// ShowTestPage shows a test: how it has been doing, what it runs where, and
// its runs; it starts and edits the test.
const ShowTestPage: React.FC<TestContentProps> = ({test}) => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const navigate = useNavigate();
    const app = state.apps?.find(a => a.id === appId);
    const config = test.testConfig;
    const base = `/project/${projectIdentifier}/app:${appId}/test/${test.id}`;

    const [runs, setRuns] = useState<ITestRunListEntry[] | null>(null);
    const [more, setMore] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [runDialog, setRunDialog] = useState(false);
    const now = Date.now();

    const load = useCallback((before?: number) => {
        if (appId === null) return;
        setLoading(true);
        setError(null);
        getRunsPage(projectIdentifier, appId, test.id, before).then(page => {
            setRuns(prev => before && prev ? [...prev, ...page.runs] : page.runs);
            setMore(page.more);
        }).catch(e => setError(messageOf(e))).finally(() => setLoading(false));
    }, [projectIdentifier, appId, test.id]);
    useEffect(() => load(), [load]);

    const stats = useMemo(() => runStats(runs ?? []), [runs]);
    const summarized = !!runs && runs.some(r => r.status !== undefined);
    const last = runs?.[0];
    const lastStatus = last?.status ? STATUS[last.status] : null;

    const unity = config.type === TestType.Unity ? config.unity : null;
    const functionsByClass = useMemo(() => {
        const groups = new Map<string, string[]>();
        (unity?.testFunctions ?? []).forEach(f => {
            const key = f.class || f.assembly || '';
            groups.set(key, [...(groups.get(key) ?? []), f.method]);
        });
        return Array.from(groups);
    }, [unity]);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{test.name}</Typography>
                    <Typography variant="body2" color="text.secondary">{describeConfig(config)}</Typography>
                </Box>
                <Stack direction="row" spacing={1} sx={{ml: {md: 'auto'}}}>
                    <Button variant="outlined" startIcon={<EditRounded/>} component={RouterLink} to="edit">Edit</Button>
                    <Button variant="contained" startIcon={<PlayArrowRounded/>} disabled={!app} onClick={() => setRunDialog(true)}>Run</Button>
                </Stack>
            </Stack>

            <Grid container spacing={2}>
                <Grid size={{xs: 12, lg: 8}}>
                    <Stack spacing={2}>
                        {summarized && runs!.length > 0 && <Grid container spacing={1.5}>
                            <Grid size={{xs: 6, md: 3}}>
                                <Stat label="Last run" value={lastStatus?.label ?? '–'} note={last ? ago(last.createdAt, now) : undefined}/>
                            </Grid>
                            <Grid size={{xs: 6, md: 3}}>
                                <Stat label="Passed" value={stats.judged ? `${stats.passed} of ${stats.judged}` : '–'}
                                      note={`of the runs below with a result`}/>
                            </Grid>
                            <Grid size={{xs: 6, md: 3}}>
                                <Stat label="Last passed" value={stats.lastPassedAt ? ago(stats.lastPassedAt, now) : 'not yet'}
                                      note={stats.lastPassedAt ? formatDateTime(stats.lastPassedAt, 'YYYY/MM/DD HH:mm') : undefined}/>
                            </Grid>
                            <Grid size={{xs: 6, md: 3}}>
                                <Stat label="Usually takes" value={stats.typicalMs !== null ? formatDuration(stats.typicalMs) : '–'}
                                      note="when it passes"/>
                            </Grid>
                        </Grid>}

                        <Section title="Runs" action={runs && runs.length > 0 &&
                            <Button size="small" component={RouterLink} to="runs/last">Latest run</Button>}>
                            {error
                                ? <Alert severity="error" action={<Button color="inherit" size="small" onClick={() => load()}>Retry</Button>}>{error}</Alert>
                                : runs === null
                                    ? <LinearProgress/>
                                    : runs.length === 0
                                        ? <Stack alignItems="flex-start" spacing={1}>
                                            <Typography variant="body2" color="text.secondary">This test hasn't run yet.</Typography>
                                            <Button variant="outlined" startIcon={<PlayArrowRounded/>} disabled={!app} onClick={() => setRunDialog(true)}>Run it</Button>
                                        </Stack>
                                        : <>
                                            <TableContainer sx={{mx: -2, width: 'auto'}}>
                                                <Table size="small" aria-label="Runs of this test">
                                                    <TableHead>
                                                        <TableRow>
                                                            <TableCell sx={{pl: 2}}>Run</TableCell>
                                                            <TableCell>Started</TableCell>
                                                            <TableCell sx={{display: {xs: 'none', sm: 'table-cell'}}}>Took</TableCell>
                                                            <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>Results</TableCell>
                                                            <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>Bundle and parameters</TableCell>
                                                        </TableRow>
                                                    </TableHead>
                                                    <TableBody>
                                                        {runs.map(r => {
                                                            const s = r.status ? STATUS[r.status] : null;
                                                            const ms = r.status === 'running' ? null : between(r.createdAt, r.endedAt);
                                                            const params = parseParameters(r.parameter);
                                                            return <TableRow key={r.id} hover sx={{cursor: 'pointer'}} onClick={() => navigate(`${base}/run/${r.id}`)}>
                                                                <TableCell sx={{pl: 2, whiteSpace: 'nowrap'}}>
                                                                    <Stack direction="row" spacing={1} alignItems="center">
                                                                        {s && <Chip size="small" color={s.tone} label={s.label}
                                                                                    icon={r.status === 'running' ? <CircularProgress size={12} color="inherit"/> : undefined}/>}
                                                                        <Link component={RouterLink} to={`${base}/run/${r.id}`} underline="hover"
                                                                              onClick={e => e.stopPropagation()}>#{r.id}</Link>
                                                                    </Stack>
                                                                </TableCell>
                                                                <TableCell sx={{whiteSpace: 'nowrap'}}>
                                                                    <Tooltip title={formatDateTime(r.createdAt)}><span>{ago(r.createdAt, now)}</span></Tooltip>
                                                                </TableCell>
                                                                <TableCell sx={{display: {xs: 'none', sm: 'table-cell'}, whiteSpace: 'nowrap'}}>
                                                                    {r.status === 'running' ? 'running' : ms !== null ? formatDuration(ms) : '–'}
                                                                </TableCell>
                                                                <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>
                                                                    <Typography variant="body2" color="text.secondary">{results(r) || '–'}</Typography>
                                                                </TableCell>
                                                                <TableCell sx={{display: {xs: 'none', md: 'table-cell'}, maxWidth: 320}}>
                                                                    <Typography variant="body2" noWrap title={r.bundle ? `${r.bundle.version} (#${r.bundle.id})` : undefined}>
                                                                        {r.bundle ? r.bundle.version || r.bundle.name : '–'}
                                                                    </Typography>
                                                                    {params.length > 0 && <Typography variant="caption" color="text.secondary" noWrap component="div"
                                                                                                      title={params.map(p => `${p.name}=${p.value}`).join('\n')}>
                                                                        {params.map(p => p.value ? `${p.name}=${p.value}` : p.name).join(' · ')}
                                                                    </Typography>}
                                                                </TableCell>
                                                            </TableRow>;
                                                        })}
                                                    </TableBody>
                                                </Table>
                                            </TableContainer>
                                            {more && <Button size="small" sx={{mt: 1}} disabled={loading} onClick={() => load(runs[runs.length - 1].id)}>
                                                {loading ? 'Loading…' : 'Older runs'}
                                            </Button>}
                                            {!summarized && <Typography variant="caption" color="text.secondary" component="div" sx={{mt: 1}}>
                                                The master doesn't summarize runs yet; open a run for its results.
                                            </Typography>}
                                        </>}
                        </Section>
                    </Stack>
                </Grid>

                <Grid size={{xs: 12, lg: 4}}>
                    <Stack spacing={2}>
                        <Section title="What it runs" action={<Button size="small" component={RouterLink} to="edit">Edit</Button>}>
                            {unity
                                ? <>
                                    <Typography variant="body2" sx={{mb: 1}}>
                                        {unity.testCategoryType === UnityTestCategory.RunAllTests
                                            ? 'All Unity tests of the app.'
                                            : unity.testCategoryType === UnityTestCategory.RunAllOfCategory
                                                ? <>All Unity tests of {unity.categories ? <>the categories <b>{unity.categories}</b></> : 'a category (none set)'}.</>
                                                : `${unity.testFunctions.length} picked Unity ${unity.testFunctions.length === 1 ? 'test' : 'tests'}:`}
                                    </Typography>
                                    {unity.testCategoryType === UnityTestCategory.RunSelectedTestsOnly && <Box sx={{maxHeight: 320, overflowY: 'auto'}}>
                                        {functionsByClass.map(([cls, methods]) => <Box key={cls} sx={{mb: 1}}>
                                            <Typography variant="caption" color="text.secondary" sx={{wordBreak: 'break-all'}}>{cls}</Typography>
                                            {methods.map(m => <Typography key={m} variant="body2" sx={{fontFamily: 'monospace', fontSize: 13, pl: 1}}>{m}</Typography>)}
                                        </Box>)}
                                    </Box>}
                                </>
                                : <Typography variant="body2">A {TestType[config.type] ?? ''} test.</Typography>}
                        </Section>

                        <Section title="Where it runs">
                            <Typography variant="body2" sx={{mb: 1}}>
                                {config.executionType === TestExecutionType.Simultaneously
                                    ? 'Every test runs on every device, for results across devices.'
                                    : 'Each test runs on the next free device, for faster results.'}
                            </Typography>
                            {config.allDevices
                                ? <Typography variant="body2" color="text.secondary">On all devices.</Typography>
                                : (config.devices ?? []).length === 0
                                    ? <Alert severity="warning">No devices are picked, so the test can't run. Pick devices under Edit.</Alert>
                                    : <Stack spacing={0.5}>
                                        {config.devices.map((d, i) => <Link key={`${d.deviceId}_${i}`} component={RouterLink} underline="hover"
                                                                            to={`/project/${projectIdentifier}/device/${d.deviceId}`} variant="body2">
                                            {d.device ? (d.device.alias || d.device.name) : `Device ${d.deviceId}`}
                                        </Link>)}
                                    </Stack>}
                        </Section>
                    </Stack>
                </Grid>
            </Grid>

            {app && <RunTestDialog projectIdentifier={projectIdentifier} app={app} test={runDialog ? test : null}
                                   onClose={() => setRunDialog(false)}/>}
        </Box>
    );
};

export default ShowTestPage;
