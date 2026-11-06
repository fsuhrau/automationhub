import React, {useEffect, useMemo, useRef, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    CircularProgress,
    IconButton,
    InputAdornment,
    Link,
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
import Grid from '@mui/material/Grid';
import {useTheme} from '@mui/material/styles';
import {
    ArrowBackRounded,
    ArrowForwardRounded,
    CancelRounded,
    CheckCircleRounded,
    ChevronRightRounded,
    ErrorRounded,
    ExpandLessRounded,
    ExpandMoreRounded,
    HourglassEmptyRounded,
    ListAltRounded,
    ReplayRounded,
    SearchRounded,
    StopCircleRounded,
} from '@mui/icons-material';
import {useSSE} from 'react-hooks-sse';
import {Link as RouterLink, useNavigate} from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import ITesRunLogEntryData from '../../types/test.run.log.entry';
import ITestProtocolData from '../../types/test.protocol';
import {TestResultState} from '../../types/test.result.state.enum';
import {cancelTestRun, executeTest} from '../../services/test.service';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useError} from '../../ErrorProvider';
import {formatDateTime} from '../../components/DateTime';
import {
    between,
    buildRows,
    countResults,
    filterRows,
    formatDuration,
    mergeProtocol,
    parseParameters,
    ResultCounts,
    ResultFilter,
    RunLifecycle,
    runEnd,
    runStatus,
    RunStatus,
    TestRow,
} from './testRunModel';

interface TestRunPageProps {
    testRun: ITestRunData
    nextRunId: number | null
    prevRunId: number | null
    state?: RunLifecycle
}

interface NewTestRunPayload {
    testRunId: number,
    entry: ITesRunLogEntryData,
}

type PaletteKey = 'success' | 'error' | 'warning' | 'info';

const STATUS: Record<RunStatus, { label: string, color: PaletteKey | 'default', hint: string }> = {
    running: {label: 'Running', color: 'info', hint: 'The run is executing; results update live.'},
    failed: {label: 'Failed', color: 'error', hint: 'At least one test failed.'},
    unstable: {label: 'Unstable', color: 'warning', hint: 'No test failed, but some were unstable.'},
    interrupted: {label: 'Interrupted', color: 'default', hint: 'The run stopped before all tests finished (cancelled, crashed, or the master restarted).'},
    success: {label: 'Passed', color: 'success', hint: 'All tests passed.'},
    not_run: {label: 'Did not run', color: 'error', hint: 'No test ran; the execution log has the errors.'},
    empty: {label: 'No results', color: 'default', hint: 'The run has no test results.'},
};

// results in the order they matter when reading a run
const RESULT_ORDER: TestResultState[] = [
    TestResultState.TestResultFailed,
    TestResultState.TestResultUnstable,
    TestResultState.TestResultOpen,
    TestResultState.TestResultSuccess,
];

const resultLabel = (r: TestResultState, running: boolean): string => {
    switch (r) {
        case TestResultState.TestResultFailed:
            return 'Failed';
        case TestResultState.TestResultUnstable:
            return 'Unstable';
        case TestResultState.TestResultSuccess:
            return 'Passed';
        default:
            return running ? 'Running' : 'Not finished';
    }
};

const resultPalette = (r: TestResultState): PaletteKey | 'default' => {
    switch (r) {
        case TestResultState.TestResultFailed:
            return 'error';
        case TestResultState.TestResultUnstable:
            return 'warning';
        case TestResultState.TestResultSuccess:
            return 'success';
        default:
            return 'info';
    }
};

const ResultIcon: React.FC<{ result: TestResultState, running: boolean }> = ({result, running}) => {
    const label = resultLabel(result, running);
    let icon: React.ReactElement;
    switch (result) {
        case TestResultState.TestResultFailed:
            icon = <CancelRounded color="error" fontSize="small"/>;
            break;
        case TestResultState.TestResultUnstable:
            icon = <ErrorRounded color="warning" fontSize="small"/>;
            break;
        case TestResultState.TestResultSuccess:
            icon = <CheckCircleRounded color="success" fontSize="small"/>;
            break;
        default:
            icon = running
                ? <CircularProgress size={16} color="info"/>
                : <HourglassEmptyRounded color="disabled" fontSize="small"/>;
    }
    return <Tooltip title={label}><Box component="span" sx={{display: 'inline-flex'}} aria-label={label}>{icon}</Box></Tooltip>;
};

// breakAfterDots lets long dotted names wrap after a dot, not mid-word
const breakAfterDots = (name: string): React.ReactNode =>
    name.split('.').map((part, i, all) => <React.Fragment key={i}>{part}{i < all.length - 1 && <>.<wbr/></>}</React.Fragment>);

const deviceName = (p: ITestProtocolData) => p.device ? (p.device.alias?.length > 0 ? p.device.alias : p.device.name) : '';

// useNow ticks every second while active, for live durations
const useNow = (active: boolean): number => {
    const [now, setNow] = useState(Date.now());
    useEffect(() => {
        if (!active) return;
        const timer = setInterval(() => setNow(Date.now()), 1000);
        return () => clearInterval(timer);
    }, [active]);
    return now;
};

const Section: React.FC<{ title: string, action?: React.ReactNode, children: React.ReactNode }> = ({title, action, children}) =>
    <Card variant="outlined" sx={{height: '100%'}}>
        <CardContent>
            <Stack direction="row" alignItems="center" spacing={1} useFlexGap sx={{mb: 1.5, minHeight: 32}}>
                <Typography component="h3" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
                {action && <Box sx={{ml: 'auto'}}>{action}</Box>}
            </Stack>
            {children}
        </CardContent>
    </Card>;

// ResultBar shows the share of each result as one stacked bar; the chips
// below it are the filter of the test list.
const ResultBar: React.FC<{ counts: ResultCounts, total: number }> = ({counts, total}) => {
    const theme = useTheme();
    if (total === 0) {
        return <Box sx={{height: 8, borderRadius: 4, bgcolor: 'action.hover'}}/>;
    }
    return <Box role="img" aria-label={RESULT_ORDER.map(r => `${counts[r]} ${resultLabel(r, false)}`).join(', ')}
                sx={{display: 'flex', height: 8, borderRadius: 4, overflow: 'hidden', gap: '2px'}}>
        {RESULT_ORDER.filter(r => counts[r] > 0).map(r => {
            const key = resultPalette(r);
            return <Box key={r} sx={{flexGrow: counts[r], bgcolor: (theme.vars ?? theme).palette[key as PaletteKey].main}}/>;
        })}
    </Box>;
};

const TestRows: React.FC<{
    row: TestRow,
    running: boolean,
    expanded: boolean,
    onToggle: () => void,
    onOpen: (protocolId: number) => void,
    nested?: boolean,
}> = ({row, running, expanded, onToggle, onOpen, nested}) => {
    const hasChildren = row.children.length > 0;
    const ms = between(row.startedAt, row.endedAt);
    const open = () => hasChildren ? onToggle() : onOpen(row.protocol.id as number);
    const device = deviceName(row.protocol);
    return <>
        <TableRow hover onClick={open} sx={{cursor: 'pointer', '& > td': {borderBottomColor: 'divider'}}}
                  aria-expanded={hasChildren ? expanded : undefined}>
            <TableCell sx={{width: 40, pl: nested ? 4 : 2}}>
                <ResultIcon result={row.result} running={running}/>
            </TableCell>
            <TableCell sx={{minWidth: 0}}>
                <Stack direction="row" alignItems="center" spacing={0.5}>
                    {hasChildren && (expanded ? <ExpandLessRounded fontSize="small"/> : <ExpandMoreRounded fontSize="small"/>)}
                    <Box sx={{minWidth: 0}}>
                        <Typography variant="body2" sx={{fontWeight: 600, wordBreak: 'break-word'}}>
                            {row.name}
                            {hasChildren && <Typography component="span" variant="caption" color="text.secondary" sx={{ml: 1}}>
                                {row.children.length} tests
                            </Typography>}
                        </Typography>
                        {row.suite && <Typography variant="caption" color="text.secondary" component="div">
                            {breakAfterDots(row.suite)}
                        </Typography>}
                    </Box>
                </Stack>
            </TableCell>
            <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>
                <Typography variant="body2">{device}</Typography>
                {row.protocol.device && <Typography variant="caption" color="text.secondary">
                    {row.protocol.device.os} {row.protocol.device.osVersion}
                </Typography>}
            </TableCell>
            <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>
                <Typography variant="body2" color={ms === null ? 'text.secondary' : 'text.primary'}>
                    {ms === null ? (running ? 'running' : '–') : formatDuration(ms)}
                </Typography>
            </TableCell>
            <TableCell align="right" sx={{width: 88, whiteSpace: 'nowrap', pr: 1}}>
                {row.runner && <Tooltip title="Runner protocol: how the hub started and ended this test">
                    <IconButton size="small" aria-label={`Runner protocol of ${row.name}`}
                                onClick={e => {
                                    e.stopPropagation();
                                    onOpen(row.runner!.id as number);
                                }}>
                        <ListAltRounded fontSize="small"/>
                    </IconButton>
                </Tooltip>}
                {!hasChildren && <Tooltip title="Open protocol">
                    <IconButton size="small" aria-label={`Open protocol of ${row.name}`}
                                onClick={e => {
                                    e.stopPropagation();
                                    onOpen(row.protocol.id as number);
                                }}>
                        <ChevronRightRounded fontSize="small"/>
                    </IconButton>
                </Tooltip>}
            </TableCell>
        </TableRow>
        {hasChildren && expanded && row.children.map(child =>
            <TestRows key={child.key} row={child} running={running} expanded={false} onToggle={() => undefined}
                      onOpen={onOpen} nested/>)}
    </>;
};

const KeyValue: React.FC<{ name: string, children: React.ReactNode }> = ({name, children}) =>
    <Box sx={{display: 'grid', gridTemplateColumns: 'minmax(80px, 35%) 1fr', gap: 1, py: 0.5}}>
        <Typography variant="body2" color="text.secondary" sx={{wordBreak: 'break-word'}}>{name}</Typography>
        <Typography variant="body2" sx={{wordBreak: 'break-all', fontFamily: 'monospace'}}>{children}</Typography>
    </Box>;

// TestRunPage shows one run. The loader mounts it with key={testRun.id}, so
// all state below belongs to exactly one run.
const TestRunPage: React.FC<TestRunPageProps> = (props: TestRunPageProps) => {

    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {testRun, nextRunId, prevRunId} = props;
    const {setError} = useError();
    const navigate = useNavigate();
    const theme = useTheme();

    // live protocol and log updates are folded into state by the reducers, so
    // events arriving in the same render are not lost
    const protocols = useSSE<ITestProtocolData[], ITestProtocolData | null>(
        `test_run_${testRun.id}_protocol`, testRun.protocols ?? [],
        {stateReducer: (state, action) => mergeProtocol(state, action.data)});
    const log = useSSE<ITesRunLogEntryData[], NewTestRunPayload | null>(
        `test_run_${testRun.id}_log`, testRun.log ?? [],
        {stateReducer: (state, action) => action.data?.entry && !state.some(e => e.id === action.data!.entry.id) ? [...state, action.data.entry] : state});
    const finished = useSSE<object | null>(`test_run_${testRun.id}_finished`, null);

    const [cancelling, setCancelling] = useState(false);
    const lifecycle: RunLifecycle = finished ? 'finished' : props.state;
    const running = lifecycle === 'running';

    const [resultFilter, setResultFilter] = useState<ResultFilter>('all');
    const [search, setSearch] = useState('');
    const [expanded, setExpanded] = useState<Set<number>>(new Set());
    const [errorsOnly, setErrorsOnly] = useState(false);

    const rows = useMemo(() => buildRows(protocols), [protocols]);
    const counts = useMemo(() => countResults(rows), [rows]);
    const visibleRows = useMemo(() => filterRows(rows, resultFilter, search), [rows, resultFilter, search]);
    const total = RESULT_ORDER.reduce((sum, r) => sum + counts[r], 0);
    const logErrors = useMemo(() => log.filter(l => l.level === 'error'), [log]);
    const status = runStatus(lifecycle, counts, logErrors.length);
    const filtering = resultFilter !== 'all' || search.trim() !== '';

    const now = useNow(running);
    const end = runEnd(lifecycle, protocols, log.map(l => l.createdAt));
    const elapsed = running ? now - new Date(testRun.createdAt).valueOf() : between(testRun.createdAt, end);

    const base = `/project/${projectIdentifier}/app:${appId}`;
    const runUrl = (runId: number | null | undefined) => `${base}/test/${testRun.testId}/run/${runId}`;
    const openProtocol = (protocolId: number) => navigate(`${runUrl(testRun.id)}/${protocolId}`);

    const onRerun = (): void => {
        executeTest(projectIdentifier, appId, testRun.testId, {
            appBinaryId: testRun.appBinaryId,
            params: testRun.parameter,
            startUrl: testRun.startUrl,
        }).then(run => navigate(runUrl(run.id))).catch(error => setError(error));
    };

    const onCancel = (): void => {
        setCancelling(true);
        cancelTestRun(projectIdentifier, appId, testRun.testId, testRun.id!).catch(error => {
            setCancelling(false);
            setError(error);
        });
    };

    const toggle = (key: number) => setExpanded(prev => {
        const next = new Set(prev);
        if (next.has(key)) next.delete(key); else next.add(key);
        return next;
    });

    const parameters = parseParameters(testRun.parameter);
    const errorCount = logErrors.length;
    const visibleLog = errorsOnly ? logErrors : log;
    const devices = new Set(protocols.map(p => p.deviceId).filter(id => id)).size || testRun.deviceStatus.length;

    // keep the log at its end while the run writes to it, unless scrolled up
    const logBox = useRef<HTMLDivElement>(null);
    const atEnd = useRef(true);
    useEffect(() => {
        const box = logBox.current;
        if (box && running && atEnd.current) box.scrollTop = box.scrollHeight;
    }, [log.length, running]);

    const statusInfo = STATUS[status];
    const done = total - counts[TestResultState.TestResultOpen];

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            {/* what run this is, how it went, and what can be done with it */}
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Stack direction="row" alignItems="center" spacing={1.5} flexWrap="wrap" useFlexGap>
                        <Typography component="h1" variant="h5">
                            {testRun.test.name} <Typography component="span" variant="h5" color="text.secondary">· Run {testRun.id}</Typography>
                        </Typography>
                        <Tooltip title={statusInfo.hint}>
                            <Chip label={statusInfo.label} color={statusInfo.color} size="small"
                                  icon={running ? <CircularProgress size={12} color="inherit"/> : undefined}/>
                        </Tooltip>
                    </Stack>
                    <Typography variant="body2" color="text.secondary" sx={{mt: 0.5}}>
                        Started {formatDateTime(testRun.createdAt, 'YYYY/MM/DD HH:mm')}
                        {elapsed !== null && <> · {running ? 'running for' : 'took'} {formatDuration(elapsed)}</>}
                        {devices > 0 && <> · {devices} {devices === 1 ? 'device' : 'devices'}</>}
                        {testRun.appBinary && <> · {testRun.appBinary.name} {testRun.appBinary.version}</>}
                    </Typography>
                </Box>
                <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {md: 'auto'}, flexWrap: 'wrap'}} useFlexGap>
                    {(prevRunId || nextRunId) ? <Stack direction="row" spacing={0.5} alignItems="center"
                                                       role="navigation" aria-label="Runs of this test">
                        {prevRunId ? <Tooltip title="Previous run of this test">
                            <Button size="small" variant="text" startIcon={<ArrowBackRounded/>}
                                    onClick={() => navigate(runUrl(prevRunId))}>
                                Run {prevRunId}
                            </Button>
                        </Tooltip> : null}
                        {nextRunId
                            ? <Tooltip title="Next run of this test">
                                <Button size="small" variant="text" endIcon={<ArrowForwardRounded/>}
                                        onClick={() => navigate(runUrl(nextRunId))}>
                                    Run {nextRunId}
                                </Button>
                            </Tooltip>
                            : <Typography variant="caption" color="text.secondary" sx={{px: 1}}>Latest run</Typography>}
                    </Stack> : null}
                    {running && <Button size="small" variant="outlined" color="error" startIcon={<StopCircleRounded/>}
                                        disabled={cancelling} onClick={onCancel}>
                        {cancelling ? 'Cancelling…' : 'Cancel run'}
                    </Button>}
                    <Tooltip title="Start this test again with the same bundle and parameters">
                        <Button size="small" variant="contained" startIcon={<ReplayRounded/>} onClick={onRerun}>
                            Run again
                        </Button>
                    </Tooltip>
                </Stack>
            </Stack>

            <Grid container spacing={2}>
                <Grid size={{xs: 12, lg: 8}}>
                    <Stack spacing={2}>
                        <Section title="Tests" action={total > 0 && <Typography variant="body2" color="text.secondary">
                            {running ? `${done} of ${total} done` : `${total} ${total === 1 ? 'test' : 'tests'}`}
                        </Typography>}>
                            {total > 0 && <ResultBar counts={counts} total={total}/>}
                            {total > 0 && <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{mt: 1.5, mb: 1}}>
                                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap role="group" aria-label="Show tests by result">
                                    <Chip size="small" label={`All ${total}`}
                                          color={resultFilter === 'all' ? 'primary' : 'default'}
                                          variant={resultFilter === 'all' ? 'filled' : 'outlined'}
                                          onClick={() => setResultFilter('all')}/>
                                    {RESULT_ORDER.filter(r => counts[r] > 0 || r !== TestResultState.TestResultOpen).map(r =>
                                        <Chip key={r} size="small" label={`${resultLabel(r, running)} ${counts[r]}`}
                                              disabled={counts[r] === 0}
                                              color={resultFilter === r ? resultPalette(r) : 'default'}
                                              variant={resultFilter === r ? 'filled' : 'outlined'}
                                              onClick={() => setResultFilter(resultFilter === r ? 'all' : r)}/>)}
                                </Stack>
                                {total > 5 && <TextField size="small" placeholder="Find a test" value={search}
                                                         onChange={e => setSearch(e.target.value)}
                                                         sx={{ml: {sm: 'auto'}, width: {xs: '100%', sm: 220}}}
                                                         slotProps={{
                                                             input: {
                                                                 startAdornment: <InputAdornment position="start">
                                                                     <SearchRounded fontSize="small"/>
                                                                 </InputAdornment>,
                                                             },
                                                             htmlInput: {'aria-label': 'Find a test'},
                                                         }}/>}
                            </Stack>}
                            {total === 0
                                ? (status === 'not_run'
                                    ? <Alert severity="error" variant="outlined">
                                        <Typography variant="body2" sx={{fontWeight: 600, mb: 0.5}}>No test ran. The run stopped with:</Typography>
                                        {logErrors.slice(-3).map(e => <Typography key={e.id} variant="body2"
                                                                                    sx={{fontFamily: 'monospace', fontSize: 13}}>
                                            {e.log}
                                        </Typography>)}
                                    </Alert>
                                    : <Typography variant="body2" color="text.secondary">
                                        {running ? 'Waiting for the first test to start…' : 'This run has no test results.'}
                                    </Typography>)
                                : visibleRows.length === 0
                                    ? <Typography variant="body2" color="text.secondary" sx={{py: 2}}>
                                        No tests match.{' '}
                                        <Link component="button" variant="body2" onClick={() => {
                                            setResultFilter('all');
                                            setSearch('');
                                        }}>Show all</Link>
                                    </Typography>
                                    : <TableContainer sx={{mx: -2, width: 'auto'}}>
                                        <Table size="small" aria-label="Tests of this run">
                                            <TableHead>
                                                <TableRow>
                                                    <TableCell sx={{pl: 2}}><Box component="span" sx={visuallyHidden}>Result</Box></TableCell>
                                                    <TableCell>Test</TableCell>
                                                    <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>Device</TableCell>
                                                    <TableCell align="right">Duration</TableCell>
                                                    <TableCell><Box component="span" sx={visuallyHidden}>Actions</Box></TableCell>
                                                </TableRow>
                                            </TableHead>
                                            <TableBody>
                                                {visibleRows.map(row =>
                                                    <TestRows key={row.key} row={row} running={running}
                                                              expanded={filtering || expanded.has(row.key)}
                                                              onToggle={() => toggle(row.key)} onOpen={openProtocol}/>)}
                                            </TableBody>
                                        </Table>
                                    </TableContainer>}
                        </Section>

                        <Section title={`Execution log`} action={errorCount > 0 &&
                            <Chip size="small" color={errorsOnly ? 'error' : 'default'} variant={errorsOnly ? 'filled' : 'outlined'}
                                  label={`Errors only ${errorCount}`} onClick={() => setErrorsOnly(v => !v)}/>}>
                            {log.length === 0
                                ? <Typography variant="body2" color="text.secondary">No log yet.</Typography>
                                : <Box ref={logBox} onScroll={e => {
                                    const box = e.currentTarget;
                                    atEnd.current = box.scrollHeight - box.scrollTop - box.clientHeight < 24;
                                }} sx={{maxHeight: 420, overflowY: 'auto', mx: -2, px: 2}}
                                       role="log" aria-label="Execution log">
                                    {visibleLog.map(entry => {
                                        const isError = entry.level === 'error';
                                        return <Box key={entry.id} sx={{
                                            display: 'grid', gridTemplateColumns: '72px 1fr', gap: 1.5, py: 0.5,
                                            borderLeft: 3, pl: 1, ml: -1.5,
                                            borderColor: isError ? (theme.vars ?? theme).palette.error.main : 'transparent',
                                        }}>
                                            <Tooltip title={formatDateTime(entry.createdAt)} placement="left">
                                                <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>
                                                    {formatDateTime(entry.createdAt, 'HH:mm:ss')}
                                                </Typography>
                                            </Tooltip>
                                            <Typography variant="body2" color={isError ? 'error' : 'text.primary'}
                                                        sx={{fontFamily: 'monospace', fontSize: 13, wordBreak: 'break-word'}}>
                                                {entry.log}
                                            </Typography>
                                        </Box>;
                                    })}
                                </Box>}
                        </Section>
                    </Stack>
                </Grid>

                <Grid size={{xs: 12, lg: 4}}>
                    <Stack spacing={2}>
                        <Section title="Devices">
                            {testRun.deviceStatus.length === 0
                                ? <Typography variant="body2" color="text.secondary">
                                    {running ? 'No device has started yet.' : 'No device started the app.'}
                                </Typography>
                                : testRun.deviceStatus.map(d => {
                                    const avg = d.histAvgStartupTime;
                                    const slower = avg > 0 && d.startupTime > avg * 1.2;
                                    const faster = avg > 0 && d.startupTime < avg * 0.8;
                                    return <Box key={d.id ?? d.deviceId} sx={{py: 0.5}}>
                                        <Link component={RouterLink} to={`/project/${projectIdentifier}/device/${d.deviceId}`}
                                              variant="body2" underline="hover" sx={{fontWeight: 600}}>
                                            {d.device ? (d.device.alias?.length > 0 ? d.device.alias : d.device.name) : `Device ${d.deviceId}`}
                                        </Link>
                                        <Typography variant="body2" color="text.secondary">
                                            App started in{' '}
                                            <Box component="span" sx={{
                                                color: slower ? 'warning.main' : faster ? 'success.main' : 'text.primary',
                                                fontWeight: 600,
                                            }}>{(d.startupTime / 1000).toFixed(1)} s</Box>
                                            {avg > 0 && <> (usually {(avg / 1000).toFixed(1)} s)</>}
                                        </Typography>
                                    </Box>;
                                })}
                        </Section>

                        <Section title="Parameters">
                            {parameters.length === 0 && !testRun.startUrl
                                ? <Typography variant="body2" color="text.secondary">The run had no parameters.</Typography>
                                : <>
                                    {testRun.startUrl && <KeyValue name="Start URL">{testRun.startUrl}</KeyValue>}
                                    {parameters.map((p, i) => <KeyValue key={i} name={p.name}>
                                        {p.value || <Typography component="span" variant="body2" color="text.secondary">(empty)</Typography>}
                                    </KeyValue>)}
                                </>}
                        </Section>

                        <Section title="App bundle" action={testRun.appBinary &&
                            <Link component={RouterLink} to={`${base}/bundles`} variant="body2">All bundles</Link>}>
                            {testRun.appBinary
                                ? <>
                                    <KeyValue name="Name">{testRun.appBinary.name}</KeyValue>
                                    <KeyValue name="Version">{testRun.appBinary.version}</KeyValue>
                                    <KeyValue name="Identifier">{testRun.appBinary.identifier}</KeyValue>
                                    <KeyValue name="Uploaded">{formatDateTime(testRun.appBinary.createdAt)}</KeyValue>
                                    {testRun.appBinary.additional && <KeyValue name="Addons">{testRun.appBinary.additional}</KeyValue>}
                                    <KeyValue name="Hash">{testRun.appBinary.hash}</KeyValue>
                                </>
                                : <Typography variant="body2" color="text.secondary">
                                    No bundle was installed for this run; it used the app already on the device or editor.
                                </Typography>}
                        </Section>
                    </Stack>
                </Grid>
            </Grid>

            {status === 'interrupted' && !running && <Alert severity="info" sx={{mt: 2}}>
                Some tests never finished. If the master restarted during the run, their results are lost; run the test again.
            </Alert>}
        </Box>
    );
};

const visuallyHidden = {
    border: 0, clip: 'rect(0 0 0 0)', height: '1px', margin: '-1px', overflow: 'hidden',
    padding: 0, position: 'absolute', whiteSpace: 'nowrap', width: '1px',
} as const;

export default TestRunPage;
