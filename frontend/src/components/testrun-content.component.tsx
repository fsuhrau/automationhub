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
import moment from 'moment';
import ITestProtocolData, { duration } from '../types/test.protocol';
import { executeTest } from '../services/test.service';
import { useHistory } from 'react-router-dom';
import { TestContext } from '../context/test.context';
import { KeyboardArrowLeft, KeyboardArrowRight } from '@mui/icons-material';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';

interface TestRunContentProps {
    testRun: ITestRunData
    nextRunId: number
    prevRunId: number
}

interface NewTestRunPayload {
    TestRunID: number,
    Entry: ITesRunLogEntryData,
}

interface NewProtocolPayload {
    TestRunID: number,
    Protocol: ITestProtocolData,
}

const TestRunContent: React.FC<TestRunContentProps> = (props) => {

    const { testRun, nextRunId, prevRunId } = props;

    const testContext = useContext(TestContext);
    const { test, setTest } = testContext;

    const history = useHistory();

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
        executeTest(testRun.TestID, testRun.AppID, testRun.Parameter).then(response => {
            history.push(`/web/test/${ testRun.TestID }/run/${ response.data.ID }`);
        }).catch(error => {
            console.log(error);
        });
    };

    return (
        <Paper sx={{ maxWidth: 1200, margin: 'auto', overflow: 'hidden' }}>
            <AppBar
                position="static"
                color="default"
                elevation={0}
                sx={{ borderBottom: '1px solid rgba(0, 0, 0, 0.12)' }}
            >
                <Toolbar>
                    <Grid container={ true } spacing={ 2 } alignItems="center">
                        <Grid item={ true }>
                            { prevRunId > 0 && <Button variant="contained" color="primary" size="small"
                                href={ `/web/test/${ testRun.TestID }/run/${ prevRunId } ` }>
                                <KeyboardArrowLeft/> Prev
                            </Button>
                            }
                        </Grid>
                        <Grid item={ true } xs={ true }>
                            <Typography variant={ 'h6' }>
                                Test: { testRun.Test.Name } Run: { testRun.ID }
                            </Typography>
                        </Grid>
                        <Grid item={ true }>
                            { nextRunId > 0 && <Button variant="contained" color="primary" size="small"
                                href={ `/web/test/${ testRun.TestID }/run/${ nextRunId } ` }>
                                Next <KeyboardArrowRight/>
                            </Button>
                            }
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Grid container={ true }>
                <Grid item={ true } xs={ 6 }>
                    <Box sx={ { p: 2, m: 2 } }>
                        <Typography variant={ 'h6' }>App Details</Typography>
                        <Divider/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                Name:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { testRun?.App?.Name }
                            </Grid>

                            <Grid item={ true } xs={ 2 }>
                                Identifier:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { testRun?.App?.Identifier }
                            </Grid>

                            <Grid item={ true } xs={ 2 }>
                                Platform:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { testRun?.App?.Platform }
                            </Grid>

                            <Grid item={ true } xs={ 2 }>
                                Version:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { testRun?.App?.Version }
                            </Grid>

                            <Grid item={ true } xs={ 2 }>
                                Hash:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { testRun?.App?.Hash }
                            </Grid>

                            <Grid item={ true } xs={ 2 }>
                                Created:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <Moment format="YYYY/MM/DD HH:mm:ss">{ testRun?.App?.CreatedAt }</Moment>
                            </Grid>

                            <Grid item={ true } xs={ 2 }>
                                Addons:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { testRun?.App?.Additional }
                            </Grid>
                        </Grid>
                        <br/>
                        <Typography variant={ 'h6' }>Environment Parameter</Typography>
                        <Divider/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 10 }>
                                { testRun?.Parameter }
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                <br/>
                                <Button variant="contained" color="primary" onClick={ onTestRerun }>
                                    Rerun
                                </Button>
                            </Grid>
                        </Grid>
                    </Box>
                </Grid>
                <Grid item={ true } xs={ 2 }>
                </Grid>
                <Grid item={ true } xs={ 4 }>
                    <Box sx={ { p: 2, m: 2 } }>
                        <Typography variant={ 'h6' }>Test Results</Typography>
                        <Divider/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 4 }>
                                Open
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                Failed
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                Success
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                { runsOpen }
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                { runsFailed }
                            </Grid>
                            <Grid item={ true } xs={ 4 }>
                                { runsSuccess }
                            </Grid>
                        </Grid>
                    </Box>
                </Grid>
                <Grid item={ true } xs={ 12 }>
                    <Box sx={ { p: 2, m: 2 } }>
                        <Typography variant={ 'h6' }>Test Details</Typography>
                        <Divider/>
                        <TableContainer>
                            <Table size="small" aria-label="a dense table">
                                <TableHead>
                                    <TableRow>
                                        <TableCell>Test</TableCell>
                                        <TableCell>Device</TableCell>
                                        <TableCell align="right">OS</TableCell>
                                        <TableCell>Duration</TableCell>
                                        <TableCell align="right">Status</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    { protocols.map((protocol) => (
                                        <TableRow key={ protocol.ID }>
                                            <TableCell component="th" scope="row">
                                                <Link
                                                    href={ `/web/test/${ testRun.TestID }/run/${ testRun.ID }/${ protocol.ID }` }
                                                    underline="none">
                                                    { protocol.TestName }
                                                </Link>
                                            </TableCell>
                                            <TableCell>
                                                { protocol.Device?.Name }
                                            </TableCell>
                                            <TableCell
                                                align="right">{ protocol.Device?.OS } { protocol.Device?.OSVersion }
                                            </TableCell>
                                            <TableCell>
                                                { duration(protocol.StartedAt, protocol.EndedAt) }
                                            </TableCell>
                                            <TableCell align="right">
                                                <TestStatusIconComponent status={ protocol.TestResult }/>
                                            </TableCell>
                                        </TableRow>
                                    )) }
                                </TableBody>
                            </Table>
                        </TableContainer>
                    </Box>
                </Grid>
                <Grid item={ true } xs={ 12 }>
                    <Box sx={ { p: 2, m: 2 } }>
                        <Typography variant={ 'h6' }>Executor Log</Typography>
                        <Divider/>
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
                                        <TableCell component="th" scope="row" style={ { whiteSpace: 'nowrap' } }>
                                            <Moment format="YYYY/MM/DD HH:mm:ss">{ entry.CreatedAt }</Moment>
                                        </TableCell>
                                        <TableCell>{ entry.Level }</TableCell>
                                        <TableCell>{ entry.Log }</TableCell>
                                    </TableRow>) }
                                </TableBody>
                            </Table>
                        </TableContainer>
                    </Box>
                </Grid>
            </Grid>
        </Paper>
    );
};

export default TestRunContent;
