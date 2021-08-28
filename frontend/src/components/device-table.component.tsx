import React, {useEffect, useState} from 'react';
import {createStyles, Theme, withStyles, WithStyles} from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';

import IDeviceData from "../types/device";
import DeviceDataService from "../services/device.service";

const styles = (theme: Theme) =>
    createStyles({
        table: {
            minWidth: 650,
        },
    });

export interface DeviceProps extends WithStyles<typeof styles> {
}

function DeviceTable(props: DeviceProps) {
    const {classes} = props;
    const [devices, setDevices] = useState<IDeviceData[]>([]);

    useEffect(() => {
        DeviceDataService.getAll().then(response => {
            console.log(response.data);
            setDevices(response.data);
        }).catch(e => {
            console.log(e);
        })
    }, [])

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} aria-label="simple table">
                <TableHead>
                    <TableRow>
                        <TableCell>Name</TableCell>
                        <TableCell align="right">Identifier</TableCell>
                        <TableCell align="right">Type</TableCell>
                        <TableCell align="right">OS</TableCell>
                        <TableCell align="right">Status</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {devices.map((device) => <TableRow key={device.Name}>
                        <TableCell component="th" scope="row">
                            {device.Name}
                        </TableCell>
                        <TableCell align="right">{device.DeviceIdentifier}</TableCell>
                        <TableCell align="right">{device.DeviceType}</TableCell>
                        <TableCell align="right">{device.OSVersion}</TableCell>
                        <TableCell align="right">{device.Status}</TableCell>
                    </TableRow>)}
                </TableBody>
            </Table>
        </TableContainer>
    );
}

export default withStyles(styles)(DeviceTable);
