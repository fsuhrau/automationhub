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
import { getAllDevices, runTests } from '../services/device.service';
import { Button } from '@material-ui/core';
import { PlayArrow } from '@material-ui/icons';
import { useSSE } from 'react-hooks-sse';

const styles = (): ReturnType<typeof createStyles> =>
    createStyles({
        table: {
            minWidth: 650,
        },
    });

export type DeviceProps = WithStyles<typeof styles>;

interface DeviceChangePayload {
    DeviceID: number,
    DeviceState: number
}

const DeviceTable: FC<DeviceProps> = (props) => {
    const { classes } = props;
    const [devices, setDevices] = useState<IDeviceData[]>([]);

    const deviceStateChange = useSSE<DeviceChangePayload | null>('devices', null);

    useEffect(() => {
        if (deviceStateChange === null)
            return;
        console.log(deviceStateChange);
        setDevices(previousDevices => previousDevices.map(device => device.ID === deviceStateChange.DeviceID ? { ...device, Status: deviceStateChange.DeviceState } : device));
    }, [deviceStateChange]);

    useEffect(() => {
        getAllDevices().then(response => {
            setDevices(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    function deviceState(state: number): string {
        switch (state) {
            case 0:
                return 'null';
            case 1:
                return 'unknown';
            case 2:
                return 'shutdown';
            case 3:
                return 'remote disconnected';
            case 4:
                return 'booted';
            case 5:
                return 'locked';
        }
        return '';
    }

    function handleRunTests(id: number | null | undefined, e: MouseEvent<HTMLButtonElement>): void {
        e.preventDefault();
        runTests(id).then(response => {
            console.log(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    }

    return (
        <TableContainer component={ Paper }>
            <Table className={ classes.table } size="small" aria-label="a dense table">
                <TableHead>
                    <TableRow>
                        <TableCell>ID</TableCell>
                        <TableCell>Name</TableCell>
                        <TableCell>Identifier</TableCell>
                        <TableCell>OS</TableCell>
                        <TableCell align="right">Version</TableCell>
                        <TableCell>Status</TableCell>
                        <TableCell>Session</TableCell>
                        <TableCell align="right">Actions</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    { devices.map((device) => <TableRow key={ device.Name }>
                        <TableCell component="th" scope="row">
                            { device.ID }
                        </TableCell>
                        <TableCell>
                            { device.Name }
                        </TableCell>
                        <TableCell>{ device.DeviceIdentifier }</TableCell>
                        <TableCell>{ device.OS }</TableCell>
                        <TableCell align="right">{ device.OSVersion }</TableCell>
                        <TableCell>{ deviceState(device.Status) }</TableCell>
                        <TableCell>{ device.Connection?.appID }</TableCell>
                        <TableCell align="right">
                            <Button color="primary" size="small" variant="outlined" endIcon={ <PlayArrow/> }
                                onClick={ (e) => handleRunTests(device.ID, e) }>
                                Run
                            </Button>
                        </TableCell>
                    </TableRow>) }
                </TableBody>
            </Table>
        </TableContainer>
    );
};

export default withStyles(styles)(DeviceTable);
