import { FC, MouseEvent, useEffect, useState } from 'react';
import { createStyles, withStyles, WithStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';

import ITestData from '../types/test';
import { executeTest, getAllTests } from '../services/test.service';
import { Button, Link } from '@material-ui/core';
import { PlayArrow } from '@material-ui/icons';
import TestStatusIconComponent from './test-status-icon.component';
import { TestResultState } from '../types/test.result.state.enum';

const styles = (): ReturnType<typeof createStyles> =>
    createStyles({
        table: {
            minWidth: 650,
        },
    });

export type TestProps = WithStyles<typeof styles>;

const TestsTable: FC<TestProps> = (props) => {
    const { classes } = props;
    const [tests, setTests] = useState<ITestData[]>([]);

    useEffect(() => {
        getAllTests().then(response => {
            console.log(response.data);
            setTests(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    const typeString = (type: number): string => {
        switch (type) {
            case 0: return 'Unity';
            case 1: return 'Cocos';
            case 2: return 'Serenity';
            case 3: return 'Scenario';
        }
        return '';
    };

    const executionString = (type: number): string => {
        switch (type) {
            case 0: return 'Parallel';
            case 1: return 'Synchronous';
        }
        return '';
    };

    const handleRunTest = (id: number | null | undefined, appid: number, event: MouseEvent<HTMLButtonElement>): void => {
        event.preventDefault();
        executeTest(id, appid).then(response => {
            console.log(response.data);
        }).catch(error => {
            console.log(error);
        });
    };

    const getTestStatus = (test: ITestData): TestResultState => {
        if (test.TestRuns == null) {
            return TestResultState.TestResultOpen;
        }

        const lastRun = test.TestRuns[test.TestRuns.length - 1];
        return lastRun.TestResult;
    };

    const getDevices = (test: ITestData): string => {
        if (test.TestConfig.AllDevices) {
            return 'all';
        }
        if (test.TestConfig.Devices !== null) {
            return test.TestConfig.Devices.length.toString();
        }
        return 'n/a';
    };

    const getTests = (test: ITestData): string => {
        if (test.TestConfig.Unity !== undefined && test.TestConfig.Unity !== null) {
            if (test.TestConfig.Unity?.RunAllTests) {
                return 'all';
            }
            if (test.TestConfig.Unity.UnityTestFunctions !== null) {
                return test.TestConfig.Unity.UnityTestFunctions.length.toString();
            }
        }

        return 'n/a';
    };

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} size="small" aria-label="a dense table">
                <TableHead>
                    <TableRow>
                        <TableCell>Name</TableCell>
                        <TableCell align="right">Typ</TableCell>
                        <TableCell align="right">Execution</TableCell>
                        <TableCell align="right">Devices</TableCell>
                        <TableCell align="right">Tests</TableCell>
                        <TableCell align="right">Status</TableCell>
                        <TableCell align="right"/>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {tests.map((test) => <TableRow key={test.Name}>
                        <TableCell component="th" scope="row">
                            <Link href={`test/${test.ID}/runs/last` } underline="none">
                                {test.Name}
                            </Link>
                        </TableCell>
                        <TableCell align="right">{typeString(test.TestConfig.Type)}</TableCell>
                        <TableCell align="right">{executionString(test.TestConfig.ExecutionType)}</TableCell>
                        <TableCell align="right">{getDevices(test)}</TableCell>
                        <TableCell align="right">{getTests(test)}</TableCell>
                        <TableCell align="right"><TestStatusIconComponent status={getTestStatus(test)}/></TableCell>
                        <TableCell align="right">
                            <Button color="primary" size="small" variant="outlined" endIcon={<PlayArrow />} onClick={(e) => handleRunTest(test.ID, 1, e)}>
                                Run
                            </Button>
                        </TableCell>
                    </TableRow>)}
                </TableBody>
            </Table>
        </TableContainer>
    );
};

export default withStyles(styles)(TestsTable);
