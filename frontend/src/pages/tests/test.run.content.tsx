import { FC, useEffect, useState } from 'react';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import TestRunDataService from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import {
    Button,
    Card,
    CardActions,
    CardContent, Link,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
} from '@material-ui/core';
import ITestRunData from '../../types/test.run';
import { TestResultState } from '../../types/test.result.state.enum';
import ITestProtocolData from '../../types/test.protocol';

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
    const { } = props;

    const { testId } = useParams<number>();

    const [testRun, setTestRun] = useState<ITestRunData>();

    const [runsOpen, setRunsOpen] = useState<number>();
    const [runsUnstable, setRunsUnstable] = useState<number>();
    const [runsFailed, setRunsFailed] = useState<number>();
    const [runsSuccess, setRunsSuccess] = useState<number>();


    function getStatus(protocol: ITestProtocolData) : string {
        return TestResultState[protocol.TestResult];
    }

    function rebuildStatistics(run: ITestRunData) : void {
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
                <Card sx={{ minWidth: 275 }}>
                    <CardContent>
                        <Grid container={true} spacing={2}>
                            <Grid item={true} xs={2}>
                                Open
                            </Grid>
                            <Grid item={true} xs={2}>
                                Unstable
                            </Grid>
                            <Grid item={true} xs={2}>
                                Failed
                            </Grid>
                            <Grid item={true} xs={2}>
                                Success
                            </Grid>
                        </Grid>
                        <Grid container={true} spacing={2}>
                            <Grid item={true} xs={2}>
                                {runsOpen}
                            </Grid>
                            <Grid item={true} xs={2}>
                                {runsUnstable}
                            </Grid>
                            <Grid item={true} xs={2}>
                                {runsFailed}
                            </Grid>
                            <Grid item={true} xs={2}>
                                {runsSuccess}
                            </Grid>
                        </Grid>
                    </CardContent>
                    <CardActions>
                        <Button size="small">Learn More</Button>
                    </CardActions>
                </Card>
            </Grid>
            <Grid item={true} xs={12}>
                <TableContainer component={Paper}>
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
                                <TableRow key={protocol.ID} sx={{ '&:last-child td, &:last-child th': { border: 0 } }}>
                                    <TableCell component="th" scope="row">
                                        <Link href={`/test/${testRun?.TestID}/run/${testRun?.ID}/${protocol.ID}` } underline="none">
                                            {protocol.Device.Name}
                                        </Link>
                                    </TableCell>
                                    <TableCell align="right">{protocol.Device.OS} {protocol.Device.OSVersion}</TableCell>
                                    <TableCell align="right">{getStatus(protocol)}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </TableContainer>
            </Grid>
        </Grid>
    );
};

export default withStyles(styles)(TestRun);
