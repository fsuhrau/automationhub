import React, {useContext, useEffect, useState} from 'react';

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
import {TestContext} from '../../context/test.context';
import {
    Cancel,
    CheckCircle,
    DirectionsRun,
    KeyboardArrowDown,
    KeyboardArrowLeft,
    KeyboardArrowRight,
    KeyboardArrowUp
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

const TestRunPage: React.FC<TestRunPageProps> = (props: TestRunPageProps) => {

    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();

    const {testRun, nextRunId, prevRunId} = props;

    const {setError} = useError()

    const testContext = useContext(TestContext);
    const {test, setTest} = testContext;

    const navigate = useNavigate();

    const [state, setState] = useState<{
        currentRunID: number,
        currentTestRun: ITestRunData;
        nextRunId: number | null;
        prevRunId: number | null;
        log: ITesRunLogEntryData[],
        protocols: ITestProtocolData[],
        runsOpen: number,
        runsFailed: number,
        runsSuccess: number
    }>({
        currentRunID: testRun.id as number,
        currentTestRun: testRun,
        nextRunId: nextRunId,
        prevRunId: prevRunId,
        log: [],
        protocols: testRun.protocols ? testRun.protocols : [],
        runsOpen: 0,
        runsSuccess: 0,
        runsFailed: 0
    })

    const patchProtocols = (parentId: null | number, protocols: ITestProtocolData[], protocol: ITestProtocolData): ITestProtocolData[] => {
        if (parentId !== null) {
            return protocols.map(p => p.id === protocol.parentTestProtocolId ? {
                ...p,
                childProtocols: patchProtocols(null, p.childProtocols, protocol)
            } : p);
        }
        if (protocols.findIndex(p => p.id === protocol.id) >= 0)
            return protocols.map(p => p.id === protocol.id ? protocol : p);
        return [...protocols, protocol]
    }

    console.log(`listen to protocol test_run_${state.currentRunID}_protocol`)
    const testProtocolEntry = useSSE<ITestProtocolData | null>(`test_run_${state.currentRunID}_protocol`, null)
    useEffect(() => {
        console.log("Received new protocol entry: " + JSON.stringify(testProtocolEntry))
        if (testProtocolEntry === null) return;
        setState(prevState => ({
            ...prevState,
            protocols: [...patchProtocols(testProtocolEntry.parentTestProtocolId, prevState.protocols, testProtocolEntry)],
        }));
    }, [testProtocolEntry]);

    useEffect(() => {
        if (state.currentTestRun == null) {
            return;
        }

        console.log("testRun changed to: " + state.currentTestRun.id)
        setState(prevState => ({...prevState, currentRunID: state.currentTestRun.id as number}));
    }, [state.currentTestRun.id]);

    /*
    const testRunEntry = useSSE<NewTestRunPayload | null>(`test_run_${state.currentTestRun.id}_log`, null);
    useEffect(() => {
        if (testRunEntry === null || testRunEntry.entry == null)
            return;
        console.log("Received test Run log:  " + testRunEntry.entry.id)
        setState(prevState => ({...prevState, log: [...prevState.log, testRunEntry.entry]}))
    }, [testRunEntry]);
*/
    type FilterType = {
        Success: boolean,
        Failed: boolean,
        Pending: boolean,
    };

    const [filter, setFilter] = useState<FilterType>({
        Success: true,
        Failed: true,
        Pending: true,
    });

    const rebuildStatistics = (protocols: ITestProtocolData[]): { ro: number, rf: number, rs: number } => {
        let ro: number;
        let rf: number;
        let rs: number;
        ro = 0;
        rf = 0;
        rs = 0;

        protocols.forEach(value => {
            switch (value.testResult) {
                case TestResultState.TestResultOpen:
                    ro++;
                    break;
                case TestResultState.TestResultFailed:
                    rf++;
                    break;
                case TestResultState.TestResultSuccess:
                    rs++;
                    break;
            }
        });
        return {ro: ro, rf: rf, rs: rs}
    }

    useEffect(() => {
        let {ro, rf, rs} = rebuildStatistics(state.protocols);
        setState(prevState => ({
            ...prevState,
            log: state.currentTestRun.log,
            runsOpen: ro,
            runsFailed: rf,
            runsSuccess: rs,
            // protocols: state.protocols
        }))
    }, [state.protocols]);

    const onTestRerun = (): void => {
        executeTest(projectIdentifier, appId, state.currentTestRun.testId, {
            appBinaryId: state.currentTestRun.appBinaryId,
            params: state.currentTestRun.parameter,
            startUrl: state.currentTestRun.startUrl,
        }).then(test => {
            //navigate(          `/project/${projectIdentifier}/app:${appId}/test/${testRun.testId}/run/${test.id}`);
            window.location.href = `/project/${projectIdentifier}/app:${appId}/test/${state.currentTestRun.testId}/run/${test.id}`;
        }).catch(error => {
            setError(error);
        });
    };

    const onCancelTestRun = (): void => {
        cancelTestRun(projectIdentifier, appId, state.currentTestRun.testId, state.currentTestRun.id!).then(response => {
            console.log("test run cancelled", response);
        }).catch(error => {
            setError(error);
        });
    };

    const startupTimes = state.currentTestRun.deviceStatus.map(item => {
            return {
                Name: item.device && (item.device.alias.length > 0 ? item.device.alias : item.device.name),
                StartupTime: item.startupTime
            }
        }
    )

    const environmentParameters = state.currentTestRun?.parameter.split(";");

    const applyProtocolFilter = (value: ITestProtocolData): boolean => {
        return (filter.Success && value.testResult === TestResultState.TestResultSuccess)
            || (filter.Failed && value.testResult === TestResultState.TestResultFailed)
            || (filter.Pending && value.testResult === TestResultState.TestResultOpen)
            || value.testResult === TestResultState.TestResultUnstable;
    }

    const groupProtocols = (protocols: ITestProtocolData[]): ITestProtocolData[] => {
        const protocolMap: { [key: number]: ITestProtocolData } = {};

        protocols.forEach(protocol => {
            if (protocol.parentTestProtocolId === null || protocol.parentTestProtocolId === undefined || protocol.parentTestProtocolId === 0) {
                protocolMap[protocol.id!] = {...protocol, childProtocols: []};
            } else {
                const parent = protocolMap[protocol.parentTestProtocolId];
                if (parent) {
                    parent.childProtocols.push(protocol);
                }
            }
        });
        return Object.values(protocolMap);
    };

    function TestProtocolRow(props: { row: ITestProtocolData }) {
        const {row} = props;
        const [open, setOpen] = React.useState(false);
        const testName = row.testName.split("/");

        return (
            <React.Fragment>
                <TableRow sx={{'& > *': {borderBottom: 'unset'}}}>
                    <TableCell>
                        <IconButton
                            aria-label="expand row"
                            size="small"
                            onClick={() => setOpen(!open)}
                        >
                            {open ? <KeyboardArrowUp/> : <KeyboardArrowDown/>}
                        </IconButton>
                    </TableCell>
                    <TableCell component="th" scope="row">
                        <Link
                            onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/${state.currentTestRun.testId}/run/${state.currentTestRun.id}/${row.id}`)}
                            underline="none">
                            {testName}
                        </Link>
                    </TableCell>
                    <TableCell>
                        <Grid container={true}>
                            <Grid size={{xs: 12, md: 12}}>
                                {row.device && (row.device.alias.length > 0 ? row.device.alias : row.device.name)}
                            </Grid>
                            <Grid size={{xs: 12, md: 12}}>
                                {row.device?.os} {row.device?.osVersion}
                            </Grid>
                        </Grid>
                    </TableCell>
                    <TableCell align="right"><TestStatusIconComponent status={row.testResult}/></TableCell>
                </TableRow>
                <TableRow>
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
                                        {row.childProtocols?.map((historyRow) => {
                                            const historyTestName = historyRow.testName.split("/");

                                            return (
                                                <TableRow key={historyRow.id}>
                                                    <TableCell component="th" scope="row">
                                                    </TableCell>
                                                    <TableCell>
                                                        <Link
                                                            onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/${state.currentTestRun.testId}/run/${state.currentTestRun.id}/${historyRow.id}`)}
                                                            underline="none">
                                                            {historyTestName.length > 1 ? historyTestName[1] : historyTestName[0]}
                                                        </Link>
                                                    </TableCell>
                                                    <TableCell>
                                                        <Grid container={true}>
                                                            <Grid size={{xs: 12, md: 12}}>
                                                                {historyRow.device && (historyRow.device.alias.length > 0 ? historyRow.device.alias : historyRow.device.name)}
                                                            </Grid>
                                                            <Grid size={{xs: 12, md: 12}}>
                                                                {historyRow.device?.os} {historyRow.device?.osVersion}
                                                            </Grid>
                                                        </Grid>
                                                    </TableCell>
                                                    <TableCell align="right"><TestStatusIconComponent
                                                        status={historyRow.testResult}/></TableCell>
                                                </TableRow>
                                            )
                                        })}
                                    </TableBody>
                                </Table>
                            </Box>
                        </Collapse>
                    </TableCell>
                </TableRow>
            </React.Fragment>
        );
    }

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={`Test: ${state.currentTestRun.test.name} Run: ${state.currentTestRun.id}`}>
                { /*test Previous and Next Navigation*/}
                <Grid container={true} spacing={2}>
                    <Grid size={{xs: 6}} container={true} sx={{
                        padding: 2,
                        borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                    }}>
                        {state.prevRunId !== null && state.prevRunId > 0 &&
                            <Button variant="contained" color="primary" size="small"
                                    onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/${state.currentTestRun.testId}/run/${state.prevRunId}`)}>
                                <KeyboardArrowLeft/> Prev
                            </Button>
                        }
                    </Grid>
                    <Grid size={{xs: 6}} container={true} justifyContent={"flex-end"} sx={{
                        padding: 1,
                        borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                    }}>
                        {state.nextRunId !== null && state.nextRunId > 0 &&
                            <Button variant="contained" color="primary" size="small"
                                    onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/${state.currentTestRun.testId}/run/${state.nextRunId}`)}>
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
                state.currentTestRun?.appBinary &&
                <TitleCard title={"App Bundle"}>
                    <Box sx={{p: 1, m: 1}}>
                        <Grid container={true}>
                            <Grid size={{xs: 12, md: 2}}>
                                Name:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {state.currentTestRun?.appBinary?.name}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Identifier:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {state.currentTestRun?.appBinary?.identifier}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Platform:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {state.currentTestRun?.appBinary?.platform}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Version:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {state.currentTestRun?.appBinary?.version}
                            </Grid>
                            <Grid size={{xs: 12, md: 2}}>
                                Hash:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {state.currentTestRun?.appBinary?.hash}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Created:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                <DateTime format="YYYY/MM/DD HH:mm:ss" value={state.currentTestRun?.appBinary?.createdAt}/>
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Addons:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {state.currentTestRun?.appBinary?.additional}
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
                                {name: "Open", value: state.runsOpen},
                                {name: "Failed", value: state.runsFailed},
                                {name: "Success", value: state.runsSuccess},
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
                            <Cell fill={'green'}/>
                        </Pie>
                        <Tooltip/>
                    </PieChart>
                    <Grid container={true}>
                        <Grid size={12}>
                            <Typography
                                variant={"caption"}>Open: {state.runsOpen}</Typography>
                        </Grid>
                        <Grid size={12}>
                            <Typography
                                variant={"caption"}>Failed: {state.runsFailed}</Typography>
                        </Grid>
                        <Grid size={12}>
                            <Typography
                                variant={"caption"}>Success: {state.runsSuccess}</Typography>
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
                        {state.runsOpen > 0 &&
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
                                {state.protocols.map((protocol) => (
                                    <TestProtocolRow key={`protocol_${protocol.id}`} row={protocol}/>
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
                            {state.log.map((entry) => <TableRow key={entry.id}>
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
