import React, {ChangeEvent, useState} from 'react';

import IDeviceData from '../types/device';
import {postUnlockDevice, runTest} from '../services/device.service';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    FormControlLabel,
    Stack,
    Switch,
    TextField,
    Typography,
} from '@mui/material';
import {useNavigate} from 'react-router-dom';
import _, {Dictionary} from "lodash";
import {TitleCard} from "./title.card.component";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useError} from "../ErrorProvider";
import DevicesTable from "./DevicesTable";
import {useHubState} from "../hooks/HubStateProvider";
import {HubStateActions} from "../application/HubState";
import {getPlatformName, PlatformType} from "../types/platform.type.enum";

// device platforms an app of the given platform runs on
const devicePlatformsFor = (appPlatform: PlatformType): PlatformType[] => {
    switch (appPlatform) {
        case PlatformType.iOS:
            return [PlatformType.iOS, PlatformType.IOSSimulator];
        default:
            return [appPlatform];
    }
};

const selectedAppId = (): number | null => {
    try {
        const value = localStorage.getItem('appId');
        return value ? Number(value) : null;
    } catch {
        return null;
    }
};

const NodeDevices: React.FC = () => {

    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()
    const navigate = useNavigate();
    const {state, dispatch} = useHubState()

    // dialog
    const [testName, seTestName] = useState<string>('');
    const [selectedDeviceID, setSelectedDeviceID] = useState<number>(0);
    const [envParameter, setEnvParameter] = useState<string>('');

    const [open, setOpen] = useState(false);
    const [showAll, setShowAll] = useState(false);

    // only devices matching the selected app's platform, unless "all" is on
    const selectedApp = state.apps?.find(a => a.id === selectedAppId());
    const platforms = selectedApp ? devicePlatformsFor(selectedApp.platform) : null;
    const devices = (state.devices ?? []).filter(d => showAll || platforms === null || platforms.includes(d.platformType));
    const handleClickOpen = (): void => {
        setOpen(true);
    };
    const handleClose = (): void => {
        setOpen(false);
    };

    const onRunTest = (): void => {
        runTest(projectIdentifier, selectedDeviceID, testName, envParameter).then(testRun => {
            console.log(testRun);
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
        postUnlockDevice(projectIdentifier, deviceId).then(device => {
            dispatch({type: HubStateActions.UnlockDevice, payload: {deviceId: deviceId}})
        }).catch(ex => setError(ex));
    }

    const openDetails = (id: number | null): void => {
        navigate(`/project/${projectIdentifier}/device/${id}`);
    };

    const getGroups = (devices: IDeviceData[]) => {
        return _.groupBy(devices, function (device) {
            return device.node ? `${device.node.name} (${state.nodes?.find(n => n.id === device.nodeId)?.status === 1 ? 'Connected' : 'Disconnected'})` : "Master";
        });
    }

    const getSortedGroups = (groups: Dictionary<IDeviceData[]>) => {
        return _.orderBy(Object.entries(groups), [
            ([group]) => group.includes('Connected') ? 0 : 1, // status: Connected first
            ([group]) => group.replace(/ \(.*\)$/, '') // name: Ascending
        ], ['asc', 'asc']);
    }

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
            {selectedApp && <Stack direction="row" alignItems="center" justifyContent="flex-end" spacing={2} sx={{px: 1}}>
                <Typography variant="body2" color="text.secondary">
                    {showAll ? 'All devices' : `${getPlatformName(selectedApp.platform)} devices for ${selectedApp.name}`}
                </Typography>
                <FormControlLabel control={<Switch size="small" checked={showAll} onChange={(_, v) => setShowAll(v)}/>}
                                  label="All platforms"/>
            </Stack>}
            {devices.length === 0 && <Typography variant="body2" sx={{p: 2}}>No matching devices.</Typography>}
            {
                getSortedGroups(getGroups(devices)).map(([group, items]) => {
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
