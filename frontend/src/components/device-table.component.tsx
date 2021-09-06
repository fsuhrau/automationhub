import { FC, MouseEvent, useEffect, useState } from 'react';
import { createStyles, withStyles, WithStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';

import IDeviceData from '../types/device';
import DeviceDataService from '../services/device.service';
import { Button } from '@material-ui/core';
import { PlayArrow } from '@material-ui/icons';

const styles = (): ReturnType<typeof createStyles> =>
    createStyles({
        table: {
            minWidth: 650,
        },
    });

export type DeviceProps = WithStyles<typeof styles>;

const DeviceTable: FC<DeviceProps> = (props) => {
    const { classes } = props;
    const [devices, setDevices] = useState<IDeviceData[]>([]);

    useEffect(() => {
        DeviceDataService.getAll().then(response => {
            console.log(response.data);
            setDevices(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    function deviceState(state: number) : string {
        switch (state) {
            case 0:
                return 'unknown';
            case 1:
                return 'shutdown';
            case 2:
                return 'remote disconnected';
            case 3:
                return 'booted';
            case 4:
                return 'locked';
        }
        return '';
    }

    function handleRunTests(id: number | null | undefined, e: MouseEvent<HTMLButtonElement>) : void {
        e.preventDefault();
        DeviceDataService.runTests(id).then(response => {
            console.log(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    }

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} size="small" aria-label="a dense table">
                <TableHead>
                    <TableRow>
                        <TableCell>Name</TableCell>
                        <TableCell align="right">Identifier</TableCell>
                        <TableCell align="right">OS</TableCell>
                        <TableCell align="right">Version</TableCell>
                        <TableCell align="right">Status</TableCell>
                        <TableCell align="right"></TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {devices.map((device) => <TableRow key={device.Name}>
                        <TableCell component="th" scope="row">
                            {device.Name}
                        </TableCell>
                        <TableCell align="right">{device.DeviceIdentifier}</TableCell>
                        <TableCell align="right">{device.OS}</TableCell>
                        <TableCell align="right">{device.OSVersion}</TableCell>
                        <TableCell align="right">{deviceState(device.Status)}</TableCell>
                        <TableCell align="right">
                            <Button color="primary" size="small" variant="outlined" endIcon={<PlayArrow />} onClick={(e) => handleRunTests(device.ID, e)}>
                                Run
                            </Button>
                        </TableCell>
                    </TableRow>)}
                </TableBody>
            </Table>
        </TableContainer>
    );
};

export default withStyles(styles)(DeviceTable);
