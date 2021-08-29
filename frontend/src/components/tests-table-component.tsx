import React, {useEffect, useState} from 'react';
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

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} size="small" aria-label="a dense table">
                <TableHead>
                    <TableRow>
                        <TableCell>Name</TableCell>
                        <TableCell align="right">Typ</TableCell>
                        <TableCell align="right">Started</TableCell>
                        <TableCell align="right">Devices</TableCell>
                        <TableCell align="right">Status</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {tests.map((test) => <TableRow key={test.Name}>
                        <TableCell component="th" scope="row">
                            {test.Name}
                        </TableCell>
                        <TableCell align="right">{test.TestConfig.Type}</TableCell>
                        <TableCell align="right">{test.Last.Log.StartedAt}</TableCell>
                        <TableCell align="right">0</TableCell>
                        <TableCell align="right">{test.Last.Result.Status}</TableCell>
                    </TableRow>)}
                </TableBody>
            </Table>
        </TableContainer>
    );
}

export default withStyles(styles)(TestsTable);
