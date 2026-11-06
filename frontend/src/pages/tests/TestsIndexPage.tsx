import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {
    Alert,
    Box,
    Button,
    ButtonBase,
    Card,
    CardContent,
    Chip,
    CircularProgress,
    InputAdornment,
    LinearProgress,
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
import {useTheme} from '@mui/material/styles';
import {AddRounded, PlayArrowRounded, SearchRounded} from '@mui/icons-material';
import {Link as RouterLink, useNavigate} from 'react-router-dom';
import ITestData, {IRunSummary} from '../../types/test';
import {getAllTests} from '../../services/test.service';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {formatDateTime} from '../../components/DateTime';
import RunTestDialog from '../../components/RunTestDialog';
import {ago} from '../apps/bundlesModel';
import {countHealth, describeConfig, filterTests, HealthFilter, STATUS, Tone} from './testsListModel';

const FILTERS: { key: HealthFilter, label: string }[] = [
    {key: 'all', label: 'All'},
    {key: 'failing', label: 'Failing'},
    {key: 'passing', label: 'Passing'},
    {key: 'running', label: 'Running'},
    {key: 'never', label: 'Never ran'},
];

const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};

// History shows the latest runs oldest to newest, one square each; a square
// opens its run. Status is in the tooltip and the label, not only the color.
const History: React.FC<{ runs: IRunSummary[], runUrl: (id: number) => string }> = ({runs, runUrl}) => {
    const theme = useTheme();
    const color = (tone: Tone) => tone === 'default' ? theme.palette.action.disabled : (theme.vars ?? theme).palette[tone].main;
    return <Stack direction="row" spacing={0.5} role="list" aria-label="Latest runs, oldest first">
        {[...runs].reverse().map(r => {
            const s = STATUS[r.status];
            const counts = [r.passed && `${r.passed} passed`, r.failed && `${r.failed} failed`, r.unstable && `${r.unstable} unstable`]
                .filter(Boolean).join(', ');
            return <Tooltip key={r.id} title={<>Run {r.id}: {s.label}{counts && ` (${counts})`}<br/>{formatDateTime(r.createdAt, 'YYYY/MM/DD HH:mm')}</>}>
                <ButtonBase component={RouterLink} to={runUrl(r.id)} role="listitem" aria-label={`Run ${r.id}: ${s.label}`}
                            onClick={e => e.stopPropagation()}
                            sx={{width: 12, height: 20, borderRadius: 0.5, bgcolor: color(s.tone)}}/>
            </Tooltip>;
        })}
    </Stack>;
};

// TestsIndexPage lists the tests of the app picked in the menu with how
// they are doing, and starts them.
const TestsIndexPage: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const navigate = useNavigate();
    const app = state.apps?.find(a => a.id === appId);

    const [tests, setTests] = useState<ITestData[] | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [filter, setFilter] = useState<HealthFilter>('all');
    const [search, setSearch] = useState('');
    const [running, setRunning] = useState<ITestData | null>(null);
    const now = Date.now();

    const base = `/project/${projectIdentifier}/app:${appId}`;
    const load = useCallback(() => {
        if (appId === null) return;
        setError(null);
        getAllTests(projectIdentifier, appId).then(setTests).catch(e => setError(messageOf(e)));
    }, [projectIdentifier, appId]);
    useEffect(load, [load]);

    const counts = useMemo(() => countHealth(tests ?? []), [tests]);
    // older masters don't send the latest runs; then nothing is known about them
    const historyKnown = !!tests && tests.some(t => t.recentRuns !== undefined);
    const visible = useMemo(() => filterTests(tests ?? [], filter, search), [tests, filter, search]);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Typography component="h1" variant="h5">Tests</Typography>
                    <Typography variant="body2" color="text.secondary">
                        Test setups of {app?.name ?? 'this app'}: what runs, on which devices. Open one for its runs, or start it.
                    </Typography>
                </Box>
                <Button variant="contained" startIcon={<AddRounded/>} sx={{ml: {md: 'auto'}}}
                        component={RouterLink} to={`${base}/test/new`}>New test</Button>
            </Stack>

            <Card variant="outlined">
                <CardContent>
                    {error
                        ? <Alert severity="error" action={<Button color="inherit" size="small" onClick={load}>Retry</Button>}>{error}</Alert>
                        : tests === null
                            ? <LinearProgress/>
                            : tests.length === 0
                                ? <Stack alignItems="center" spacing={1.5} sx={{py: 6, textAlign: 'center'}}>
                                    <Typography variant="subtitle1">No tests yet</Typography>
                                    <Typography variant="body2" color="text.secondary">
                                        A test says what to run (for Unity: all tests, a category or picked tests) and on which devices.
                                    </Typography>
                                    <Button variant="outlined" startIcon={<AddRounded/>} component={RouterLink} to={`${base}/test/new`}>Create a test</Button>
                                </Stack>
                                : <>
                                    {(historyKnown || tests.length > 5) && <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap sx={{mb: 1.5}}>
                                        {historyKnown && <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap role="group" aria-label="Show tests by how they are doing">
                                            {FILTERS.filter(f => f.key === 'all' || counts[f.key] > 0).map(f =>
                                                <Chip key={f.key} size="small" label={`${f.label} ${counts[f.key]}`}
                                                      color={filter === f.key ? (f.key === 'failing' ? 'error' : 'primary') : 'default'}
                                                      variant={filter === f.key ? 'filled' : 'outlined'}
                                                      onClick={() => setFilter(filter === f.key ? 'all' : f.key)}/>)}
                                        </Stack>}
                                        {tests.length > 5 && <TextField size="small" placeholder="Find a test" value={search}
                                                                        onChange={e => setSearch(e.target.value)}
                                                                        sx={{ml: {sm: 'auto'}, width: {xs: '100%', sm: 240}}}
                                                                        slotProps={{
                                                                            input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                                                                            htmlInput: {'aria-label': 'Find a test'},
                                                                        }}/>}
                                    </Stack>}
                                    {visible.length === 0
                                        ? <Typography variant="body2" color="text.secondary" sx={{py: 2}}>No test matches.</Typography>
                                        : <TableContainer sx={{mx: -2, width: 'auto'}}>
                                            <Table size="small" aria-label="Tests">
                                                <TableHead>
                                                    <TableRow>
                                                        <TableCell sx={{pl: 2}}>Test</TableCell>
                                                        <TableCell>Last run</TableCell>
                                                        <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>Latest runs</TableCell>
                                                        <TableCell/>
                                                    </TableRow>
                                                </TableHead>
                                                <TableBody>
                                                    {visible.map(t => {
                                                        const last = t.recentRuns?.[0];
                                                        const s = last ? STATUS[last.status] : null;
                                                        const testUrl = `${base}/test/${t.id}`;
                                                        return <TableRow key={t.id} hover sx={{cursor: 'pointer'}} onClick={() => navigate(testUrl)}>
                                                            <TableCell sx={{pl: 2}}>
                                                                <Typography variant="body2" sx={{fontWeight: 600}}>{t.name}</Typography>
                                                                <Typography variant="caption" color="text.secondary">{describeConfig(t.testConfig)}</Typography>
                                                            </TableCell>
                                                            <TableCell sx={{whiteSpace: 'nowrap'}}>
                                                                {last && s
                                                                    ? <Stack direction="row" spacing={1} alignItems="center">
                                                                        <Chip size="small" color={s.tone} label={s.label}
                                                                              icon={last.status === 'running' ? <CircularProgress size={12} color="inherit"/> : undefined}
                                                                              component={RouterLink} to={`${testUrl}/run/${last.id}`} clickable
                                                                              onClick={(e: React.MouseEvent) => e.stopPropagation()}/>
                                                                        <Tooltip title={formatDateTime(last.createdAt)}>
                                                                            <Typography variant="body2" color="text.secondary">{ago(last.createdAt, now)}</Typography>
                                                                        </Tooltip>
                                                                    </Stack>
                                                                    : <Typography variant="body2" color="text.secondary">{historyKnown ? 'Never ran' : '–'}</Typography>}
                                                            </TableCell>
                                                            <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>
                                                                {t.recentRuns && t.recentRuns.length > 0
                                                                    ? <History runs={t.recentRuns} runUrl={id => `${testUrl}/run/${id}`}/>
                                                                    : <Typography variant="body2" color="text.secondary">–</Typography>}
                                                            </TableCell>
                                                            <TableCell align="right" sx={{pr: 2}}>
                                                                <Button size="small" variant="outlined" startIcon={<PlayArrowRounded/>} disabled={!app}
                                                                        onClick={e => {
                                                                            e.stopPropagation();
                                                                            setRunning(t);
                                                                        }}>Run</Button>
                                                            </TableCell>
                                                        </TableRow>;
                                                    })}
                                                </TableBody>
                                            </Table>
                                        </TableContainer>}
                                </>}
                </CardContent>
            </Card>

            {app && <RunTestDialog projectIdentifier={projectIdentifier} app={app} test={running} onClose={() => setRunning(null)}/>}
        </Box>
    );
};

export default TestsIndexPage;
