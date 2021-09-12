import { FC, useEffect, useState } from 'react';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import TestRunDataService from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import { Box, Link, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core';
import ITestRunData from '../../types/test.run';
import { TestResultState } from '../../types/test.result.state.enum';
import TestStatusIconComponent from '../../components/test-status-icon.component';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        paper: {
            maxWidth: 936,
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

const TestRun: FC<TestRunProps> = (props) => {
    const {} = props;

    const { testId } = useParams<number>();

    const [testRun, setTestRun] = useState<ITestRunData>();

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
        TestRunDataService.getLast(testId).then(response => {
            console.log(response.data);
            setTestRun(response.data);
            rebuildStatistics(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    }, [testId]);

    return (
        <Grid container={true} spacing={12}>
            <Grid item={true} xs={8}>
            </Grid>
            <Grid item={true} xs={4}>
                <Box component={Paper} sx={{ p: 2, m: 2 }}>
                    <Grid container={true} spacing={12}>
                        <Grid item={true} xs={3}>
                            Open
                        </Grid>
                        <Grid item={true} xs={3}>
                            Unstable
                        </Grid>
                        <Grid item={true} xs={3}>
                            Failed
                        </Grid>
                        <Grid item={true} xs={3}>
                            Success
                        </Grid>
                        <Grid item={true} xs={3}>
                            {runsOpen}
                        </Grid>
                        <Grid item={true} xs={3}>
                            {runsUnstable}
                        </Grid>
                        <Grid item={true} xs={3}>
                            {runsFailed}
                        </Grid>
                        <Grid item={true} xs={3}>
                            {runsSuccess}
                        </Grid>
                    </Grid>
                </Box>
            </Grid>
            <Grid item={true} xs={12}>
                <Box component={Paper} sx={{ m: 2 }}>
                    <TableContainer>
                        <Table sx={{ minWidth: 650 }} size="small" aria-label="a dense table">
                            <TableHead>
                                <TableRow>
                                    <TableCell>Device</TableCell>
                                    <TableCell align="right">OS</TableCell>
                                    <TableCell align="right">Status</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {testRun?.Protocols.map((protocol) => (
                                    <TableRow key={protocol.ID}
                                        sx={{ '&:last-child td, &:last-child th': { border: 0 } }}>
                                        <TableCell component="th" scope="row">
                                            <Link
                                                href={`/test/${testRun?.TestID}/run/${testRun?.ID}/${protocol.ID}`}
                                                underline="none">
                                                {protocol.Device.Name}
                                            </Link>
                                        </TableCell>
                                        <TableCell
                                            align="right">{protocol.Device.OS} {protocol.Device.OSVersion}</TableCell>
                                        <TableCell align="right">
                                            <TestStatusIconComponent status={protocol.TestResult} />
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>
                </Box>
            </Grid>
        </Grid>
    );
};

export default withStyles(styles)(TestRun);
