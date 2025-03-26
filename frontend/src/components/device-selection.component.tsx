import React, {ReactElement, useEffect} from 'react';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Checkbox from '@mui/material/Checkbox';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import IDeviceData from '../types/device';
import {Avatar, ListItemAvatar, ListItemButton, Typography} from '@mui/material';
import AndroidRoundedIcon from '@mui/icons-material/AndroidRounded';
import AppleIcon from '@mui/icons-material/Apple';
import Grid from "@mui/material/Grid2";

interface DeviceSelectionProps {
    devices: IDeviceData[],
    selectedDevices: number[];
    onSelectionChanged: (devices: number[]) => void;
}

function not(a: number[], b: number[]): number[] {
    return a.filter((value) => b.indexOf(value) === -1);
}

function intersection(a: number[], b: number[]): number[] {
    return a.filter((value) => b.indexOf(value) !== -1);
}

const DeviceSelection: React.FC<DeviceSelectionProps> = (props) => {
    const {devices, selectedDevices, onSelectionChanged} = props;

    type SelectionState = {
        left: number[],
        right: number[],
        checked: number[],
    }

    const deviceIDs = devices.map(dev => dev.ID) as number[];
    const left = not(deviceIDs, selectedDevices);

    const [state, setState] = React.useState<SelectionState>({
        left: left,
        right: selectedDevices,
        checked: [],
    });

    const handleToggle = (value: number) => () => {
        const currentIndex = state.checked.indexOf(value);
        if (currentIndex === -1) {
            setState(prevState => ({...prevState, checked: [...prevState.checked, value]}))
        } else {
            setState(prevState => ({
                ...prevState,
                checked: [...prevState.checked.slice(0, currentIndex), ...prevState.checked.slice(currentIndex + 1)]
            }))
        }
    };

    const handleAllRight = (): void => {
        setState(prevState => ({
            ...prevState,
            right: deviceIDs,
            left: []
        }))
    };

    const handleAllLeft = (): void => {
        setState(prevState => ({
            ...prevState,
            right: [],
            left: deviceIDs
        }))
    };

    const handleCheckedRight = (): void => {
        setState(prevState => ({
            ...prevState,
            right: [...prevState.right, ...prevState.left.filter(value => prevState.checked.findIndex(v => v === value) !== -1)],
            left: not(prevState.left, prevState.checked),
        }))
    };

    const handleCheckedLeft = (): void => {
        setState(prevState => ({
            ...prevState,
            right: not(prevState.right, prevState.checked),
            left: [...prevState.left, ...prevState.right.filter(value => prevState.checked.findIndex(v => v === value) !== -1)],
        }))
    };

    useEffect(() => {
        onSelectionChanged(state.right);
    }, [state.right, onSelectionChanged]);

    const customList = (items: number[]): ReactElement => (
        <Paper sx={{minWidth: 400, minHeight: 600, maxHeight: 600, margin: 'auto', overflow: 'auto'}}>
            <List dense={true} component="div" role="list">
                {items.map((value) => {
                    const labelId = `transfer-list-item-${value}-label`;
                    const device = devices.find(d => d.ID === value);
                    return (
                        <ListItem key={value}
                                  role="listitem"
                                  onClick={handleToggle(value)}
                                  secondaryAction={<Checkbox
                                      checked={state.checked.indexOf(value) !== -1}
                                      tabIndex={-1}
                                      disableRipple={true}
                                      inputProps={{'aria-labelledby': labelId}}
                                  />}
                        >
                            <ListItemButton>
                                <ListItemAvatar>
                                    <Avatar>
                                        {device?.OS === "android" && <AndroidRoundedIcon/>}
                                        {device?.OS === "iPhone OS" && <AppleIcon/>}
                                    </Avatar>
                                </ListItemAvatar>
                                <ListItemText id={labelId}
                                              primary={`${device && (device.Alias.length > 0 ? device.Alias : device.Name)} (${device?.OSVersion})`}
                                              secondary={device?.DeviceIdentifier}/>
                            </ListItemButton>
                        </ListItem>
                    );
                })}
                <ListItem/>
            </List>
        </Paper>
    );

    return (
        <Grid container={true}
              spacing={2}
              justifyContent={'center'}
              alignItems={'center'}>
            <Grid spacing={2} container={true} size={5}>
                <Grid size={12}>
                    <Typography variant={'subtitle1'}>Available Devices</Typography>
                </Grid>
                <Grid size={12}>
                    {customList(state.left)}
                </Grid>
            </Grid>
            <Grid spacing={2} container={true} size={1}>
                <Grid container={true} direction="column" alignItems="center" justifyContent={"center"}>
                    <Button
                        variant="outlined"
                        size="small"
                        onClick={handleAllRight}
                        disabled={state.left.length === 0}
                        aria-label="move all right"
                    >
                        ≫
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        onClick={handleCheckedRight}
                        disabled={state.left.length === 0}
                        aria-label="move selected right"
                    >
                        &gt;
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        onClick={handleCheckedLeft}
                        disabled={state.right.length === 0}
                        aria-label="move selected left"
                    >
                        &lt;
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        onClick={handleAllLeft}
                        disabled={state.right.length === 0}
                        aria-label="move all left"
                    >
                        ≪
                    </Button>
                </Grid>
            </Grid>
            <Grid spacing={2} container={true} size={5}>
                <Grid size={12}>
                    <Typography variant={'subtitle1'}>Selected Devices</Typography>
                </Grid>
                <Grid size={12}>
                    {customList(state.right)}
                </Grid>
            </Grid>
        </Grid>
    );
};

export default DeviceSelection;
