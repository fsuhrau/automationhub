import React from 'react';
import {Alert, Box, Button, Checkbox, CircularProgress, FormControlLabel, Stack, Typography} from '@mui/material';
import DeviceStatus from '../../../../components/DeviceStatus';
import IDeviceData from '../../../../types/device';

interface DevicePickerProps {
    // null while loading
    devices: IDeviceData[] | null,
    picked: number[],
    onChange: (picked: number[]) => void,
    appName?: string,
}

const DeviceLabel: React.FC<{ device: IDeviceData }> = ({device: d}) =>
    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true}>
        <Typography variant="body2" sx={{fontWeight: 600}}>{d.alias || d.name}</Typography>
        <Typography variant="caption" color="text.secondary">{d.os} {d.osVersion}</Typography>
        <Typography variant="caption" component="span" color="text.secondary">
            <DeviceStatus state={d.status} platform={d.platformType}/>
        </Typography>
    </Stack>;

const DeviceChecklist: React.FC<Omit<DevicePickerProps, 'appName'> & { devices: IDeviceData[] }> = ({devices, picked, onChange}) =>
    <Box sx={{border: 1, borderColor: 'divider', borderRadius: 1, p: 1, maxHeight: 320, overflowY: 'auto'}} role="group" aria-label="Devices">
        {devices.map(d => <FormControlLabel key={d.id} sx={{display: 'flex', mr: 0}} label={<DeviceLabel device={d}/>}
            control={<Checkbox size="small" checked={picked.includes(d.id)}
                onChange={(_, on) => onChange(on ? [...picked, d.id] : picked.filter(x => x !== d.id))}/>}/>)}
    </Box>;

const Devices: React.FC<DevicePickerProps> = ({devices, appName, ...rest}) => {
    if (devices === null) return <CircularProgress size={18}/>;
    if (devices.length === 0) return <Alert severity="warning">No {appName ? `${appName} ` : ''}devices are known to the hub yet.</Alert>;
    return <DeviceChecklist devices={devices} {...rest}/>;
};

// devices picked earlier that the list doesn't show (other platform, gone)
const MissingDevices: React.FC<DevicePickerProps> = ({devices, picked, onChange}) => {
    const missing = picked.filter(id => devices && !devices.some(d => d.id === id));
    if (missing.length === 0) return null;
    return <Alert severity="info" sx={{mt: 1}}
        action={<Button color="inherit" size="small" onClick={() => onChange(picked.filter(id => !missing.includes(id)))}>Remove</Button>}>
        {missing.length === 1 ? 'One picked device isn\'t' : `${missing.length} picked devices aren't`} in the list (another platform, or deleted).
    </Alert>;
};

// DevicePicker checks the devices of the app's platform to run on
export const DevicePicker: React.FC<DevicePickerProps & { error?: string }> = ({error, ...props}) =>
    <Box sx={{mt: 1, mb: 2}}>
        {error && <Typography variant="body2" color="error" sx={{mb: 1}}>{error}</Typography>}
        <Devices {...props}/>
        <MissingDevices {...props}/>
    </Box>;
