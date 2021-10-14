import { FC, useContext, useEffect, useState } from 'react';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
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
} from '@material-ui/core';
import ITestRunData from '../types/test.run';
import { TestResultState } from '../types/test.result.state.enum';
import TestStatusIconComponent from '../components/test-status-icon.component';
import Moment from 'react-moment';
import { useSSE } from 'react-hooks-sse';
import ITesRunLogEntryData from '../types/test.run.log.entry';
import moment from 'moment';
import ITestProtocolData from '../types/test.protocol';
import { executeTest } from '../services/test.service';
import { useHistory } from 'react-router-dom';
import { AppContext } from '../context/app.context';
import { TestContext } from '../context/test.context';
import ITestData from '../types/test';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        paper: {
            maxWidth: 1200,
            margin: 'auto',
            overflow: 'hidden',
        },
        searchBar: {
            borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
        },
        searchInput: {
            fontSize: theme.typography.fontSize,
        },
        block: {
            display: 'block',
        },
        addUser: {
            marginRight: theme.spacing(1),
        },
        contentWrapper: {
            margin: '40px 16px',
        },
        table: {
            minWidth: 650,
        },
    });

interface TestRunContentProps extends WithStyles<typeof styles> {
    testRun: ITestRunData
}

interface NewTestRunPayload {
    TestRunID: number,
    Entry: ITesRunLogEntryData,
}

interface NewProtocolPayload {
    TestRunID: number,
    Protocol: ITestProtocolData,
}

const TestRunContent: FC<TestRunContentProps> = (props) => {
    const { testRun, classes } = props;

    const testContext = useContext(TestContext);
    const { test, setTest } = testContext;

    const history = useHistory();

    const [log, setLog] = useState<Array<ITesRunLogEntryData>>([]);
    const [protocols, setProtocols] = useState<Array<ITestProtocolData>>([]);

    const testProtocol = useSSE<NewProtocolPayload | null>(`test_run_${ testRun.ID }_protocol`, null);
    useEffect(() => {
        if (testProtocol === null)
            return;
        setProtocols(prevState => {
            const newState = [...prevState];
            newState.push(testProtocol.Protocol);
            return newState;
        });
    }, [testProtocol]);


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

    const getDuration = (startedAt: Date, endedAt: Date | undefined): string => {
        if (endedAt !== null && endedAt !== undefined) {
            const start = new Date(startedAt);
            const end = new Date(endedAt);
            const duration = end.valueOf() - start.valueOf();
            const m = moment.utc(duration);
            const secs = duration / 1000;
            if (secs > 60 * 60) {
                return m.format('h') + 'Std ' + m.format('m') + 'Min ' + m.format('s') + 'Sec';
            }
            if (secs > 60) {
                return m.format('m') + 'Min ' + m.format('s') + 'Sec';
            }
            return m.format('s') + 'Sec';
        }
        return 'running';
    };

    useEffect(() => {
        // setTest(testRun.Test);
        setLog(testRun.Log);
        setProtocols(testRun.Protocols);
        rebuildStatistics(testRun);
    }, [testRun]);

    const onTestRerun = (): void => {
        executeTest(testRun.TestID, testRun.AppID, testRun.Parameter).then(response => {
            history.push(`/web/test/${testRun.TestID}/run/${response.data.ID}`);
        }).catch(error => {
            console.log(error);
        });
    };

    return (
        <div>
            <Grid container={ true }>
                <Grid item={ true } xs={ 6 }>
                    <Box component={ Paper } sx={ { p: 2, m: 2 } }>
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
                    <Box component={ Paper } sx={ { p: 2, m: 2 } }>
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
                    <Box component={ Paper } sx={ { p: 2, m: 2 } }>
                        <Typography variant={ 'h6' }>Test Details</Typography>
                        <Divider/>
                        <TableContainer>
                            <Table className={ classes.table } size="small" aria-label="a dense table">
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
                                                { getDuration(protocol.StartedAt, protocol.EndedAt) }
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
                    <Box component={ Paper } sx={ { p: 2, m: 2 } }>
                        <Typography variant={ 'h6' }>Executor Log</Typography>
                        <Divider/>
                        <TableContainer>
                            <Table className={ classes.table } size="small" aria-label="a dense table">
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
        </div>
    );
};

export default withStyles(styles)(TestRunContent);
