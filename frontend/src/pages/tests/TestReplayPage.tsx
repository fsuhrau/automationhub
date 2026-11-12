import React, {memo, useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    IconButton,
    MenuItem,
    Select,
    Slider,
    Stack,
    Tooltip,
    Typography,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import {
    ArrowBackRounded,
    CameraAltRounded,
    CheckCircleRounded,
    ErrorRounded,
    Forward5Rounded,
    PauseRounded,
    PlayArrowRounded,
    PlayCircleOutlineRounded,
    Replay5Rounded,
    SkipNextRounded,
    TouchAppRounded,
    WarningRounded,
} from '@mui/icons-material';
import {Link as RouterLink, useParams, useSearchParams} from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import {TestResultState} from '../../types/test.result.state.enum';
import {getRun} from '../../services/test.run.service';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {useLoader} from '../../hooks/useLoader';
import PageLoader from '../../components/PageLoader';
import ReplayVideo from '../../components/ReplayVideo';
import {formatDateTime} from '../../components/DateTime';
import {
    buildTimeline,
    eventIndexAt,
    eventIsError,
    formatOffset,
    replayRange,
    ReplayEvent,
    segmentsOf,
} from './replayModel';

const RATES = [0.5, 1, 2, 4];

// useReplayClock is the time of the replay, in ms since the epoch. It moves
// with requestAnimationFrame while playing; the page re-renders ten times a
// second, which is enough for the slider and the log.
const useReplayClock = (start: number, end: number, initial: number) => {
    const [position, setPosition] = useState(initial);
    const [playing, setPlaying] = useState(false);
    const [rate, setRate] = useState(1);
    const pos = useRef(initial);

    useEffect(() => {
        if (!playing) return;
        let last = performance.now();
        let lastSet = last;
        let frame = 0;
        const tick = (now: number) => {
            pos.current = Math.min(end, pos.current + (now - last) * rate);
            last = now;
            const atEnd = pos.current >= end;
            if (now - lastSet >= 100 || atEnd) {
                setPosition(pos.current);
                lastSet = now;
            }
            if (atEnd) {
                setPlaying(false);
                return;
            }
            frame = requestAnimationFrame(tick);
        };
        frame = requestAnimationFrame(tick);
        return () => cancelAnimationFrame(frame);
    }, [playing, rate, end]);

    const seek = useCallback((time: number) => {
        pos.current = Math.max(start, Math.min(end, time));
        setPosition(pos.current);
    }, [start, end]);

    const toggle = useCallback(() => {
        if (!playing && pos.current >= end) {
            pos.current = start;
            setPosition(start);
        }
        setPlaying(p => !p);
    }, [playing, start, end]);

    return {position, playing, rate, setRate, seek, toggle};
};

type EventFilter = 'all' | 'tests' | 'errors';

const shownBy = (filter: EventFilter) => (e: ReplayEvent): boolean => {
    switch (filter) {
        case 'tests':
            return e.kind !== 'log' || eventIsError(e);
        case 'errors':
            return eventIsError(e);
        default:
            return true;
    }
};

const EventIcon: React.FC<{ event: ReplayEvent }> = ({event}) => {
    const sx = {fontSize: 16, mt: '2px'};
    if (event.kind === 'test-start') return <PlayCircleOutlineRounded color="primary" sx={sx}/>;
    if (event.kind === 'test-end') {
        if (event.result === TestResultState.TestResultFailed) return <ErrorRounded color="error" sx={sx}/>;
        if (event.result === TestResultState.TestResultUnstable) return <WarningRounded color="warning" sx={sx}/>;
        return <CheckCircleRounded color="success" sx={sx}/>;
    }
    if (event.kind === 'step') return <TouchAppRounded color="action" sx={sx}/>;
    if (event.kind === 'screen') return <CameraAltRounded color="action" sx={sx}/>;
    if (eventIsError(event)) return <ErrorRounded color="error" sx={sx}/>;
    return <Box sx={{width: 16}}/>;
};

const eventText = (e: ReplayEvent): string => {
    switch (e.kind) {
        case 'test-start':
            return `Test started: ${e.text}`;
        case 'test-end':
            return `Test ${e.result === TestResultState.TestResultFailed ? 'failed' : e.result === TestResultState.TestResultUnstable ? 'unstable' : 'passed'}: ${e.text}`;
        default:
            return e.text;
    }
};

const EventRow = memo<{
    event: ReplayEvent, offset: number, active: boolean, future: boolean,
    onSeek: (time: number) => void, rowRef?: React.Ref<HTMLDivElement>,
}>(({event, offset, active, future, onSeek, rowRef}) => {
    const error = eventIsError(event);
    const test = event.kind === 'test-start' || event.kind === 'test-end';
    return <Box ref={rowRef} role="button" tabIndex={0} aria-current={active ? 'true' : undefined}
                onClick={() => onSeek(event.time)}
                onKeyDown={e => (e.key === 'Enter' || e.key === ' ') && (e.preventDefault(), onSeek(event.time))}
                sx={theme => ({
                    display: 'grid', gridTemplateColumns: '64px 18px 1fr', gap: 1, py: 0.5, px: 1, mx: -1,
                    cursor: 'pointer', borderRadius: 1, borderLeft: 3,
                    borderColor: active ? (theme.vars ?? theme).palette.primary.main : 'transparent',
                    bgcolor: active ? (theme.vars ?? theme).palette.action.selected : 'transparent',
                    opacity: future && !active ? 0.55 : 1,
                    '&:hover': {bgcolor: (theme.vars ?? theme).palette.action.hover},
                    // long logs: let the browser skip rows out of view
                    contentVisibility: 'auto', containIntrinsicSize: 'auto 26px',
                })}>
        <Tooltip title={formatDateTime(new Date(event.time), 'HH:mm:ss.SSS')} placement="left">
            <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>
                {formatOffset(offset)}
            </Typography>
        </Tooltip>
        <EventIcon event={event}/>
        <Box sx={{minWidth: 0}}>
            <Typography variant="body2" sx={{
                fontFamily: test ? undefined : 'monospace', fontSize: test ? undefined : 13, fontWeight: test ? 600 : undefined,
                wordBreak: 'break-word', color: error ? 'error.main' : 'text.primary',
            }}>
                {event.source && event.kind !== 'step' && <Typography component="span" variant="caption" color="text.secondary" sx={{mr: 1}}>
                    {event.source}
                </Typography>}
                {eventText(event)}
            </Typography>
            {event.image && <Box component="img" src={event.image} alt={event.text} loading="lazy"
                                 sx={{mt: 0.5, maxHeight: 120, maxWidth: '100%', borderRadius: 0.5, border: 1, borderColor: 'divider'}}/>}
        </Box>
    </Box>;
});

// Markers on the timeline: where tests ended (by result) and errors happened
const Markers: React.FC<{ events: ReplayEvent[], start: number, end: number }> = ({events, start, end}) => {
    const span = Math.max(1, end - start);
    const marks = useMemo(() => events.filter(e => e.kind === 'test-end' || eventIsError(e)).slice(0, 600), [events]);
    return <Box aria-hidden sx={{position: 'relative', height: 8, mx: '10px'}}>
        {marks.map(e => {
            const color = e.kind === 'test-end'
                ? (e.result === TestResultState.TestResultFailed ? 'error.main' : e.result === TestResultState.TestResultUnstable ? 'warning.main' : 'success.main')
                : 'error.main';
            return <Box key={e.key} sx={{
                position: 'absolute', left: `${((e.time - start) / span) * 100}%`, top: 0, width: 2, height: 8,
                bgcolor: color, borderRadius: 1, transform: 'translateX(-1px)',
            }}/>;
        })}
    </Box>;
};

interface TestReplayPageProps {
    run: ITestRunData,
}

const TestReplayPage: React.FC<TestReplayPageProps> = ({run}) => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state: hub} = useHubState();
    const [params, setParams] = useSearchParams();

    const segments = useMemo(() => segmentsOf(run.recordings), [run.recordings]);
    // devices with a recording, then those that only ran tests
    const deviceIds = useMemo(() => {
        const ids = [...segments.keys()];
        (run.protocols ?? []).forEach(p => p.deviceId && !ids.includes(p.deviceId) && ids.push(p.deviceId));
        return ids;
    }, [segments, run.protocols]);
    const requested = Number(params.get('device'));
    const deviceId = deviceIds.includes(requested) ? requested : deviceIds[0] ?? null;
    const deviceSegments = useMemo(() => (deviceId !== null ? segments.get(deviceId) : undefined) ?? [], [segments, deviceId]);

    const events = useMemo(() => buildTimeline(run, deviceId), [run, deviceId]);
    const range = useMemo(() => replayRange(deviceSegments, events), [deviceSegments, events]);
    const start = range?.start ?? 0;
    const end = range?.end ?? 0;
    const requestedTime = Number(params.get('t'));
    const initial = requestedTime > 0 ? Math.max(start, Math.min(end, requestedTime)) : (deviceSegments[0]?.start ?? start);
    const clock = useReplayClock(start, end, initial);
    const {position, playing, rate, setRate, seek, toggle} = clock;

    const [filter, setFilter] = useState<EventFilter>('all');
    const [follow, setFollow] = useState(true);
    const visible = useMemo(() => events.filter(shownBy(filter)), [events, filter]);
    const activeIndex = eventIndexAt(visible, position);
    const activeKey = activeIndex >= 0 ? visible[activeIndex].key : null;

    const list = useRef<HTMLDivElement>(null);
    const activeRow = useRef<HTMLDivElement>(null);
    useEffect(() => {
        if (!follow || !activeRow.current || !list.current) return;
        const box = list.current;
        const row = activeRow.current;
        const top = row.offsetTop - box.offsetTop;
        if (top < box.scrollTop || top > box.scrollTop + box.clientHeight - row.clientHeight) {
            box.scrollTo({top: Math.max(0, top - box.clientHeight / 3), behavior: playing ? 'smooth' : 'auto'});
        }
    }, [activeKey, follow, playing]);

    // the next test start or error after the position
    const nextMark = useMemo(() => events.find(e => e.time > position + 50 && (e.kind === 'test-start' || eventIsError(e))), [events, position]);

    useEffect(() => {
        const onKey = (e: KeyboardEvent) => {
            const target = e.target as HTMLElement;
            if (target.closest('input, textarea, [role="button"], button, [role="slider"]')) return;
            if (e.key === ' ') {
                e.preventDefault();
                toggle();
            } else if (e.key === 'ArrowLeft') {
                seek(position - 5000);
            } else if (e.key === 'ArrowRight') {
                seek(position + 5000);
            }
        };
        window.addEventListener('keydown', onKey);
        return () => window.removeEventListener('keydown', onKey);
    }, [toggle, seek, position]);

    const deviceName = (id: number) => {
        const d = hub.devices?.find(x => x.id === id) ?? run.protocols?.find(p => p.deviceId === id)?.device;
        return d ? (d.alias || d.name) : `Device ${id}`;
    };
    const pickDevice = (id: number) => setParams(p => {
        const next = new URLSearchParams(p);
        next.set('device', String(id));
        next.set('t', String(Math.round(position)));
        return next;
    }, {replace: true});

    const runUrl = `/project/${projectIdentifier}/app:${appId}/test/${run.testId}/run/${run.id}`;

    return <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
        <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
            <Box sx={{minWidth: 0}}>
                <Typography component="h1" variant="h5">
                    {run.test?.name} <Typography component="span" variant="h5" color="text.secondary">· Run {run.id} · Replay</Typography>
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{mt: 0.5}}>
                    Started {formatDateTime(run.createdAt, 'YYYY/MM/DD HH:mm')}
                </Typography>
            </Box>
            <Button size="small" startIcon={<ArrowBackRounded/>} component={RouterLink} to={runUrl} sx={{ml: {md: 'auto'}}}>
                Back to the run
            </Button>
        </Stack>

        {!range
            ? <Alert severity="info">This run has nothing to replay yet.</Alert>
            : <Grid container spacing={2}>
                <Grid size={{xs: 12, lg: 5}}>
                    <Card variant="outlined">
                        <CardContent>
                            {deviceIds.length > 1 && <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{mb: 1.5}}
                                                            role="group" aria-label="Device">
                                {deviceIds.map(id => <Chip key={id} size="small" label={deviceName(id)}
                                                           color={id === deviceId ? 'primary' : 'default'}
                                                           variant={id === deviceId ? 'filled' : 'outlined'}
                                                           onClick={() => pickDevice(id)}/>)}
                            </Stack>}
                            <ReplayVideo segments={deviceSegments} position={position} playing={playing} rate={rate}
                                         label={`Screen of ${deviceId !== null ? deviceName(deviceId) : 'the device'}`}/>

                            <Box sx={{mt: 1.5}}>
                                <Markers events={events} start={start} end={end}/>
                                <Slider size="small" min={0} max={Math.max(1, end - start)} value={position - start}
                                        onChange={(_, v) => seek(start + (v as number))}
                                        aria-label="Position in the run"
                                        valueLabelDisplay="auto" valueLabelFormat={v => formatOffset(v as number)}
                                        sx={{mt: -0.5}}/>
                            </Box>
                            <Stack direction="row" alignItems="center" spacing={0.5} flexWrap="wrap" useFlexGap>
                                <Tooltip title={playing ? 'Pause (space)' : 'Play (space)'}>
                                    <IconButton onClick={toggle} aria-label={playing ? 'Pause' : 'Play'} color="primary">
                                        {playing ? <PauseRounded/> : <PlayArrowRounded/>}
                                    </IconButton>
                                </Tooltip>
                                <Tooltip title="Back 5 s (←)">
                                    <IconButton size="small" onClick={() => seek(position - 5000)} aria-label="Back 5 seconds"><Replay5Rounded/></IconButton>
                                </Tooltip>
                                <Tooltip title="Ahead 5 s (→)">
                                    <IconButton size="small" onClick={() => seek(position + 5000)} aria-label="Ahead 5 seconds"><Forward5Rounded/></IconButton>
                                </Tooltip>
                                <Tooltip title={nextMark ? `Next test or error: ${eventText(nextMark)}` : 'No test or error after this point'}>
                                    <span>
                                        <IconButton size="small" disabled={!nextMark} onClick={() => nextMark && seek(nextMark.time)}
                                                    aria-label="Next test or error"><SkipNextRounded/></IconButton>
                                    </span>
                                </Tooltip>
                                <Typography variant="body2" sx={{fontFamily: 'monospace', mx: 1}}>
                                    {formatOffset(position - start)} / {formatOffset(end - start)}
                                </Typography>
                                <Select size="small" value={rate} onChange={e => setRate(Number(e.target.value))}
                                        inputProps={{'aria-label': 'Speed'}} sx={{ml: 'auto', minWidth: 80}}>
                                    {RATES.map(r => <MenuItem key={r} value={r}>{r}×</MenuItem>)}
                                </Select>
                            </Stack>
                            {deviceSegments.length === 0 && <Alert severity="info" variant="outlined" sx={{mt: 1}}>
                                This device has no recording. Turn on "Record the session" in the test settings to record the next runs.
                            </Alert>}
                        </CardContent>
                    </Card>
                </Grid>

                <Grid size={{xs: 12, lg: 7}}>
                    <Card variant="outlined">
                        <CardContent>
                            <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap" useFlexGap sx={{mb: 1}}>
                                <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600, mr: 1}}>What happened</Typography>
                                <Stack direction="row" spacing={1} role="group" aria-label="Show">
                                    {([['all', 'Everything'], ['tests', 'Tests and steps'], ['errors', 'Errors']] as [EventFilter, string][]).map(([f, label]) =>
                                        <Chip key={f} size="small" label={label}
                                              color={filter === f ? 'primary' : 'default'} variant={filter === f ? 'filled' : 'outlined'}
                                              onClick={() => setFilter(f)}/>)}
                                </Stack>
                                <Tooltip title="Keep the current line in view while the video plays">
                                    <Chip size="small" label="Follow the video" sx={{ml: 'auto'}}
                                          color={follow ? 'primary' : 'default'} variant={follow ? 'filled' : 'outlined'}
                                          onClick={() => setFollow(f => !f)}/>
                                </Tooltip>
                            </Stack>
                            {visible.length === 0
                                ? <Typography variant="body2" color="text.secondary">Nothing to show.</Typography>
                                : <Box ref={list} role="list" aria-label="What happened in the run"
                                       onWheel={() => follow && setFollow(false)} onTouchMove={() => follow && setFollow(false)}
                                       sx={{maxHeight: {xs: 480, lg: '75vh'}, overflowY: 'auto', position: 'relative', px: 1}}>
                                    {visible.map((e, i) =>
                                        <EventRow key={e.key} event={e} offset={e.time - start} active={i === activeIndex}
                                                  future={e.time > position} onSeek={seek}
                                                  rowRef={i === activeIndex ? activeRow : undefined}/>)}
                                </Box>}
                        </CardContent>
                    </Card>
                </Grid>
            </Grid>}
    </Box>;
};

export const TestReplayPageLoader: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {testId, runId} = useParams();
    const {data, error, retry} = useLoader(
        () => getRun(projectIdentifier, appId as number, testId as string, runId as string),
        [projectIdentifier, appId, testId, runId]);
    if (!data || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    return <TestReplayPage key={data.testRun.id} run={data.testRun}/>;
};

export default TestReplayPage;
