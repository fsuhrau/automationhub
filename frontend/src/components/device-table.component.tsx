import React, {ChangeEvent, useEffect, useState} from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';

import IDeviceData from '../types/device';
import {getAllDevices, postUnlockDevice, runTest} from '../services/device.service';
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
import {ArrowForward, PlayArrow} from '@mui/icons-material';
import {useSSE} from 'react-hooks-sse';
import {useNavigate} from 'react-router-dom';
import _ from "lodash";
import {TitleCard} from "./title.card.component";
import Grid from "@mui/material/Grid";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useError} from "../ErrorProvider";

interface DeviceChangePayload {
    DeviceID: number,
    DeviceState: number
}

export const deviceState = (state: number): string => {
    switch (state) {
        case 0:
            return 'null';
        case 1:
            return 'unknown';
        case 2:
            return 'shutdown';
        case 3:
            return 'disconnected';
        case 4:
            return 'booted';
        case 5:
            return 'locked';
    }
    return '';
};

const DeviceTable: React.FC = () => {

    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()
    const navigate = useNavigate();

    const [devices, setDevices] = useState<IDeviceData[]>([]);
    const deviceStateChange = useSSE<DeviceChangePayload | null>('devices', null);

    useEffect(() => {
        if (deviceStateChange === null)
            return;
        setDevices(previousDevices => previousDevices.map(device => device.ID === deviceStateChange.DeviceID ? {
            ...device,
            Status: deviceStateChange.DeviceState,
        } : device));
    }, [deviceStateChange]);

    useEffect(() => {
        getAllDevices(projectIdentifier).then(response => {
            setDevices(response.data);
        }).catch(e => {
            setError(e);
        });
    }, [projectIdentifier]);

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
        runTest(projectIdentifier, selectedDeviceID, testName, envParameter).then(response => {
            console.log(response.data);
        }).catch(ex => {
            setError(ex);
        });
    };

    const onEnvParamsChanged = (event: ChangeEvent<HTMLInputElement>): void => {
        setEnvParameter(event.target.value);
    };

    const onTestNameChanged = (event: ChangeEvent<HTMLInputElement>): void => {
        seTestName(event.target.value);
    };

    const unlockDevice = (deviceId: string) => {
        postUnlockDevice(projectIdentifier, deviceId).then(response => {
            setDevices(devices.map(d => d.DeviceIdentifier === deviceId ? response.data : d) as IDeviceData[]);
        }).catch(ex => setError(ex));
    }

    const openDetails = (id: number): void => {
        navigate(`/project/${projectIdentifier}/device/${id}`);
    };

    const groups = _.groupBy(devices, function (device) {
        return device.Node ? device.Node.Name : "Master";
    });
    return (
        <>
            <Dialog open={open} onClose={handleClose} aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">Run Test</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Enter Test to execute
                    </DialogContentText>
                    <Typography variant={'subtitle1'}>
                        Parameter:
                    </Typography>
                    <Typography variant={'subtitle2'}>
                        server=http://localhost:8080<br/>
                        user=autohub
                    </Typography>
                    <br/>
                    <TextField
                        id="outlined-multiline-static"
                        placeholder="Parameter"
                        fullWidth={true}
                        multiline={true}
                        rows={4}
                        defaultValue=""
                        variant="outlined"
                        onChange={onEnvParamsChanged}
                    />
                    <Typography variant={'subtitle1'}>
                        Test:
                    </Typography>
                    <br/>
                    <TextField
                        id="outlined-static"
                        placeholder="Parameter"
                        fullWidth={true}
                        defaultValue=""
                        variant="outlined"
                        onChange={onTestNameChanged}
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose} color="primary">
                        Cancel
                    </Button>
                    <Button onClick={() => {
                        onRunTest();
                        handleClose();
                    }} color="primary" variant={'contained'}>
                        Start
                    </Button>
                </DialogActions>
            </Dialog>
            {
                _.map(groups, (items, group) => {
                    return (
                        <TitleCard key={`device_table_group_${group}`}
                                   title={`Node: ${group === "null" ? "Master" : group}`}>
                            <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                                <Grid container={true}>
                                    <Grid item={true} xs={12}>
                                        <Table size="small">
                                            <TableHead>
                                                <TableRow>
                                                    <TableCell>ID</TableCell>
                                                    <TableCell>Name/Identifier</TableCell>
                                                    <TableCell>OS</TableCell>
                                                    <TableCell align="right">Version</TableCell>
                                                    <TableCell>Status</TableCell>
                                                    <TableCell>Session</TableCell>
                                                    <TableCell align="right">Actions</TableCell>
                                                </TableRow>
                                            </TableHead>
                                            <TableBody>
                                                {items.map((device) => <TableRow key={`device_entry_${device.ID}`}>
                                                    <TableCell component="th" scope="row">
                                                        {device.ID}
                                                    </TableCell>
                                                    <TableCell>
                                                        {device.Alias.length > 0 ? device.Alias : device.Name}<br/>
                                                        {device.DeviceIdentifier}
                                                    </TableCell>
                                                    <TableCell>{device.OS}</TableCell>
                                                    <TableCell align="right">{device.OSVersion}</TableCell>
                                                    <TableCell>{deviceState(device.Status)}
                                                        {device.Status === 5 && <Button
                                                            onClick={() => unlockDevice(device.DeviceIdentifier)}>Unlock</Button>}
                                                    </TableCell>
                                                    <TableCell>{device.Connection?.appID}</TableCell>
                                                    <TableCell align="right">
                                                        {device.Connection && (
                                                            <Button color="primary" size="small" variant="outlined"
                                                                    endIcon={<PlayArrow/>}
                                                                    onClick={(e) => {
                                                                        setSelectedDeviceID(device.ID as number);
                                                                        handleClickOpen();
                                                                    }}>
                                                                Run
                                                            </Button>)}
                                                        <IconButton color="primary" size={'small'}
                                                                    onClick={(e) => {
                                                                        openDetails(device.ID);
                                                                    }}>
                                                            <ArrowForward/>
                                                        </IconButton>
                                                    </TableCell>
                                                </TableRow>)}
                                            </TableBody>
                                        </Table>
                                    </Grid>
                                </Grid>
                            </Paper>
                        </TitleCard>)
                })
            }
        </>
    );
};

export default DeviceTable;
