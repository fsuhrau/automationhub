import React, {useEffect, useState} from 'react';
import Paper from '@mui/material/Paper';
import {getAllRuns} from '../../services/test.run.service';
import {useParams} from 'react-router-dom';
import {Divider, Grid, Table, TableBody, TableCell, TableContainer, TableHead, TableRow} from '@mui/material';
import ITestRunData from '../../types/test.run';
import {TestResultState} from '../../types/test.result.state.enum';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useApplicationContext} from "../../hooks/ApplicationProvider";

const TestRuns: React.FC = () => {

    const {testId} = useParams();

    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();

    const [testRuns, setTestRuns] = useState<ITestRunData[]>([]);


    useEffect(() => {
        if (testId === undefined) {
            return;
        }

        getAllRuns(projectIdentifier, appId as number, testId).then(response => {
            setTestRuns(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, [projectIdentifier, appId, testId]);


    function getStatus(run: ITestRunData): string {
        return TestResultState[run.TestResult];
    }

    return (
        <Grid container={true} spacing={2}>
            <Grid item={true} xs={6}>
            </Grid>
            <Grid item={true} xs={5}>
                <Paper sx={{maxWidth: 1800, margin: 'auto', overflow: 'hidden'}}>
                    <Grid container={true} spacing={2}>
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
                    <Grid container={true} spacing={2}>
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
                    <Divider variant="middle"/>
                    <Grid container={true} spacing={2}>
                        <Grid item={true} xs={12}>
                            see Testrun
                        </Grid>
                    </Grid>
                </Paper>
            </Grid>
            <Grid item={true} xs={12}>
                <TableContainer component={Paper}>
                    <Table sx={{minWidth: 850}} size="medium" aria-label="a dense table">
                        <TableHead>
                            <TableRow>
                                <TableCell>Run</TableCell>
                                <TableCell align="right">Devices</TableCell>
                                <TableCell align="right">Status</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {testRuns.map((run) => (
                                <TableRow key={run.ID} sx={{'&:last-child td, &:last-child th': {border: 0}}}>
                                    <TableCell component="th" scope="row">
                                        {run.CreatedAt.toLocaleString()}
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

export default TestRuns;
