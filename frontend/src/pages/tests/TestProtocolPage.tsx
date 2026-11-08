import React, {useEffect, useMemo, useRef, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    ButtonBase,
    Card,
    CardContent,
    Chip,
    CircularProgress,
    Dialog,
    DialogContent,
    IconButton,
    InputAdornment,
    Link,
    Stack,
    Tab,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Tabs,
    TextField,
    ToggleButton,
    ToggleButtonGroup,
    Tooltip,
    Typography,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import {useColorScheme, useTheme} from '@mui/material/styles';
// error text needs the lighter red of the ramp to read on the dark surface
import {gray, red} from '../../shared-theme/themePrimitives';
import {LineChart} from '@mui/x-charts/LineChart';
import {
    ArrowBackRounded,
    ArrowForwardRounded,
    BrokenImageRounded,
    CancelRounded,
    CheckCircleRounded,
    ChevronLeftRounded,
    ChevronRightRounded,
    CloseRounded,
    CompareArrowsRounded,
    DownloadRounded,
    ErrorRounded,
    HourglassEmptyRounded,
    SearchRounded,
    SkipNextRounded,
    SouthRounded,
    NorthRounded,
} from '@mui/icons-material';
import {useSSE} from 'react-hooks-sse';
import {Link as RouterLink, useNavigate, useSearchParams} from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import ITestProtocolData from '../../types/test.protocol';
import IProtocolEntryData from '../../types/protocol.entry';
import {TestResultState} from '../../types/test.result.state.enum';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {formatDateTime} from '../../components/DateTime';
import {useLoader} from '../../hooks/useLoader';
import {getBaseline, IBaseline, IBaselineCandidate} from '../../services/test.run.service';
import IProtocolPerformanceEntryData from '../../types/protocol.performance.entry';
import {parseTags} from '../apps/bundlesModel';
import BaselinePicker from './BaselinePicker';
import {between, buildRows, formatDuration, leaves, mergeProtocol, splitTestName} from './testRunModel';
import {
    averageOf,
    bundleLabel,
    checkpointRows,
    compare,
    earlierRuns,
    filterLog,
    formatRuntime,
    isError,
    isScreen,
    logAsText,
    measured,
    Metric,
    METRICS,
    neighbours,
    sourceName,
    splitMessage,
    summarize,
    Verdict,
} from './protocolModel';

interface TestProtocolPageProps {
    run: ITestRunData
    protocol: ITestProtocolData
    // all protocols of the run, once loaded
    runProtocols: ITestProtocolData[] | null
}

interface NewTestProtocolLogPayload {
    testProtocolId: number,
    entry: IProtocolEntryData,
}

type TabKey = 'overview' | 'log' | 'screenshots' | 'performance';

// old links used other tab names
const tabFromQuery = (t: string | null): TabKey => {
    switch (t) {
        case 'log':
        case 'screenshots':
        case 'performance':
            return t;
        case 'checkpoint':
            return 'performance';
        default:
            return 'overview';
    }
};

const RESULT: Record<TestResultState, { label: string, color: 'success' | 'error' | 'warning' | 'info' }> = {
    [TestResultState.TestResultSuccess]: {label: 'Passed', color: 'success'},
    [TestResultState.TestResultFailed]: {label: 'Failed', color: 'error'},
    [TestResultState.TestResultUnstable]: {label: 'Unstable', color: 'warning'},
    [TestResultState.TestResultOpen]: {label: 'Running', color: 'info'},
};

const ResultIcon: React.FC<{ result: TestResultState }> = ({result}) => {
    switch (result) {
        case TestResultState.TestResultFailed:
            return <CancelRounded color="error" fontSize="small"/>;
        case TestResultState.TestResultUnstable:
            return <ErrorRounded color="warning" fontSize="small"/>;
        case TestResultState.TestResultSuccess:
            return <CheckCircleRounded color="success" fontSize="small"/>;
        default:
            return <HourglassEmptyRounded color="disabled" fontSize="small"/>;
    }
};

const screenUrl = (e: IProtocolEntryData) => `/api/data/${encodeURIComponent(e.data)}`;

const Section: React.FC<{ title: string, action?: React.ReactNode, children: React.ReactNode }> = ({title, action, children}) =>
    <Card variant="outlined">
        <CardContent>
            <Stack direction="row" alignItems="center" spacing={1} useFlexGap sx={{mb: 1.5, minHeight: 32}}>
                <Typography component="h3" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
                {action && <Box sx={{ml: 'auto'}}>{action}</Box>}
            </Stack>
            {children}
        </CardContent>
    </Card>;

// Screenshot shows a captured screen, or says that its file is gone (old
// data is cleaned up on the master)
const Screenshot: React.FC<{ entry: IProtocolEntryData, height: number | string, onClick?: () => void }> = ({entry, height, onClick}) => {
    const [broken, setBroken] = useState(false);
    const content = broken
        ? <Stack alignItems="center" justifyContent="center" spacing={0.5}
                 sx={{height, width: '100%', bgcolor: 'action.hover', color: 'text.secondary', p: 1, textAlign: 'center'}}>
            <BrokenImageRounded/>
            <Typography variant="caption">The file is no longer on the master</Typography>
        </Stack>
        : <Box component="img" src={screenUrl(entry)} alt={entry.message || `Screenshot at ${formatRuntime(entry.runtime)}`}
               onError={() => setBroken(true)} loading="lazy"
               sx={{height, width: '100%', objectFit: 'contain', display: 'block', bgcolor: 'action.hover'}}/>;
    return onClick && !broken
        ? <ButtonBase onClick={onClick} sx={{display: 'block', width: '100%', borderRadius: 1, overflow: 'hidden'}}
                      aria-label={`Open screenshot at ${formatRuntime(entry.runtime)}`}>{content}</ButtonBase>
        : <Box sx={{borderRadius: 1, overflow: 'hidden'}}>{content}</Box>;
};

// Message shows the first line of a log message and folds the rest (stack
// traces, JSON) behind "more"
const Message: React.FC<{ text: string, error?: boolean, open?: boolean }> = ({text, error, open: initiallyOpen}) => {
    const {headline, detail} = splitMessage(text);
    const [open, setOpen] = useState(!!initiallyOpen);
    return <Box sx={{minWidth: 0}}>
        <Typography variant="body2"
                    sx={theme => ({
                        fontFamily: 'monospace', fontSize: 13, wordBreak: 'break-word', whiteSpace: 'pre-wrap',
                        color: error ? 'error.main' : 'text.primary',
                        ...(error ? theme.applyStyles('dark', {color: red[300]}) : {}),
                    })}>
            {headline || <Typography component="span" variant="body2" color="text.secondary">(no message)</Typography>}
            {detail && !open && <Link component="button" variant="caption" onClick={() => setOpen(true)} sx={{ml: 1}}>
                more ({detail.split('\n').length} {detail.split('\n').length === 1 ? 'line' : 'lines'})
            </Link>}
        </Typography>
        {detail && open && <>
            <Typography variant="body2" color="text.secondary"
                        sx={{fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-word', whiteSpace: 'pre-wrap'}}>
                {detail}
            </Typography>
            <Link component="button" variant="caption" onClick={() => setOpen(false)}>less</Link>
        </>}
    </Box>;
};

const VerdictNote: React.FC<{ verdict: Verdict, text: string, same: string }> = ({verdict, text, same}) => {
    if (verdict === null) return null;
    const color = verdict === 'better' ? 'success.main' : verdict === 'worse' ? 'warning.main' : 'text.secondary';
    return <Typography variant="caption" sx={{color}}>
        {verdict === 'same' ? same : verdict} · {text}
    </Typography>;
};

// Comparison is what the performance of the protocol is held against: the
// latest earlier run on the device, or a run the user picked
interface Comparison {
    runId: number,
    performance: IProtocolPerformanceEntryData[],
    // the run's averages; without them the usual averages of earlier runs
    averages: IBaseline | null,
}

// MetricTiles compares only with real earlier runs: older masters averaged
// in the protocol itself and its runner
const MetricTiles: React.FC<{ protocol: ITestProtocolData, metrics: Metric[], against: Comparison | null }> = ({protocol, metrics, against}) =>
    <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 1.5}}>
        {metrics.map(m => {
            const value = m.avg(protocol);
            const picked = against?.averages ?? null;
            const usual = picked ? averageOf(picked, m.key) : against ? m.usual(protocol) : null;
            return <Box key={m.key} sx={{p: 1.5, border: 1, borderColor: 'divider', borderRadius: 1}}>
                <Typography variant="caption" color="text.secondary">{m.label} (average)</Typography>
                <Typography variant="h6" component="div">{value ? m.format(value) : '–'}</Typography>
                <VerdictNote verdict={compare(value, usual, m.higherIsBetter)}
                             same={picked ? 'same' : 'as usual'}
                             text={usual ? (picked ? `run ${against!.runId}: ${m.format(usual)}` : `usually ${m.format(usual)}`) : ''}/>
            </Box>;
        })}
    </Box>;

// CompareBar says what the performance is compared with and offers to pick
// another run
const CompareBar: React.FC<{ baseline: IBaseline | null, loading: boolean, failed: boolean, earlierRunId: number | null,
    onChange: () => void, onReset: (() => void) | null }> = ({baseline, loading, failed, earlierRunId, onChange, onReset}) =>
    <Card variant="outlined">
        <CardContent sx={{py: 1.5, '&:last-child': {pb: 1.5}}}>
            <Stack direction={{xs: 'column', sm: 'row'}} spacing={1.5} alignItems={{sm: 'center'}} useFlexGap>
                <CompareArrowsRounded color="action" sx={{display: {xs: 'none', sm: 'block'}}}/>
                <Box sx={{minWidth: 0}}>
                    {failed
                        ? <Typography variant="body2" color="error">The picked run could not be loaded.</Typography>
                        : onReset && (loading || !baseline)
                            ? <Typography variant="body2" color="text.secondary">Loading the picked run…</Typography>
                            : baseline
                                ? <>
                                    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                                        <Typography variant="body2">Compared with <b>run {baseline.testRunId}</b></Typography>
                                        {baseline.bundle && <Typography variant="body2">{bundleLabel(baseline.bundle)}</Typography>}
                                        {parseTags(baseline.bundle?.tags).map(t => <Chip key={t} size="small" variant="outlined" label={t}/>)}
                                    </Stack>
                                    <Typography variant="caption" color="text.secondary">
                                        {formatDateTime(baseline.startedAt, 'YYYY/MM/DD HH:mm')}{baseline.device && ` · ${baseline.device}`}
                                    </Typography>
                                </>
                                : earlierRunId
                                    ? <Typography variant="body2">Compared with <b>run {earlierRunId}</b>, the latest earlier run on this device</Typography>
                                    : <Typography variant="body2" color="text.secondary">No earlier run on this device to compare with</Typography>}
                </Box>
                <Stack direction="row" spacing={1} sx={{ml: {sm: 'auto'}, flexShrink: 0}}>
                    {onReset && <Button size="small" onClick={onReset}>Latest earlier run</Button>}
                    <Button size="small" variant="outlined" onClick={onChange}>Compare with…</Button>
                </Stack>
            </Stack>
        </CardContent>
    </Card>;

// TestProtocolPage shows one test of a run: why it failed, what it did, and
// how it performed. The loader mounts it with key={protocol.id}.
const TestProtocolPage: React.FC<TestProtocolPageProps> = (props) => {
    const {run, runProtocols} = props;
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const navigate = useNavigate();
    const theme = useTheme();
    const [query, setQuery] = useSearchParams();
    const tab = tabFromQuery(query.get('t'));
    const setTab = (t: TabKey) => setQuery(q => {
        const next = new URLSearchParams(q);
        if (t === 'overview') next.delete('t'); else next.set('t', t);
        return next;
    }, {replace: true});

    // live result and log of a running test
    const protocol = useSSE<ITestProtocolData, ITestProtocolData | null>(
        `test_run_${run.id}_protocol`, props.protocol,
        {stateReducer: (state, action) => action.data?.id === state.id ? mergeProtocol([state], action.data)[0] : state});
    const entries = useSSE<IProtocolEntryData[], NewTestProtocolLogPayload | null>(
        `test_protocol_${props.protocol.id}_log`, props.protocol.entries ?? [],
        {stateReducer: (state, action) => action.data?.entry && !state.some(e => e.id === action.data!.entry.id) ? [...state, action.data.entry] : state});
    const running = protocol.testResult === TestResultState.TestResultOpen && !protocol.endedAt;

    // the run around it: the other tests, and the runner that wrapped this test
    const tests = useMemo(() => runProtocols ? leaves(buildRows(runProtocols)) : [], [runProtocols]);
    const place = neighbours(tests, protocol.id as number);
    const row = place.position > 0 ? tests[place.position - 1] : null;
    const runner = row?.runner && row.runner.id !== protocol.id ? row.runner : null;
    const wrappedTest = row && row.runner?.id === protocol.id ? row.protocol : null;
    const runnerEntries = runner?.entries ?? [];

    const summary = useMemo(() => summarize(entries), [entries]);
    // the runner takes the screenshots when a test ends
    const screens = useMemo(() => [...summary.screens, ...runnerEntries.filter(isScreen)], [summary.screens, runnerEntries]);
    const failed = protocol.testResult === TestResultState.TestResultFailed || protocol.testResult === TestResultState.TestResultUnstable;
    const errors = useMemo(() => [...summary.errors, ...runnerEntries.filter(isError)
        .filter(r => !summary.errors.some(e => e.message === r.message))], [summary.errors, runnerEntries]);
    const lastStep = summary.steps.length > 0 ? summary.steps[summary.steps.length - 1] : null;
    const metrics = METRICS.filter(m => measured(protocol.performance, m.key));
    const earlier = earlierRuns(protocol).find(h => (h.performance ?? []).length > 0) ?? null;

    // the run to compare with: picked by the user (?vs=<run>, and
    // &vsDevice=<device> for a run on another device), or the latest earlier
    // one. It's chosen by run so it stays while moving through the tests.
    const vs = query.get('vs');
    const vsDevice = query.get('vsDevice');
    const {data: loadedBaseline, error: baselineError, loading: baselineLoading} = useLoader(
        () => vs ? getBaseline(projectIdentifier, appId as number, run.testId, run.id as number, protocol.id as number, vs, vsDevice) : null,
        [projectIdentifier, appId, run.testId, run.id, protocol.id, vs, vsDevice]);
    const baseline = vs && loadedBaseline && String(loadedBaseline.testRunId) === vs ? loadedBaseline : null;
    const comparison: Comparison | null = vs
        ? (baseline ? {runId: baseline.testRunId, performance: baseline.performance ?? [], averages: baseline} : null)
        : (earlier ? {runId: earlier.testRunId, performance: earlier.performance ?? [], averages: null} : null);
    const [picking, setPicking] = useState(false);
    const setBaseline = (c: IBaselineCandidate | null) => setQuery(q => {
        const next = new URLSearchParams(q);
        next.delete('vs');
        next.delete('vsDevice');
        if (c) {
            next.set('vs', String(c.testRunId));
            if (c.deviceId !== protocol.deviceId && c.deviceId) next.set('vsDevice', String(c.deviceId));
        }
        return next;
    });
    const checkpoints = checkpointRows(protocol.performance, comparison?.performance);

    const base = `/project/${projectIdentifier}/app:${appId}/test/${run.testId}/run/${run.id}`;
    const openTest = (p: ITestProtocolData) => {
        const keep = new URLSearchParams();
        if (tab !== 'overview') keep.set('t', tab);
        if (vs) keep.set('vs', vs);
        if (vsDevice) keep.set('vsDevice', vsDevice);
        const q = keep.toString();
        navigate(`${base}/${p.id}${q ? `?${q}` : ''}`);
    };
    const {name, suite} = splitTestName(protocol.testName);
    // the runner times a test as the run does; the test's own start is when
    // the app connected
    const timing = runner ?? protocol;
    const ms = running ? null : between(timing.startedAt, timing.endedAt);
    const result = RESULT[protocol.testResult];
    const device = protocol.device ? (protocol.device.alias?.length > 0 ? protocol.device.alias : protocol.device.name) : null;

    // log: filters, and "show in log" from the overview
    const [hidden, setHidden] = useState<Set<string>>(new Set());
    const [errorsOnly, setErrorsOnly] = useState(false);
    const [search, setSearch] = useState('');
    const [limit, setLimit] = useState(500);
    const [focus, setFocus] = useState<number | null>(null);
    const visibleLog = useMemo(() => filterLog(entries, {hidden, errorsOnly, text: search}), [entries, hidden, errorsOnly, search]);
    const showInLog = (e: IProtocolEntryData) => {
        setHidden(new Set());
        setErrorsOnly(false);
        setSearch('');
        setLimit(l => Math.max(l, entries.indexOf(e) + 50));
        setFocus(e.id as number);
        setTab('log');
    };
    useEffect(() => {
        if (tab === 'log' && focus !== null) {
            document.getElementById(`log-entry-${focus}`)?.scrollIntoView({block: 'center'});
        }
    }, [tab, focus]);

    const download = () => {
        const blob = new Blob([logAsText(protocol.testName, entries)], {type: 'text/plain'});
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `protocol-${protocol.id}-${name}.txt`;
        a.click();
        setTimeout(() => URL.revokeObjectURL(url), 1000);
    };

    const [viewer, setViewer] = useState<number | null>(null);
    const lastScreen = screens.length > 0 ? screens[screens.length - 1] : null;

    const counts = {log: entries.length, screenshots: screens.length, performance: (protocol.performance ?? []).length};

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            {/* which test, how it went, and the way to the other tests of the run */}
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Stack direction="row" alignItems="center" spacing={1.5} flexWrap="wrap" useFlexGap>
                        <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{name}</Typography>
                        <Chip size="small" label={running ? 'Running' : result.label} color={result.color}
                              icon={running ? <CircularProgress size={12} color="inherit"/> : undefined}/>
                    </Stack>
                    {suite && <Typography variant="body2" color="text.secondary" sx={{wordBreak: 'break-all'}}>{suite}</Typography>}
                    <Typography variant="body2" color="text.secondary" sx={{mt: 0.5}}>
                        Started {formatDateTime(timing.startedAt)}
                        {ms !== null && <> · took {formatDuration(ms)}</>}
                        {device && <> · <Link component={RouterLink} to={`/project/${projectIdentifier}/device/${protocol.deviceId}`}
                                              underline="hover" color="inherit">{device}</Link>
                            {protocol.device?.osVersion && ` (${protocol.device.os} ${protocol.device.osVersion})`}</>}
                        {' · '}<Link component={RouterLink} to={base} underline="hover" color="inherit">Run {run.id}</Link>
                    </Typography>
                </Box>
                <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {md: 'auto'}, flexWrap: 'wrap'}} useFlexGap
                       role="navigation" aria-label="Tests of this run">
                    {place.total > 1 && place.position > 0 && <>
                        <Tooltip title={place.prev ? `Previous test: ${place.prev.name}` : 'This is the first test of the run'}>
                            <span><Button size="small" startIcon={<ArrowBackRounded/>} disabled={!place.prev}
                                          onClick={() => place.prev && openTest(place.prev.protocol)}>Previous</Button></span>
                        </Tooltip>
                        <Typography variant="body2" color="text.secondary">Test {place.position} of {place.total}</Typography>
                        <Tooltip title={place.next ? `Next test: ${place.next.name}` : 'This is the last test of the run'}>
                            <span><Button size="small" endIcon={<ArrowForwardRounded/>} disabled={!place.next}
                                          onClick={() => place.next && openTest(place.next.protocol)}>Next</Button></span>
                        </Tooltip>
                        {place.nextProblem && place.nextProblem !== place.next &&
                            <Tooltip title={`Next failed or unstable test: ${place.nextProblem.name}`}>
                                <Button size="small" variant="outlined" color="error" endIcon={<SkipNextRounded/>}
                                        onClick={() => openTest(place.nextProblem!.protocol)}>Next failure</Button>
                            </Tooltip>}
                    </>}
                    {(runner || wrappedTest) && <ToggleButtonGroup size="small" exclusive value={runner ? 'test' : 'runner'}
                                                                   aria-label="Protocol of the test or of its runner"
                                                                   onChange={(_, v) => {
                                                                       if (v === 'runner' && runner) openTest(runner);
                                                                       if (v === 'test' && wrappedTest) openTest(wrappedTest);
                                                                   }}>
                        <Tooltip title="What the test did: its steps, log and performance"><ToggleButton value="test">Test</ToggleButton></Tooltip>
                        <Tooltip title="How the hub started and ended the test, with the final screenshot"><ToggleButton value="runner">Runner</ToggleButton></Tooltip>
                    </ToggleButtonGroup>}
                </Stack>
            </Stack>

            <Box sx={{borderBottom: 1, borderColor: 'divider', mb: 2}}>
                <Tabs value={tab} onChange={(_, v) => setTab(v)} variant="scrollable" allowScrollButtonsMobile>
                    <Tab value="overview" label="Overview"/>
                    <Tab value="log" label={`Log ${counts.log}`} disabled={counts.log === 0}/>
                    <Tab value="screenshots" label={`Screenshots ${counts.screenshots}`} disabled={counts.screenshots === 0}/>
                    <Tab value="performance" label={`Performance ${counts.performance}`} disabled={counts.performance === 0}/>
                </Tabs>
            </Box>

            {tab === 'overview' && <Grid container spacing={2}>
                <Grid size={{xs: 12, lg: lastScreen ? 8 : 12}}>
                    <Stack spacing={2}>
                        {failed && <Section title={protocol.testResult === TestResultState.TestResultFailed ? 'Why it failed' : 'Why it is unstable'}
                                            action={errors.length > 3 && <Button size="small" onClick={() => {
                                                setErrorsOnly(true);
                                                setTab('log');
                                            }}>All {errors.length} errors</Button>}>
                            {errors.length === 0
                                ? <Typography variant="body2" color="text.secondary">
                                    The test reported no error message. The log shows what it did last.
                                </Typography>
                                : <Stack spacing={1.5}>
                                    {lastStep && <Box>
                                        <Typography variant="caption" color="text.secondary">
                                            Last step, at {formatRuntime(lastStep.runtime)}
                                        </Typography>
                                        <Message text={lastStep.message}/>
                                    </Box>}
                                    {errors.slice(-3).map(e => {
                                        const fromRunner = !summary.errors.includes(e);
                                        return <Box key={`${e.id}`} sx={{borderLeft: 3, borderColor: (theme.vars ?? theme).palette.error.main, pl: 1.5}}>
                                            <Stack direction="row" spacing={1} alignItems="center" useFlexGap sx={{mb: 0.5}}>
                                                <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace'}}>
                                                    {formatRuntime(e.runtime)}
                                                </Typography>
                                                <Chip size="small" variant="outlined" label={fromRunner ? 'runner' : sourceName(e.source)}/>
                                                {!fromRunner && <Link component="button" variant="caption" onClick={() => showInLog(e)} sx={{ml: 'auto'}}>
                                                    Show in log
                                                </Link>}
                                            </Stack>
                                            <Message text={e.message} error/>
                                        </Box>;
                                    })}
                                </Stack>}
                        </Section>}

                        <Section title="Steps" action={summary.steps.length > 0 &&
                            <Typography variant="body2" color="text.secondary">{summary.steps.length}</Typography>}>
                            {summary.steps.length === 0
                                ? <Typography variant="body2" color="text.secondary">
                                    {wrappedTest ? 'The runner has no steps; switch to "Test" for what the test did.' : 'The test reported no steps.'}
                                </Typography>
                                : <Box component="ol" sx={{listStyle: 'none', m: 0, p: 0}}>
                                    {summary.steps.map((s, i) => {
                                        const last = i === summary.steps.length - 1;
                                        return <Box component="li" key={s.id} sx={{display: 'grid', gridTemplateColumns: '80px 16px 1fr', gap: 1, alignItems: 'start'}}>
                                            <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>
                                                {formatRuntime(s.runtime)}
                                            </Typography>
                                            <Box sx={{position: 'relative', height: '100%', minHeight: 28}}>
                                                <Box sx={{position: 'absolute', left: 5, top: 7, width: 7, height: 7, borderRadius: '50%',
                                                    bgcolor: last && failed ? 'error.main' : 'text.secondary'}}/>
                                                {!last && <Box sx={{position: 'absolute', left: 8, top: 16, bottom: -4, width: '1px', bgcolor: 'divider'}}/>}
                                            </Box>
                                            <Box sx={{pb: 1}}><Message text={s.message}/></Box>
                                        </Box>;
                                    })}
                                    {!running && <Box component="li" sx={{display: 'grid', gridTemplateColumns: '80px 16px 1fr', gap: 1}}>
                                        <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>
                                            {entries.length > 0 ? formatRuntime(entries[entries.length - 1].runtime) : ''}
                                        </Typography>
                                        <Box sx={{pt: '2px'}}><ResultIcon result={protocol.testResult}/></Box>
                                        <Typography variant="body2" sx={{fontWeight: 600}}>{result.label}</Typography>
                                    </Box>}
                                </Box>}
                        </Section>

                        {metrics.length > 0 && <Section title="Performance" action={
                            <Button size="small" onClick={() => setTab('performance')}>Charts and checkpoints</Button>}>
                            <MetricTiles protocol={protocol} metrics={metrics} against={comparison}/>
                        </Section>}
                    </Stack>
                </Grid>
                {lastScreen && <Grid size={{xs: 12, lg: 4}}>
                    <Section title={failed ? 'Screen at the end' : 'Last screenshot'} action={screens.length > 1 &&
                        <Button size="small" onClick={() => setTab('screenshots')}>All {screens.length}</Button>}>
                        <Screenshot entry={lastScreen} height={420} onClick={() => setViewer(screens.length - 1)}/>
                        <Typography variant="caption" color="text.secondary">at {formatRuntime(lastScreen.runtime)}</Typography>
                    </Section>
                </Grid>}
            </Grid>}

            {tab === 'log' && <Card variant="outlined">
                <CardContent>
                    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{mb: 1.5}}>
                        <TextField size="small" placeholder="Find in log" value={search} onChange={e => setSearch(e.target.value)}
                                   sx={{width: {xs: '100%', sm: 260}}}
                                   slotProps={{
                                       input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                                       htmlInput: {'aria-label': 'Find in log'},
                                   }}/>
                        <Chip size="small" label={`Errors only ${summary.errors.length}`} disabled={summary.errors.length === 0}
                              color={errorsOnly ? 'error' : 'default'} variant={errorsOnly ? 'filled' : 'outlined'}
                              onClick={() => setErrorsOnly(v => !v)}/>
                        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap role="group" aria-label="Show sources">
                            {summary.sources.map(s => {
                                const on = !hidden.has(s.source);
                                return <Chip key={s.source} size="small" label={`${s.source} ${s.count}`}
                                             variant={on ? 'filled' : 'outlined'} aria-pressed={on}
                                             onClick={() => setHidden(prev => {
                                                 const next = new Set(prev);
                                                 if (next.has(s.source)) next.delete(s.source); else next.add(s.source);
                                                 return next;
                                             })}/>;
                            })}
                        </Stack>
                        <Button size="small" startIcon={<DownloadRounded/>} onClick={download} sx={{ml: {sm: 'auto'}}}>Download</Button>
                    </Stack>
                    <Typography variant="caption" color="text.secondary">
                        {visibleLog.length === entries.length ? `${entries.length} entries` : `${visibleLog.length} of ${entries.length} entries`}
                        {' · times since the test started'}
                    </Typography>
                    {visibleLog.length === 0
                        ? <Typography variant="body2" color="text.secondary" sx={{py: 2}}>
                            Nothing matches.{' '}
                            <Link component="button" variant="body2" onClick={() => {
                                setHidden(new Set());
                                setErrorsOnly(false);
                                setSearch('');
                            }}>Show everything</Link>
                        </Typography>
                        : <Box role="log" aria-label="Protocol log" sx={{mt: 1}}>
                            {visibleLog.slice(0, limit).map(e => {
                                const error = isError(e);
                                return <Box key={e.id} id={`log-entry-${e.id}`} sx={{
                                    display: 'grid', gridTemplateColumns: {xs: '72px 1fr', sm: '80px 92px 1fr'}, gap: 1.5, py: 0.5,
                                    borderLeft: 3, pl: 1, ml: -1.5,
                                    borderColor: error ? (theme.vars ?? theme).palette.error.main : 'transparent',
                                    bgcolor: focus === e.id ? 'action.selected' : undefined,
                                }}>
                                    <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>
                                        {formatRuntime(e.runtime)}
                                    </Typography>
                                    <Typography variant="caption" color="text.secondary" sx={{pt: '2px', display: {xs: 'none', sm: 'block'}}}>
                                        {sourceName(e.source)}
                                    </Typography>
                                    {isScreen(e)
                                        ? <Box sx={{maxWidth: 240}}>
                                            <Screenshot entry={e} height={140} onClick={() => setViewer(screens.indexOf(e))}/>
                                        </Box>
                                        : <Message text={e.message} error={error} open={focus === e.id}/>}
                                </Box>;
                            })}
                            {visibleLog.length > limit && <Button size="small" sx={{mt: 1}} onClick={() => setLimit(l => l + 500)}>
                                Show {Math.min(500, visibleLog.length - limit)} more of {visibleLog.length - limit}
                            </Button>}
                        </Box>}
                </CardContent>
            </Card>}

            {tab === 'screenshots' && <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 2}}>
                {screens.map((s, i) => <Card key={s.id} variant="outlined">
                    <Screenshot entry={s} height={360} onClick={() => setViewer(i)}/>
                    <Box sx={{p: 1}}>
                        <Typography variant="caption" color="text.secondary">
                            {formatRuntime(s.runtime)}{!summary.screens.includes(s) && ' · runner, at the end of the test'}
                        </Typography>
                        {s.message && <Typography variant="body2">{s.message}</Typography>}
                    </Box>
                </Card>)}
            </Box>}

            {tab === 'performance' && <Stack spacing={2}>
                <CompareBar baseline={baseline} loading={baselineLoading} failed={!!vs && !!baselineError}
                            earlierRunId={earlier?.testRunId ?? null}
                            onChange={() => setPicking(true)} onReset={vs ? () => setBaseline(null) : null}/>
                <BaselinePicker open={picking} onClose={() => setPicking(false)}
                                onPick={c => {
                                    setPicking(false);
                                    setBaseline(c);
                                }}
                                projectIdentifier={projectIdentifier} appId={appId as number} testId={run.testId}
                                runId={run.id as number} protocolId={protocol.id as number} deviceId={protocol.deviceId}
                                selectedRunId={comparison?.runId ?? null} metrics={metrics}/>
                {metrics.length > 0 && <Section title="Averages">
                    <MetricTiles protocol={protocol} metrics={metrics} against={comparison}/>
                </Section>}
                <Grid container spacing={2}>
                    {metrics.map(m => <Grid key={m.key} size={{xs: 12, md: 6}}>
                        <Section title={`${m.label}${m.unit ? ` (${m.unit})` : ''}`}>
                            <MetricChart metric={m} protocol={protocol} earlier={comparison}/>
                        </Section>
                    </Grid>)}
                </Grid>
                {checkpoints.length > 0 && <Section title="Checkpoints" action={comparison &&
                    <Typography variant="body2" color="text.secondary">compared with run {comparison.runId}</Typography>}>
                    <TableContainer>
                        <Table size="small" aria-label="Checkpoints">
                            <TableHead>
                                <TableRow>
                                    <TableCell>Checkpoint</TableCell>
                                    <TableCell align="right">Reached at</TableCell>
                                    <TableCell align="right">Took</TableCell>
                                    {comparison && <TableCell align="right">Run {comparison.runId}</TableCell>}
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {checkpoints.map((c, i) => {
                                    const verdict = compare(c.took, c.earlierTook, false);
                                    return <TableRow key={i}>
                                        <TableCell sx={{wordBreak: 'break-word'}}>{c.checkpoint}</TableCell>
                                        <TableCell align="right" sx={{fontFamily: 'monospace'}}>{formatRuntime(c.at)}</TableCell>
                                        <TableCell align="right" sx={{fontFamily: 'monospace', whiteSpace: 'nowrap'}}>
                                            {c.took.toFixed(2)} s
                                            {verdict === 'worse' && <Tooltip title={`Slower than run ${comparison?.runId}`}><NorthRounded fontSize="inherit" color="warning" sx={{ml: 0.5}}/></Tooltip>}
                                            {verdict === 'better' && <Tooltip title={`Faster than run ${comparison?.runId}`}><SouthRounded fontSize="inherit" color="success" sx={{ml: 0.5}}/></Tooltip>}
                                        </TableCell>
                                        {comparison && <TableCell align="right" sx={{fontFamily: 'monospace', color: 'text.secondary'}}>
                                            {c.earlierTook === null ? '–' : `${c.earlierTook.toFixed(2)} s`}
                                        </TableCell>}
                                    </TableRow>;
                                })}
                            </TableBody>
                        </Table>
                    </TableContainer>
                </Section>}
            </Stack>}

            {running && <Alert severity="info" sx={{mt: 2}}>The test is running; the log and result update live.</Alert>}

            <Dialog open={viewer !== null} onClose={() => setViewer(null)} maxWidth="lg" fullWidth
                    aria-label="Screenshot">
                {viewer !== null && screens[viewer] && <DialogContent sx={{p: 1}}>
                    <Stack direction="row" alignItems="center" spacing={1} sx={{mb: 1, px: 1}}>
                        <Typography variant="body2" color="text.secondary">
                            {viewer + 1} of {screens.length} · at {formatRuntime(screens[viewer].runtime)}
                        </Typography>
                        <Box sx={{ml: 'auto'}}/>
                        <IconButton size="small" aria-label="Previous screenshot" disabled={viewer === 0}
                                    onClick={() => setViewer(v => (v ?? 1) - 1)}><ChevronLeftRounded/></IconButton>
                        <IconButton size="small" aria-label="Next screenshot" disabled={viewer === screens.length - 1}
                                    onClick={() => setViewer(v => (v ?? 0) + 1)}><ChevronRightRounded/></IconButton>
                        <IconButton size="small" aria-label="Close" onClick={() => setViewer(null)}><CloseRounded/></IconButton>
                    </Stack>
                    <Screenshot key={screens[viewer].id} entry={screens[viewer]} height="80vh"/>
                </DialogContent>}
            </Dialog>
        </Box>
    );
};

// MetricChart plots one metric over the checkpoints of the test, with the
// compared run as a dashed reference line where it reached the same checkpoint
const MetricChart: React.FC<{ metric: Metric, protocol: ITestProtocolData, earlier: Comparison | null }> = ({metric, protocol, earlier}) => {
    const theme = useTheme();
    // the theme uses CSS variables, so theme.palette always holds the light
    // scheme; the compared run's grey has to follow the active scheme itself
    const {mode, systemMode} = useColorScheme();
    const dark = (mode === 'system' ? systemMode : mode) === 'dark';
    const perf = (protocol.performance ?? []).filter(p => p.checkpoint !== 'schedule');
    const before = new Map<string, number>();
    (earlier?.performance ?? []).forEach(p => {
        if (!before.has(p.checkpoint) && typeof p[metric.key] === 'number') before.set(p.checkpoint, p[metric.key]);
    });
    const earlierData = perf.map(p => before.get(p.checkpoint) ?? null);
    const hasEarlier = earlierData.some(v => v !== null);
    const format = (v: number | null) => v === null ? '–' : metric.format(v);
    const series = [
        {id: 'current', data: perf.map(p => p[metric.key] ?? null), label: 'This run', color: theme.palette.primary.main,
            showMark: perf.length <= 20, valueFormatter: format, curve: 'linear' as const},
        ...(hasEarlier ? [{id: 'earlier', data: earlierData, label: `Run ${earlier!.runId}`, color: dark ? gray[300] : gray[500],
            showMark: false, valueFormatter: format, connectNulls: true, curve: 'linear' as const}] : []),
    ];
    return <Box>
        <LineChart height={220} skipAnimation
                   margin={{top: hasEarlier ? 40 : 16, right: 16, bottom: 28, left: 56}}
                   xAxis={[{scaleType: 'point', data: perf.map((_, i) => i), valueFormatter: (i: number, ctx) =>
                           ctx.location === 'tooltip' ? `${formatRuntime(perf[i]?.runtime)} · ${perf[i]?.checkpoint}` : `${i + 1}`}]}
                   yAxis={[{valueFormatter: (v: number) => metric.format(v).replace(/ (fps|%)$/, '')}]}
                   series={series}
                   slotProps={{legend: {hidden: !hasEarlier}}}
                   sx={{'& .MuiLineElement-series-earlier': {strokeDasharray: '5 4'}}}/>
        <Typography variant="caption" color="text.secondary">Checkpoints in order; hover for their names.</Typography>
    </Box>;
};

export default TestProtocolPage;
