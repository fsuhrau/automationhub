import { FC, useEffect, useState } from 'react';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import { getLastRun } from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import {
    Box,
    Divider,
    Link,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Typography,
} from '@material-ui/core';
import ITestRunData from '../../types/test.run';
import { TestResultState } from '../../types/test.result.state.enum';
import TestStatusIconComponent from '../../components/test-status-icon.component';
import Moment from 'react-moment';
import { useSSE } from 'react-hooks-sse';
import ITesRunLogEntryData from '../../types/test.run.log.entry';
import moment from 'moment';

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

type TestRunProps = WithStyles<typeof styles>;

interface NewTestRunPayload {
    TestRunID: number,
    Entry: ITesRunLogEntryData,
}

const TestRun: FC<TestRunProps> = (props) => {
    const {} = props;

    const { testId } = useParams<number>();

    const [testRun, setTestRun] = useState<ITestRunData>();
    const [log, setLog] = useState<Array<ITesRunLogEntryData>>([]);

    const testRunEntry = useSSE<NewTestRunPayload | null>('testlog', null);
    useEffect(() => {
        if (testRunEntry === null)
            return;
        if (testRun?.ID !== testRunEntry.TestRunID) {
            return;
        }
        setLog(prevState => {
            const newState = [...prevState];
            newState.push(testRunEntry.Entry);
            return newState;
        });
    }, [testRunEntry, testRun?.ID]);

    const [runsOpen, setRunsOpen] = useState<number>();
    const [runsUnstable, setRunsUnstable] = useState<number>();
    const [runsFailed, setRunsFailed] = useState<number>();
    const [runsSuccess, setRunsSuccess] = useState<number>();

    function rebuildStatistics(run: ITestRunData): void {
        let ro: number;
        let ru: number;
        let rf: number;
        let rs: number;
        ro = 0;
        ru = 0;
        rf = 0;
        rs = 0;

        run.Protocols.forEach(value => {
            switch (value.TestResult) {
                case TestResultState.TestResultOpen:
                    ro++;
                    break;
                case TestResultState.TestResultUnstable:
                    ru++;
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
        setRunsUnstable(ru);
        setRunsFailed(rf);
        setRunsSuccess(rs);
    }

    useEffect(() => {
        getLastRun(testId).then(response => {
            console.log(response.data);
            setTestRun(response.data);
            rebuildStatistics(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    }, [testId]);

    useEffect(() => {
        if (testRun !== undefined) {
            setLog(testRun.Log);
        }
    }, [testRun]);

    const getDuration = (startedAt: Date, endedAt: Date): string => {
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

    return (
        <Grid container={ true } spacing={ 12 }>
            <Grid item={ true } xs={ 6 }>
                <Box component={ Paper } sx={ { p: 2, m: 2 } }>
                    <Typography variant={ 'h6' }>App Details</Typography>
                    <Divider/>
                    <Grid container={ true } spacing={ 12 }>
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
                </Box>
            </Grid>
            <Grid item={ true } xs={ 2 }>
            </Grid>
            <Grid item={ true } xs={ 4 }>
                <Box component={ Paper } sx={ { p: 2, m: 2 } }>
                    <Typography variant={ 'h6' }>Test Results</Typography>
                    <Divider/>
                    <Grid container={ true } spacing={ 12 }>
                        <Grid item={ true } xs={ 4 } align="center">
                            Open
                        </Grid>
                        <Grid item={ true } xs={ 4 } align="center">
                            Failed
                        </Grid>
                        <Grid item={ true } xs={ 4 } align="center">
                            Success
                        </Grid>
                        <Grid item={ true } xs={ 4 } align="center">
                            { runsOpen }
                        </Grid>
                        <Grid item={ true } xs={ 4 } align="center">
                            { runsFailed }
                        </Grid>
                        <Grid item={ true } xs={ 4 } align="center">
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
                        <Table sx={ { minWidth: 650 } } size="small" aria-label="a dense table">
                            <TableHead>
                                <TableRow>
                                    <TableCell>Device</TableCell>
                                    <TableCell align="right">OS</TableCell>
                                    <TableCell>Test</TableCell>
                                    <TableCell>Duration</TableCell>
                                    <TableCell align="right">Status</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                { testRun?.Protocols.map((protocol) => (
                                    <TableRow key={ protocol.ID }
                                        sx={ { '&:last-child td, &:last-child th': { border: 0 } } }>
                                        <TableCell component="th" scope="row">
                                            <Link
                                                href={ `/test/${ testRun?.TestID }/run/${ testRun?.ID }/${ protocol.ID }` }
                                                underline="none">
                                                { protocol.Device.Name }
                                            </Link>
                                        </TableCell>
                                        <TableCell
                                            align="right">{ protocol.Device.OS } { protocol.Device.OSVersion }
                                        </TableCell>
                                        <TableCell>{ protocol.TestName }</TableCell>
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
                    <Typography variant={ 'h6' }>Executer Log</Typography>
                    <Divider/>
                    <TableContainer>
                        <Table sx={ { minWidth: 650 } } size="small" aria-label="a dense table">
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
    );
};

export default withStyles(styles)(TestRun);
