import { ChangeEvent, FC, useEffect, useState } from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';

import IDeviceData from '../types/device';
import { getAllDevices, runTest } from '../services/device.service';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    IconButton,
    TextField,
    Typography,
} from '@mui/material';
import { ArrowForward, PlayArrow } from '@mui/icons-material';
import { useSSE } from 'react-hooks-sse';
import { useHistory } from 'react-router-dom';

interface DeviceChangePayload {
    DeviceID: number,
    DeviceState: number
}

const DeviceTable: FC = () => {
    const history = useHistory();

    const [devices, setDevices] = useState<IDeviceData[]>([]);

    const deviceStateChange = useSSE<DeviceChangePayload | null>('devices', null);

    useEffect(() => {
        if (deviceStateChange === null)
            return;
        console.log(deviceStateChange);
        setDevices(previousDevices => previousDevices.map(device => device.ID === deviceStateChange.DeviceID ? {
            ...device,
            Status: deviceStateChange.DeviceState,
        } : device));
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

    // dialog
    const [testName, seTestName] = useState<string>('');
    const [selectedDeviceID, setSelectedDeviceID] = useState<number>(0);
    const [envParameter, setEnvParameter] = useState<string>('');

    const [open, setOpen] = useState(false);
    const handleClickOpen = (): void => {
        setOpen(true);
    };
    const handleClose = (): void => {
        setOpen(false);
    };

    const onRunTest = (): void => {
        runTest(selectedDeviceID, testName, envParameter).then(response => {
            console.log(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    };

    const onEnvParamsChanged = (event: ChangeEvent<HTMLInputElement>): void => {
        setEnvParameter(event.target.value);
    };

    const onTestNameChanged = (event: ChangeEvent<HTMLInputElement>): void => {
        seTestName(event.target.value);
    };

    function openDetails(id: number): void {
        history.push(`/web/device/${ id }`);
    }

    return (
        <div>
            <Dialog open={ open } onClose={ handleClose } aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Run Test</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Enter Test to execute
                    </DialogContentText>
                    <Typography variant={ 'subtitle1' }>
                        Parameter:
                    </Typography>
                    <Typography variant={ 'subtitle2' }>
                        server=http://localhost:8080<br/>
                        user=autohub
                    </Typography>
                    <br/>
                    <TextField
                        id="outlined-multiline-static"
                        label="Parameter"
                        fullWidth={ true }
                        multiline={ true }
                        rows={ 4 }
                        defaultValue=""
                        variant="outlined"
                        onChange={ onEnvParamsChanged }
                    />
                    <Typography variant={ 'subtitle1' }>
                        Test:
                    </Typography>
                    <br/>
                    <TextField
                        id="outlined-static"
                        label="Parameter"
                        fullWidth={ true }
                        defaultValue=""
                        variant="outlined"
                        onChange={ onTestNameChanged }
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={ handleClose } color="primary">
                        Cancel
                    </Button>
                    <Button onClick={ () => {
                        onRunTest();
                        handleClose();
                    } } color="primary" variant={ 'contained' }>
                        Start
                    </Button>
                </DialogActions>
            </Dialog>
            <TableContainer component={ Paper }>
                <Table size="small" aria-label="a dense table">
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
                                { device.Connection && (
                                    <Button color="primary" size="small" variant="outlined" endIcon={ <PlayArrow/> }
                                        onClick={ (e) => {
                                            setSelectedDeviceID(device.ID as number);
                                            handleClickOpen();
                                        } }>
                                        Run
                                    </Button>) }
                                <IconButton color="primary" size={ 'small' }
                                    onClick={ (e) => {
                                        openDetails(device.ID);
                                    } }>
                                    <ArrowForward/>
                                </IconButton>
                            </TableCell>
                        </TableRow>) }
                    </TableBody>
                </Table>
            </TableContainer>
        </div>
    );
};

export default DeviceTable;
