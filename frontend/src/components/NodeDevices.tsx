import React, {ChangeEvent, useEffect, useState} from 'react';

import IDeviceData from '../types/device';
import {getAllDevices, postUnlockDevice, runTest} from '../services/device.service';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    TextField,
    Typography,
} from '@mui/material';
import {useSSE} from 'react-hooks-sse';
import {useNavigate} from 'react-router-dom';
import _ from "lodash";
import {TitleCard} from "./title.card.component";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useError} from "../ErrorProvider";
import DevicesTable from "./DevicesTable";
import {useHubState} from "../hooks/HubStateProvider";

interface DeviceChangePayload {
    DeviceID: number,
    DeviceState: number
}

const NodeDevices: React.FC = () => {

    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()
    const navigate = useNavigate();
    const {state} = useHubState()

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

    const unlockDevice = (deviceId: number) => {
        postUnlockDevice(projectIdentifier, deviceId).then(response => {
            setDevices(devices.map(d => d.ID === deviceId ? response.data : d) as IDeviceData[]);
        }).catch(ex => setError(ex));
    }

    const openDetails = (id: number | null): void => {
        navigate(`/project/${projectIdentifier}/device/${id}`);
    };

    const groups = _.groupBy(devices, function (device) {
        return device.Node ? `${device.Node.Name} (${state.nodes?.find(n => n.ID === device.NodeID)?.Status === 1 ? 'Connected' : 'Disconnected'})`: "Master";
    });

    const onDeviceSelected = (deviceId: number | null) => {
        setSelectedDeviceID(deviceId as number)
        handleClickOpen()
    }
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
                            <DevicesTable devices={items} onOpenDeviceDetails={openDetails}
                                          onSelectForRun={onDeviceSelected} onUnlockDevice={unlockDevice}/>
                        </TitleCard>)
                })
            }
        </>
    );
};

export default NodeDevices;
