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
import {makeStyles} from '@mui/styles';
import ITestRunData from '../../types/test.run';
import {TestResultState} from '../../types/test.result.state.enum';
import Moment from 'react-moment';
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
import Grid from "@mui/material/Grid2";
import TestStatusIconComponent from "../../components/test-status-icon.component";
import Collapse from "@mui/material/Collapse";
import Link from "@mui/material/Link";

interface TestRunPageProps {
    testRun: ITestRunData
    nextRunId: number | null
    prevRunId: number | null
}

interface NewTestRunPayload {
    TestRunID: number,
    Entry: ITesRunLogEntryData,
}

const useStyles = makeStyles(theme => ({
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
}));

const TestRunPage: React.FC<TestRunPageProps> = (props: TestRunPageProps) => {
    const classes = useStyles();

    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();

    const {testRun, nextRunId, prevRunId} = props;

    const testContext = useContext(TestContext);
    const {test, setTest} = testContext;

    const navigate = useNavigate();

    const [state, setState] = useState<{
        currentRunID: number,
        log: ITesRunLogEntryData[],
        protocols: ITestProtocolData[],
        runsOpen: number,
        runsFailed: number,
        runsSuccess: number
    }>({
        currentRunID: testRun.ID as number,
        log: [],
        protocols: testRun.Protocols === undefined ? [] : testRun.Protocols,
        runsOpen: 0,
        runsSuccess: 0,
        runsFailed: 0
    })

    const patchProtocols = (protocols: ITestProtocolData[], protocol: ITestProtocolData): ITestProtocolData[] => {
        if (protocols.findIndex(p => p.ID === protocol.ID) >= 0)
            return protocols.map(p => p.ID === protocol.ID ? protocol : p);
        return [...protocols, protocol]
    }

    const testProtocolEntry = useSSE<ITestProtocolData | null>(`test_run_${state.currentRunID}_protocol`, null)
    useEffect(() => {
        if (testProtocolEntry === null) return;
        console.log("Received new protocol entry: " + testProtocolEntry.ID)
        setState(prevState => ({
            ...prevState,
            protocols: patchProtocols(prevState.protocols, testProtocolEntry),
        }));
    }, [testProtocolEntry]);

    useEffect(() => {
        console.log("TestRun changed to: " + testRun.ID)
        setState(prevState => ({...prevState, currentRunID: testRun.ID as number}));
    }, [testRun.ID]);

    const testRunEntry = useSSE<NewTestRunPayload | null>(`test_run_${testRun.ID}_log`, null);
    useEffect(() => {
        if (testRunEntry === null)
            return;
        console.log("Received Test Run Log:  " + testRunEntry.Entry.ID)
        setState(prevState => ({...prevState, log: [...prevState.log, testRunEntry.Entry]}))
    }, [testRunEntry]);

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

    const rebuildStatistics = (run: ITestRunData): { ro: number, rf: number, rs: number } => {
        let ro: number;
        let rf: number;
        let rs: number;
        ro = 0;
        rf = 0;
        rs = 0;

        run.Protocols.forEach(value => {
            switch (value.TestResult) {
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
        let {ro, rf, rs} = rebuildStatistics(testRun);
        setState(prevState => ({
            ...prevState,
            log: testRun.Log,
            runsOpen: ro,
            runsFailed: rf,
            runsSuccess: rs,
            protocols: testRun.Protocols
        }))
    }, [testRun]);

    const onTestRerun = (): void => {
        executeTest(projectIdentifier, appId, testRun.TestID, {
            AppBinaryID: testRun.AppBinaryID,
            Params: testRun.Parameter,
            StartURL: testRun.StartURL,
        }).then(response => {
            navigate(`/project/${projectIdentifier}/app:${appId}/test/${testRun.TestID}/run/${response.data.ID}`);
        }).catch(error => {
            console.log(error);
        });
    };

    const onCancelTestRun = (): void => {
        cancelTestRun(projectIdentifier, appId, testRun.TestID, testRun.ID!).then(response => {
            console.log("Test run cancelled", response);
        }).catch(error => {
            console.log(error);
        });
    };

    const startupTimes = testRun.DeviceStatus.map(item => {
            return {
                Name: item.Device && (item.Device.Alias.length > 0 ? item.Device.Alias : item.Device.Name),
                StartupTime: item.StartupTime
            }
        }
    )

    const environmentParameters = testRun?.Parameter.split("\n");

    const applyProtocolFilter = (value: ITestProtocolData): boolean => {
        return (filter.Success && value.TestResult === TestResultState.TestResultSuccess)
            || (filter.Failed && value.TestResult === TestResultState.TestResultFailed)
            || (filter.Pending && value.TestResult === TestResultState.TestResultOpen);
    }

    const groupProtocols = (protocols: ITestProtocolData[]): ITestProtocolData[] => {
        const protocolMap: { [key: number]: ITestProtocolData } = {};

        protocols.forEach(protocol => {
            if (protocol.ParentTestProtocolID === null || protocol.ParentTestProtocolID === undefined || protocol.ParentTestProtocolID === 0) {
                protocolMap[protocol.ID!] = {...protocol, ChildProtocols: []};
            } else {
                const parent = protocolMap[protocol.ParentTestProtocolID];
                if (parent) {
                    parent.ChildProtocols.push(protocol);
                }
            }
        });
        return Object.values(protocolMap);
    };

    function TestProtocolRow(props: { row: ITestProtocolData }) {
        const {row} = props;
        const [open, setOpen] = React.useState(false);
        const testName = row.TestName.split("/");

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
                            onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/${testRun.TestID}/run/${testRun.ID}/${row.ID}`)}
                            underline="none">
                            {testName}
                        </Link>
                    </TableCell>
                    <TableCell>
                        <Grid container={true}>
                            <Grid size={{xs: 12, md: 12}}>
                                {row.Device && (row.Device.Alias.length > 0 ? row.Device.Alias : row.Device.Name)}
                            </Grid>
                            <Grid size={{xs: 12, md: 12}}>
                                {row.Device?.OS} {row.Device?.OSVersion}
                            </Grid>
                        </Grid>
                    </TableCell>
                    <TableCell align="right"><TestStatusIconComponent status={row.TestResult}/></TableCell>
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
                                        {row.ChildProtocols.map((historyRow) => {
                                            const historyTestName = historyRow.TestName.split("/");

                                            return (
                                                <TableRow key={historyRow.ID}>
                                                    <TableCell component="th" scope="row">
                                                    </TableCell>
                                                    <TableCell>
                                                        <Link
                                                            onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/${testRun.TestID}/run/${testRun.ID}/${historyRow.ID}`)}
                                                            underline="none">
                                                            {historyTestName.length > 1 ? historyTestName[1] : historyTestName[0]}
                                                        </Link>
                                                    </TableCell>
                                                    <TableCell>
                                                        <Grid container={true}>
                                                            <Grid size={{xs: 12, md: 12}}>
                                                                {historyRow.Device && (historyRow.Device.Alias.length > 0 ? historyRow.Device.Alias : historyRow.Device.Name)}
                                                            </Grid>
                                                            <Grid size={{xs: 12, md: 12}}>
                                                                {historyRow.Device?.OS} {historyRow.Device?.OSVersion}
                                                            </Grid>
                                                        </Grid>
                                                    </TableCell>
                                                    <TableCell align="right"><TestStatusIconComponent
                                                        status={historyRow.TestResult}/></TableCell>
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
            <TitleCard title={`Test: ${testRun.Test.Name} Run: ${testRun.ID}`}>
                { /*Test Previous and Next Navigation*/}
                <Grid container={true} spacing={2}>
                    <Grid size={{xs: 6}} container={true} sx={{
                        padding: 2,
                        borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                    }}>
                        {prevRunId !== null && prevRunId > 0 &&
                            <Button variant="contained" color="primary" size="small"
                                    onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/${testRun.TestID}/run/${prevRunId}`)}>
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
                                    onClick={() => navigate(`/project/${projectIdentifier}/app:${appId}/test/${testRun.TestID}/run/${nextRunId}`)}>
                                Next <KeyboardArrowRight/>
                            </Button>
                        }
                    </Grid>
                </Grid>
            </TitleCard>
            <TitleCard title={'Environment'}>
                <Grid container={true}>
                    {environmentParameters.map((e, i) => (
                        <Grid size={{xs: 12, md: 12}}>
                            <Typography key={`env_${i}`} variant={"body1"}>
                                {e}
                            </Typography>
                        </Grid>
                    ))}
                </Grid>
                <Grid sx={{flexGrow: 1}}></Grid>
            </TitleCard>
            {
                testRun?.AppBinary &&
                <TitleCard title={"App Bundle"}>
                    <Box sx={{p: 1, m: 1}}>
                        <Grid container={true}>
                            <Grid size={{xs: 12, md: 2}}>
                                Name:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun?.AppBinary?.Name}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Identifier:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun?.AppBinary?.Identifier}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Platform:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun?.AppBinary?.Platform}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Version:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun?.AppBinary?.Version}
                            </Grid>
                            <Grid size={{xs: 12, md: 2}}>
                                Hash:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun?.AppBinary?.Hash}
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Created:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                <Moment
                                    format="YYYY/MM/DD HH:mm:ss">{testRun?.AppBinary?.CreatedAt}</Moment>
                            </Grid>

                            <Grid size={{xs: 12, md: 2}}>
                                Addons:
                            </Grid>
                            <Grid size={{xs: 12, md: 10}}>
                                {testRun?.AppBinary?.Additional}
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
                <Grid container justifyContent="flex-end" className={classes.chip} padding={1}>
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
                                {groupProtocols(state.protocols.filter(applyProtocolFilter)).map((protocol) => (
                                    <TestProtocolRow row={protocol}/>
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
            <TitleCard title={"Execution Log"}>
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
                            {state.log.map((entry) => <TableRow key={entry.ID}>
                                <TableCell component="th" scope="row" style={{whiteSpace: 'nowrap'}}>
                                    <Moment format="YYYY/MM/DD HH:mm:ss">{entry.CreatedAt}</Moment>
                                </TableCell>
                                <TableCell>{entry.Level}</TableCell>
                                <TableCell>{entry.Log}</TableCell>
                            </TableRow>)}
                        </TableBody>
                    </Table>
                </TableContainer>
            </TitleCard>
        </Box>
    );
};

export default TestRunPage;
