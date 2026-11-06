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
    Chip,
    Stack,
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
import PlatformTypeIcon from "./PlatformTypeIcon";

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

    // the device pool belongs to the project, not to an app, so the filter is
    // on the page and visible instead of following the app picked in the menu
    const [platform, setPlatform] = useState<PlatformType | 'all'>('all');
    const allDevices = state.devices ?? [];
    const platformCounts = _.countBy(allDevices, d => d.platformType);
    const platforms = Object.keys(platformCounts).map(Number).sort((a, b) => a - b) as PlatformType[];
    const devices = allDevices.filter(d => platform === 'all' || d.platformType === platform);
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
            {platforms.length > 1 && <Stack direction="row" alignItems="center" spacing={1} useFlexGap
                                           sx={{flexWrap: 'wrap'}} role="group" aria-label="Filter by platform">
                <Chip label={`All (${allDevices.length})`} size="small"
                      color={platform === 'all' ? 'primary' : 'default'}
                      variant={platform === 'all' ? 'filled' : 'outlined'}
                      onClick={() => setPlatform('all')}/>
                {platforms.map(p => <Chip key={p} size="small"
                                          icon={<PlatformTypeIcon platformType={p} fontSize="small"/>}
                                          label={`${getPlatformName(p)} (${platformCounts[p]})`}
                                          color={platform === p ? 'primary' : 'default'}
                                          variant={platform === p ? 'filled' : 'outlined'}
                                          onClick={() => setPlatform(p)}/>)}
            </Stack>}
            {state.devices !== null && allDevices.length === 0 &&
                <Typography variant="body2" color="text.secondary" sx={{p: 2}}>
                    No devices yet. Devices show up here when a node or the master detects them
                    (for example an Android phone connected over USB).
                </Typography>}
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
