import React, {useMemo, useState} from 'react';

import {
    Box,
    Button,
    IconButton,
    Paper,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Typography,
} from '@mui/material';
import ITestRunData from '../../types/test.run';
import {TestResultState} from '../../types/test.result.state.enum';
import DateTime from "../../components/DateTime";
import {useSSE} from 'react-hooks-sse';
import ITesRunLogEntryData from '../../types/test.run.log.entry';
import ITestProtocolData from '../../types/test.protocol';
import {cancelTestRun, executeTest} from '../../services/test.service';
import {useNavigate} from 'react-router-dom';
import {
    Cancel,
    CheckCircle,
    DirectionsRun,
    KeyboardArrowDown,
    KeyboardArrowLeft,
    KeyboardArrowRight,
    KeyboardArrowUp,
    Warning
} from '@mui/icons-material';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import {Bar, BarChart, CartesianGrid, Cell, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis} from 'recharts';
import {useApplicationContext} from "../../hooks/ApplicationProvider";
import Grid from "@mui/material/Grid";
import TestStatusIconComponent from "../../components/test-status-icon.component";
import Collapse from "@mui/material/Collapse";
import Link from "@mui/material/Link";
import {useError} from "../../ErrorProvider";

interface TestRunPageProps {
    testRun: ITestRunData
    nextRunId: number | null
    prevRunId: number | null
}

interface NewTestRunPayload {
    testRunId: number,
    entry: ITesRunLogEntryData,
}

const runStyles = {
    chip: {
        '& .chip--error': {
            backgroundColor: 'red',
            color: '#ffffff',
            margin: '5px',
        },
        '& .chip--error--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: 'red',
            margin: '5px',
        },
        '& .chip--success': {
            backgroundColor: 'green',
            color: '#ffffff',
            margin: '5px',
        },
        '& .chip--success--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: 'green',
            margin: '5px',
        },
        '& .chip--pending': {
            backgroundColor: '#FFC857',
            color: '#000000',
            margin: '5px',
        },
        '& .chip--pending--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: '#FFC857',
            margin: '5px',
        },
    },
};

type FilterType = {
    Success: boolean,
    Failed: boolean,
    Unstable: boolean,
    Pending: boolean,
};

const isTopLevel = (p: ITestProtocolData): boolean =>
    p.parentTestProtocolId === null || p.parentTestProtocolId === undefined || p.parentTestProtocolId === 0;

// mergeProtocol upserts a protocol update into the flat list. Updates are
// published concurrently, so a late "open" update must not overwrite a
// finished protocol, and updates usually don't carry the preloaded device.
const mergeProtocol = (protocols: ITestProtocolData[], update: ITestProtocolData | null): ITestProtocolData[] => {
    if (!update || update.id === undefined || update.id === null) {
        return protocols;
    }
    const index = protocols.findIndex(p => p.id === update.id);
    if (index < 0) {
        return [...protocols, update];
    }
    const existing = protocols[index];
    const merged: ITestProtocolData = {...existing, ...update, device: update.device ?? existing.device};
    if (existing.endedAt && !update.endedAt) {
        merged.endedAt = existing.endedAt;
        merged.testResult = existing.testResult;
    }
    const next = [...protocols];
    next[index] = merged;
    return next;
};

// groupProtocols nests child protocols under their parent. Children whose
// parent hasn't arrived yet stay in the flat list and show up with it.
const groupProtocols = (protocols: ITestProtocolData[]): ITestProtocolData[] => {
    const parents = new Map<number, ITestProtocolData>();
    protocols.filter(isTopLevel).forEach(p => parents.set(p.id as number, {...p, childProtocols: []}));
    protocols.filter(p => !isTopLevel(p)).forEach(p => {
        parents.get(p.parentTestProtocolId as number)?.childProtocols.push(p);
    });
    return Array.from(parents.values());
};

const matchesFilter = (filter: FilterType, p: ITestProtocolData): boolean =>
    (filter.Success && p.testResult === TestResultState.TestResultSuccess)
    || (filter.Failed && p.testResult === TestResultState.TestResultFailed)
    || (filter.Unstable && p.testResult === TestResultState.TestResultUnstable)
    || (filter.Pending && p.testResult === TestResultState.TestResultOpen);

// applyFilter keeps a parent when it or one of its children matches, and
// filters its children the same way.
const applyFilter = (filter: FilterType, grouped: ITestProtocolData[]): ITestProtocolData[] =>
    grouped
        .map(p => ({...p, childProtocols: p.childProtocols.filter(c => matchesFilter(filter, c))}))
        .filter(p => matchesFilter(filter, p) || p.childProtocols.length > 0);

// statistics count the executed tests: the children of a protocol, or the
// protocol itself when it has none.
const statistics = (grouped: ITestProtocolData[]) => {
    const stats = {open: 0, failed: 0, unstable: 0, success: 0};
    const count = (p: ITestProtocolData) => {
        switch (p.testResult) {
            case TestResultState.TestResultOpen:
                stats.open++;
                break;
            case TestResultState.TestResultFailed:
                stats.failed++;
                break;
            case TestResultState.TestResultUnstable:
                stats.unstable++;
                break;
            case TestResultState.TestResultSuccess:
                stats.success++;
                break;
        }
    };
    grouped.forEach(p => p.childProtocols.length > 0 ? p.childProtocols.forEach(count) : count(p));
    return stats;
};

const deviceName = (p: ITestProtocolData) => p.device && (p.device.alias.length > 0 ? p.device.alias : p.device.name);

const TestProtocolRow: React.FC<{ row: ITestProtocolData, onOpen: (protocolId: number) => void }> = ({row, onOpen}) => {
    const [open, setOpen] = React.useState(false);
    const hasChildren = row.childProtocols.length > 0;

    return (
        <React.Fragment>
            <TableRow sx={{'& > *': {borderBottom: 'unset'}}}>
                <TableCell>
                    {hasChildren && <IconButton aria-label="expand row" size="small" onClick={() => setOpen(!open)}>
                        {open ? <KeyboardArrowUp/> : <KeyboardArrowDown/>}
                    </IconButton>}
                </TableCell>
                <TableCell component="th" scope="row">
                    <Link onClick={() => onOpen(row.id as number)} underline="none" sx={{cursor: 'pointer'}}>
                        {row.testName}
                    </Link>
                    {hasChildren && <Typography variant="caption" sx={{ml: 1}}>({row.childProtocols.length})</Typography>}
                </TableCell>
                <TableCell>
                    <Grid container={true}>
                        <Grid size={12}>{deviceName(row)}</Grid>
                        <Grid size={12}>{row.device?.os} {row.device?.osVersion}</Grid>
                    </Grid>
                </TableCell>
                <TableCell align="right"><TestStatusIconComponent status={row.testResult}/></TableCell>
            </TableRow>
            {hasChildren && <TableRow>
                <TableCell style={{padding: 0}} colSpan={6}>
                    <Collapse in={open} timeout="auto" unmountOnExit>
                        <Box sx={{margin: 0}}>
                            <Table size="small">
                                <TableHead>
                                    <TableRow>
                                        <TableCell></TableCell>
                                        <TableCell>Test</TableCell>
                                        <TableCell>Device</TableCell>
                                        <TableCell align="right">Result</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {row.childProtocols.map((child) => {
                                        const childName = child.testName.split("/");
                                        return (
                                            <TableRow key={child.id}>
                                                <TableCell component="th" scope="row"/>
                                                <TableCell>
                                                    <Link onClick={() => onOpen(child.id as number)} underline="none" sx={{cursor: 'pointer'}}>
                                                        {childName.length > 1 ? childName[1] : childName[0]}
                                                    </Link>
                                                </TableCell>
                                                <TableCell>
                                                    <Grid container={true}>
                                                        <Grid size={12}>{deviceName(child)}</Grid>
                                                        <Grid size={12}>{child.device?.os} {child.device?.osVersion}</Grid>
                                                    </Grid>
                                                </TableCell>
                                                <TableCell align="right"><TestStatusIconComponent status={child.testResult}/></TableCell>
                                            </TableRow>
                                        )
                                    })}
                                </TableBody>
                            </Table>
                        </Box>
                    </Collapse>
                </TableCell>
            </TableRow>}
        </React.Fragment>
    );
};

// TestRunPage shows one run. The loader mounts it with key={testRun.id}, so
// all state below belongs to exactly one run.
const TestRunPage: React.FC<TestRunPageProps> = (props: TestRunPageProps) => {

    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {testRun, nextRunId, prevRunId} = props;
    const {setError} = useError()
    const navigate = useNavigate();

    // live protocol and log updates are folded into state by the reducers, so
    // events arriving in the same render are not lost
    const protocols = useSSE<ITestProtocolData[], ITestProtocolData | null>(
        `test_run_${testRun.id}_protocol`, testRun.protocols ?? [],
        {stateReducer: (state, action) => mergeProtocol(state, action.data)});
    const log = useSSE<ITesRunLogEntryData[], NewTestRunPayload | null>(
        `test_run_${testRun.id}_log`, testRun.log ?? [],
        {stateReducer: (state, action) => action.data?.entry && !state.some(e => e.id === action.data!.entry.id) ? [...state, action.data.entry] : state});

    const [filter, setFilter] = useState<FilterType>({
        Success: true,
        Failed: true,
        Unstable: true,
        Pending: true,
    });

    const grouped = useMemo(() => groupProtocols(protocols), [protocols]);
    const visibleProtocols = useMemo(() => applyFilter(filter, grouped), [filter, grouped]);
    const stats = useMemo(() => statistics(grouped), [grouped]);

    const runUrl = (runId: number | null | undefined) => `/project/${projectIdentifier}/app:${appId}/test/${testRun.testId}/run/${runId}`;
    const openProtocol = (protocolId: number) => navigate(`${runUrl(testRun.id)}/${protocolId}`);

    const onTestRerun = (): void => {
        executeTest(projectIdentifier, appId, testRun.testId, {
            appBinaryId: testRun.appBinaryId,
            params: testRun.parameter,
            startUrl: testRun.startUrl,
        }).then(run => {
            navigate(runUrl(run.id));
        }).catch(error => {
            setError(error);
        });
    };

    const onCancelTestRun = (): void => {
        cancelTestRun(projectIdentifier, appId, testRun.testId, testRun.id!).catch(error => {
            setError(error);
        });
    };

    const startupTimes = testRun.deviceStatus.map(item => ({
        Name: item.device && (item.device.alias.length > 0 ? item.device.alias : item.device.name),
        StartupTime: item.startupTime
    }));

    const environmentParameters = testRun.parameter ? testRun.parameter.split(";") : [];

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={`Test: ${testRun.test.name} Run: ${testRun.id}`}>
                { /*test Previous and Next Navigation*/}
                <Grid container={true} spacing={2}>
                    <Grid size={{xs: 6}} container={true} sx={{
                        padding: 2,
                        borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                    }}>
                        {prevRunId !== null && prevRunId > 0 &&
                            <Button variant="contained" color="primary" size="small"
                                    onClick={() => navigate(runUrl(prevRunId))}>
                                <KeyboardArrowLeft/> Prev
                            </Button>
                        }
                    </Grid>
                    <Grid size={{xs: 6}} container={true} justifyContent={"flex-end"} sx={{
                        padding: 1,
                        borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                    }}>
                        {nextRunId !== null && nextRunId > 0 &&
                            <Button variant="contained" color="primary" size="small"
                                    onClick={() => navigate(runUrl(nextRunId))}>
                                Next <KeyboardArrowRight/>
                            </Button>
                        }
                    </Grid>
                </Grid>
            </TitleCard>
            <TitleCard title={'Environment'}>
                <Grid container={true}>
                    {environmentParameters.map((e, i) => (
                        <Grid size={12} key={`env_param_${i}`}>
                            <Typography key={`env_${i}`} variant={"body1"}>
                                {e}
                            </Typography>
                        </Grid>
                    ))}
                </Grid>
                <Grid sx={{flexGrow: 1}}></Grid>
            </TitleCard>
            {
                testRun.appBinary &&
                <TitleCard title={"App Bundle"}>
                    <Box sx={{p: 1, m: 1}}>
                        <Grid container={true}>
                            <Grid size={{xs: 12, md: 2}}>
                                Name:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun.appBinary?.name}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Identifier:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun.appBinary?.identifier}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Platform:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun.appBinary?.platform}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Version:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun.appBinary?.version}
                            </Grid>
                            <Grid size={{xs: 12, md: 2}}>
                                Hash:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun.appBinary?.hash}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Created:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                <DateTime format="YYYY/MM/DD HH:mm:ss" value={testRun.appBinary?.createdAt}/>
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Addons:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun.appBinary?.additional}
                            </Grid>
                        </Grid>
                    </Box>
                </TitleCard>
            }
            <TitleCard title={"Results"}>
                <Grid container={true} alignItems={"center"}
                      justifyContent={"center"} spacing={2}>
                    <PieChart
                        width={200}
                        height={200}
                    >
                        <Pie
                            data={[
                                {name: "Open", value: stats.open},
                                {name: "Failed", value: stats.failed},
                                {name: "Unstable", value: stats.unstable},
                                {name: "Success", value: stats.success},
                            ]}
                            cx="50%"
                            cy="50%"
                            label
                            outerRadius={70}
                            fill="#8884d8"
                            dataKey="value"
                        >
                            <Cell fill={'yellow'}/>
                            <Cell fill={'red'}/>
                            <Cell fill={'orange'}/>
                            <Cell fill={'green'}/>
                        </Pie>
                        <Tooltip/>
                    </PieChart>
                    <Grid container={true}>
                        <Grid size={12}>
                            <Typography
                                variant={"caption"}>Open: {stats.open}</Typography>
                        </Grid>
                        <Grid size={12}>
                            <Typography
                                variant={"caption"}>Failed: {stats.failed}</Typography>
                        </Grid>
                        <Grid size={12}>
                            <Typography
                                variant={"caption"}>Unstable: {stats.unstable}</Typography>
                        </Grid>
                        <Grid size={12}>
                            <Typography
                                variant={"caption"}>Success: {stats.success}</Typography>
                        </Grid>
                    </Grid>
                </Grid>
                <Grid container={true} alignItems={"flex-end"}
                      justifyContent={"flex-end"} spacing={2}>
                    <Grid>
                        <Button variant="contained" color="primary" onClick={onTestRerun}>
                            Rerun
                        </Button>
                    </Grid>
                    <Grid>
                        {stats.open > 0 &&
                            <Button variant="contained" color="secondary"
                                    onClick={onCancelTestRun}>
                                Cancel
                            </Button>}
                    </Grid>
                </Grid>
            </TitleCard>
            <TitleCard title={"Test Functions"}>
                <Grid container justifyContent="flex-end" sx={runStyles.chip} padding={1}>
                    <CheckCircle htmlColor={filter.Success ? 'green' : 'lightgray'}
                                 onClick={() => setFilter(prevState => ({
                                     ...prevState,
                                     Success: !prevState.Success
                                 }))}
                    />
                    <Cancel htmlColor={filter.Failed ? 'red' : 'lightgray'}
                            onClick={() => setFilter(prevState => ({...prevState, Failed: !prevState.Failed}))}
                    />
                    <Warning htmlColor={filter.Unstable ? 'orange' : 'lightgray'}
                             onClick={() => setFilter(prevState => ({
                                 ...prevState,
                                 Unstable: !prevState.Unstable
                             }))}
                    />
                    <DirectionsRun htmlColor={filter.Pending ? 'yellow' : 'lightgray'}
                                   onClick={() => setFilter(prevState => ({
                                       ...prevState,
                                       Pending: !prevState.Pending
                                   }))}
                    />
                </Grid>
                <Grid container={true}>
                    <TableContainer component={Paper}>
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell></TableCell>
                                    <TableCell>Test</TableCell>
                                    <TableCell>Device</TableCell>
                                    <TableCell align="right">Status</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {visibleProtocols.map((protocol) => (
                                    <TestProtocolRow key={`protocol_${protocol.id}`} row={protocol} onOpen={openProtocol}/>
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>
                </Grid>
            </TitleCard>
            <TitleCard title={"App Startup Time"}>
                <Paper sx={{margin: 'auto', overflow: 'hidden'}}>
                    <ResponsiveContainer width={'100%'} height={220}>
                        <BarChart width={600} height={200} data={startupTimes} margin={{
                            top: 20,
                            right: 30,
                            left: 20,
                            bottom: 5,
                        }}>
                            <CartesianGrid strokeDasharray="3 3"/>
                            <XAxis dataKey="Name"/>
                            <Tooltip/>
                            <Bar dataKey="StartupTime" fill="#8884d8" label={{position: 'top'}} unit={'ms'}/>
                        </BarChart>
                    </ResponsiveContainer>
                </Paper>
            </TitleCard>
            <TitleCard title={"Execution log"}>
                <TableContainer component={Paper}>
                    <Table size="small" aria-label="a dense table">
                        <TableHead>
                            <TableRow>
                                <TableCell>Date</TableCell>
                                <TableCell>Level</TableCell>
                                <TableCell>Log</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {log.map((entry) => <TableRow key={entry.id}>
                                <TableCell component="th" scope="row" style={{whiteSpace: 'nowrap'}}>
                                    <DateTime format="YYYY/MM/DD HH:mm:ss" value={entry.createdAt}/>
                                </TableCell>
                                <TableCell>{entry.level}</TableCell>
                                <TableCell>{entry.log}</TableCell>
                            </TableRow>)}
                        </TableBody>
                    </Table>
                </TableContainer>
            </TitleCard>
        </Box>
    );
};

export default TestRunPage;
