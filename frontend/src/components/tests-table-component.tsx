import React, {useEffect, useState, MouseEvent} from 'react';
import {createStyles, styled, Theme, withStyles, WithStyles} from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';

import ITestData from "../types/test";
import TestDataService from "../services/test.service";
import {Button} from "@material-ui/core";
import {PlayArrow} from "@material-ui/icons";

const styles = (theme: Theme) =>
    createStyles({
        table: {
            minWidth: 650,
        },
    });

export interface TestProps extends WithStyles<typeof styles> {
}

function TestsTable(props: TestProps) {
    const {classes} = props;
    const [tests, setTests] = useState<ITestData[]>([]);

    useEffect(() => {
        TestDataService.getAll().then(response => {
            console.log(response.data);
            setTests(response.data);
        }).catch(e => {
            console.log(e);
        })
    }, [])

    function typeString(type: number) {
        switch (type) {
            case 0: return "Unity";
            case 1: return "Cocos";
            case 2: return "Serenity";
            case 3: return "Scenario";
        }
        return "";
    }

    function executionString(type: number) {
        switch (type) {
            case 0: return "Parallel";
            case 1: return "Synchronous";
        }
        return "";
    }

    function handleRunTest(id: number | null | undefined, appid: number, devices: Array<number>, e: MouseEvent<HTMLButtonElement>) {
        e.preventDefault()
        TestDataService.executeTest(id, appid, devices).then(response => {
            console.log(response.data);
        }).catch(e => {
            console.log(e);
        })
    }

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} size="small" aria-label="a dense table">
                <TableHead>
                    <TableRow>
                        <TableCell>Name</TableCell>
                        <TableCell align="right">Typ</TableCell>
                        <TableCell align="right">Execution</TableCell>
                        <TableCell align="right">Devices</TableCell>
                        <TableCell align="right">Status</TableCell>
                        <TableCell align="right"></TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {tests.map((test) => <TableRow key={test.Name}>
                        <TableCell component="th" scope="row">
                            {test.Name}
                        </TableCell>
                        <TableCell align="right">{typeString(test.TestConfig.Type)}</TableCell>
                        <TableCell align="right">{executionString(test.TestConfig.ExecutionType)}</TableCell>
                        <TableCell align="right">0</TableCell>
                        <TableCell align="right"></TableCell>
                        <TableCell align="right">
                            <Button color="primary" size="small" variant="outlined" endIcon={<PlayArrow />} onClick={(e) => handleRunTest(test.ID, 1, [1], e)}>
                                Run
                            </Button>
                        </TableCell>
                    </TableRow>)}
                </TableBody>
            </Table>
        </TableContainer>
    );
}

export default withStyles(styles)(TestsTable);
