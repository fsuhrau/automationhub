import React, { useContext, useEffect, useState } from 'react';

import {
    Box,
    Button,
    Divider,
    Grid,
    Link,
    Paper,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Typography,
} from '@mui/material';
import ITestRunData from '../types/test.run';
import { TestResultState } from '../types/test.result.state.enum';
import TestStatusIconComponent from '../components/test-status-icon.component';
import Moment from 'react-moment';
import { useSSE } from 'react-hooks-sse';
import ITesRunLogEntryData from '../types/test.run.log.entry';
import ITestProtocolData from '../types/test.protocol';
import { executeTest } from '../services/test.service';
import { useNavigate } from 'react-router-dom';
import { TestContext } from '../context/test.context';
import {
    Cancel,
    CheckCircle, DirectionsRun,
    KeyboardArrowLeft,
    KeyboardArrowRight
} from '@mui/icons-material';
import { useProjectAppContext } from "../project/app.context";
import { TitleCard } from "./title.card.component";
import { Bar, BarChart, CartesianGrid, Cell, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis } from 'recharts';
import { makeStyles } from "@mui/styles";

interface TestRunContentProps {
    testRun: ITestRunData
    nextRunId: number | undefined
    prevRunId: number | undefined
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

const TestRunContent: React.FC<TestRunContentProps> = (props: TestRunContentProps) => {
    const classes = useStyles();

    const {projectId, appId} = useProjectAppContext();

    const {testRun, nextRunId, prevRunId} = props;

    const testContext = useContext(TestContext);
    const {test, setTest} = testContext;

    const navigate = useNavigate();

    const [log, setLog] = useState<Array<ITesRunLogEntryData>>([]);

    const protocols = useSSE<ITestProtocolData[], ITestProtocolData>(`test_run_${ testRun.ID }_protocol`, testRun.Protocols, {
        stateReducer: (state, changes) => {
            if (state.findIndex(value => value.ID == changes.data.ID) >= 0) {
                return state.map(value => {
                    return value.ID === changes.data.ID ? changes.data : value;
                });
            } else {
                return [...state, changes.data];
            }
        },
    });

    const testRunEntry = useSSE<NewTestRunPayload | null>(`test_run_${ testRun.ID }_log`, null);
    useEffect(() => {
        if (testRunEntry === null)
            return;

        setLog(prevState => {
            const newState = [...prevState];
            newState.push(testRunEntry.Entry);
            return newState;
        });
    }, [testRunEntry]);

    const [runsOpen, setRunsOpen] = useState<number>();
    const [runsFailed, setRunsFailed] = useState<number>();
    const [runsSuccess, setRunsSuccess] = useState<number>();

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

    function rebuildStatistics(run: ITestRunData): void {
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

        setRunsOpen(ro);
        setRunsFailed(rf);
        setRunsSuccess(rs);
    }

    useEffect(() => {
        setLog(testRun.Log);
        rebuildStatistics(testRun);
    }, [testRun]);

    const onTestRerun = (): void => {
        executeTest(projectId, appId, testRun.TestID, testRun.AppBinaryID, testRun.Parameter).then(response => {
            navigate(`/project/${ projectId }/app/${ appId }/test/${ testRun.TestID }/run/${ response.data.ID }`);
        }).catch(error => {
            console.log(error);
        });
    };

    const startupTimes = testRun.DeviceStatus.map(item => {
            return {Name: item.Device.Name, StartupTime: item.StartupTime}
        }
    )

    const environmentParameters = testRun?.Parameter.split("\n");

    return (
        <Grid container={ true } spacing={ 2 }>
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "h1" }>Test: { testRun.Test.Name } Run: { testRun.ID }</Typography>
            </Grid>
            <Grid item={ true } xs={ 12 }>
                <Divider/>
            </Grid>
            <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" } justifyContent={ "center" }>
                <Grid
                    item={ true }
                    xs={ 12 }
                    style={ {maxWidth: 800} }
                >
                    <Grid item={ true } container={ true } xs={ 12 } spacing={ 2 } alignItems={ "center" }>
                        <Grid item={ true }>
                            { prevRunId !== undefined && prevRunId > 0 && <Button variant="contained" color="primary" size="small"
                                                       href={ `/project/${ projectId }/app/${ appId }/test/${ testRun.TestID }/run/${ prevRunId } ` }>
                                <KeyboardArrowLeft/> Prev
                            </Button>
                            }
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                            { nextRunId !== undefined && nextRunId > 0 && <Button variant="contained" color="primary" size="small"
                                                       href={ `/project/${ projectId }/app/${ appId }/test/${ testRun.TestID }/run/${ nextRunId } ` }>
                                Next <KeyboardArrowRight/>
                            </Button>
                            }
                        </Grid>
                    </Grid>
                    <TitleCard title={ "Results" }>
                        <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" }
                              justifyContent={ "center" }>
                            <PieChart
                                width={ 200 }
                                height={ 200 }
                            >
                                <Pie
                                    data={ [
                                        {name: "Open", value: runsOpen},
                                        {name: "Failed", value: runsFailed},
                                        {name: "Success", value: runsSuccess},
                                    ] }
                                    cx="50%"
                                    cy="50%"
                                    label
                                    outerRadius={ 70 }
                                    fill="#8884d8"
                                    dataKey="value"
                                >
                                    <Cell fill={ 'yellow' }/>
                                    <Cell fill={ 'red' }/>
                                    <Cell fill={ 'green' }/>
                                </Pie>
                                <Tooltip/>
                            </PieChart>
                            <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" }
                                  justifyContent={ "center" }>
                                <Typography
                                    variant={ "caption" }> Open: { runsOpen } Failed: { runsFailed } Success: { runsSuccess }</Typography>
                            </Grid>
                        </Grid>
                    </TitleCard>
                    <TitleCard title={ "Environment" }>
                        <Grid item={ true } container={ true } xs={ 12 }>
                            <Grid container={ true }>
                                <Grid item={ true } container={true} xs={ 12 }>
                                    { environmentParameters.map((e, i) => (
                                        <Grid key={`env_${i}`} item={true} xs={12}>
                                            <Typography variant={"body1"}>
                                                { e }
                                            </Typography>
                                        </Grid>
                                    ))}
                                </Grid>
                                <Grid item={ true } xs={ true }/>
                                <Grid item={ true } container={ true } xs={ 12 } alignItems={ "flex-end" }
                                      justifyContent={ "flex-end" }>
                                    <Button variant="contained" color="primary" onClick={ onTestRerun }>
                                        Rerun
                                    </Button>
                                </Grid>
                            </Grid>
                        </Grid>
                    </TitleCard>
                    <TitleCard title={ "Test Functions" }>
                        <Grid container justifyContent="flex-end" className={ classes.chip } padding={1} >
                            <CheckCircle htmlColor={filter.Success ? 'green' : 'lightgray'}
                                         onClick={ () => setFilter(prevState => ({...prevState, Success: !prevState.Success})) }
                            />
                            <Cancel htmlColor={filter.Failed ? 'red' : 'lightgray'}
                                    onClick={ () => setFilter(prevState => ({...prevState, Failed: !prevState.Failed})) }
                            />
                            <DirectionsRun htmlColor={filter.Pending ? 'yellow' : 'lightgray'}
                                      onClick={ () => setFilter(prevState => ({...prevState, Pending: !prevState.Pending})) }
                            />
                        </Grid>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} } >
                            <TableContainer>
                                <Table size="small" aria-label="a dense table">
                                    <TableHead>
                                        <TableRow>
                                            <TableCell>Test</TableCell>
                                            <TableCell>Device</TableCell>
                                            <TableCell align="right">Status</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        { protocols.filter(value => (filter.Success && value.TestResult === TestResultState.TestResultSuccess) || (filter.Failed && value.TestResult === TestResultState.TestResultFailed) || (filter.Pending && value.TestResult === TestResultState.TestResultOpen) ).map((protocol) => {
                                            const testName = protocol.TestName.split("/");
                                            return (
                                                <TableRow key={ protocol.ID }>
                                                    <TableCell component="th" scope="row">
                                                        <Link
                                                            href={ `/project/${ projectId }/app/${ testRun.Test.AppID }/test/${ testRun.TestID }/run/${ testRun.ID }/${ protocol.ID }` }
                                                            underline="none">
                                                            { testName[0] } <br/> { testName[1] }
                                                        </Link>
                                                    </TableCell>
                                                    <TableCell>
                                                        <Grid container={ true }>
                                                            <Grid item={ true } xs={ 12 }>
                                                                { protocol.Device?.Name }
                                                            </Grid>
                                                            <Grid item={ true } xs={ 12 }>
                                                                { protocol.Device?.OS } { protocol.Device?.OSVersion }
                                                            </Grid>
                                                        </Grid>
                                                    </TableCell>
                                                    <TableCell align="right">
                                                        <TestStatusIconComponent status={ protocol.TestResult }/>
                                                    </TableCell>
                                                </TableRow>
                                            )
                                        }) }
                                    </TableBody>
                                </Table>
                            </TableContainer>
                        </Paper>
                    </TitleCard>
                    { testRun?.AppBinary &&
                        <TitleCard title={ "App Bundle" }>
                            <Box sx={ {p: 1, m: 1} }>
                                <Grid container={ true }>
                                    <Grid item={ true } xs={ 2 }>
                                        Name:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { testRun?.AppBinary?.Name }
                                    </Grid>

                                    <Grid item={ true } xs={ 2 }>
                                        Identifier:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { testRun?.AppBinary?.Identifier }
                                    </Grid>

                                    <Grid item={ true } xs={ 2 }>
                                        Platform:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { testRun?.AppBinary?.Platform }
                                    </Grid>

                                    <Grid item={ true } xs={ 2 }>
                                        Version:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { testRun?.AppBinary?.Version }
                                    </Grid>
                                    <Grid item={ true } xs={ 2 }>
                                        Hash:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { testRun?.AppBinary?.Hash }
                                    </Grid>

                                    <Grid item={ true } xs={ 2 }>
                                        Created:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        <Moment
                                            format="YYYY/MM/DD HH:mm:ss">{ testRun?.AppBinary?.CreatedAt }</Moment>
                                    </Grid>

                                    <Grid item={ true } xs={ 2 }>
                                        Addons:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { testRun?.AppBinary?.Additional }
                                    </Grid>
                                </Grid>
                            </Box>
                        </TitleCard>
                    }
                    <TitleCard title={ "App Startup Time" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <ResponsiveContainer width={ '100%' } height={ 200 }>
                                <BarChart width={ 600 } height={ 200 } data={startupTimes} margin={ {
                                    top: 5,
                                    right: 30,
                                    left: 20,
                                    bottom: 5,
                                } }>
                                    <CartesianGrid strokeDasharray="3 3"/>
                                    <XAxis dataKey="Name"/>
                                    <Tooltip/>
                                    <Bar dataKey="StartupTime" fill="#8884d8" label={{ position: 'top' }} unit={'ms'}/>
                                </BarChart>
                            </ResponsiveContainer>
                        </Paper>
                    </TitleCard>
                    <TitleCard title={ "Execution Log" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <Box sx={ {p: 1, m: 1} }>
                                <TableContainer>
                                    <Table size="small" aria-label="a dense table">
                                        <TableHead>
                                            <TableRow>
                                                <TableCell>Date</TableCell>
                                                <TableCell>Level</TableCell>
                                                <TableCell>Log</TableCell>
                                            </TableRow>
                                        </TableHead>
                                        <TableBody>
                                            { log.map((entry) => <TableRow key={ entry.ID }>
                                                <TableCell component="th" scope="row" style={ {whiteSpace: 'nowrap'} }>
                                                    <Moment format="YYYY/MM/DD HH:mm:ss">{ entry.CreatedAt }</Moment>
                                                </TableCell>
                                                <TableCell>{ entry.Level }</TableCell>
                                                <TableCell>{ entry.Log }</TableCell>
                                            </TableRow>) }
                                        </TableBody>
                                    </Table>
                                </TableContainer>
                            </Box>
                        </Paper>
                    </TitleCard>
                </Grid>
            </Grid>
        </Grid>
    );
};

export default TestRunContent;
