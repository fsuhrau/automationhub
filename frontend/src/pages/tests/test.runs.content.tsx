import { FC, useEffect, useState } from 'react';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import TestRunDataService from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import { Divider, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core';
import ITestRunData from '../../types/test.run';
import { TestResultState } from '../../types/test.result.state.enum';

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

type TestRunsProps = WithStyles<typeof styles>;

const TestRuns: FC<TestRunsProps> = (props) => {
    const { classes } = props;

    const { testId } = useParams<number>();

    const [testRuns, setTestRuns] = useState<ITestRunData[]>([]);


    useEffect(() => {
        TestRunDataService.getAll(testId).then(response => {
            console.log(response.data);
            setTestRuns(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, [testId]);


    function getStatus(run: ITestRunData) : string {
        return TestResultState[run.Result.Status];
    }

    return (
        <Grid container={true} spacing={2}>
            <Grid item={true} xs={6}>
            </Grid>
            <Grid item={true} xs={5}>
                <Paper className={classes.paper}>
                    <Grid  container={true} spacing={2}>
                        <Grid item={true} xs={2}>
                            Failed
                        </Grid>
                        <Grid item={true} xs={2}>
                            Unstable
                        </Grid>
                        <Grid item={true} xs={2}>
                            Success
                        </Grid>
                        <Grid item={true} xs={2}>
                            Skipped
                        </Grid>
                    </Grid>
                    <Grid  container={true} spacing={2}>
                        <Grid item={true} xs={2}>
                            0
                        </Grid>
                        <Grid item={true} xs={2}>
                            0
                        </Grid>
                        <Grid item={true} xs={2}>
                            0
                        </Grid>
                        <Grid item={true} xs={2}>
                            0
                        </Grid>
                    </Grid>
                    <Divider variant="middle" />
                    <Grid  container={true} spacing={2}>
                        <Grid item={true} xs={12}>
                            see Testrun
                        </Grid>
                    </Grid>
                </Paper>
            </Grid>
            <Grid item={true} xs={12}>
                <TableContainer component={Paper}>
                    <Table sx={{ minWidth: 650 }} size="small" aria-label="a dense table">
                        <TableHead>
                            <TableRow>
                                <TableCell>Run</TableCell>
                                <TableCell align="right">Devices</TableCell>
                                <TableCell align="right">Status</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {testRuns.map((run) => (
                                <TableRow key={run.ID}  sx={{ '&:last-child td, &:last-child th': { border: 0 } }} >
                                    <TableCell component="th" scope="row">
                                        {run.CreatedAt}
                                    </TableCell>
                                    <TableCell align="right">0</TableCell>
                                    <TableCell align="right">{getStatus(run)}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </TableContainer>
            </Grid>
        </Grid>
    );
};

export default withStyles(styles)(TestRuns);
