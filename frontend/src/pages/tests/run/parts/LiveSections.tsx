import React from 'react';
import {Button, Chip, Stack, Tooltip, Typography} from '@mui/material';
import {TouchAppRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import DeviceLiveScreen from '../../../../components/DeviceLiveScreen';
import DeviceLogStream from '../../../../components/DeviceLogStream';
import {Section} from '../../../../components/common';
import {routes} from '../../../../lib/routes';
import IDeviceData from '../../../../types/device';
import {PlatformType} from '../../../../types/platform.type.enum';
import {deviceLabel} from '../runStatusModel';
import {LiveDevices} from '../useLiveDevices';

// the system log of the watched device; only Android has one
export const DeviceLogSection: React.FC<{ projectId: string, device: IDeviceData | undefined }> = ({projectId, device}) => {
    if (device?.platformType !== PlatformType.Android) return null;
    return <Section title={`Device log · ${deviceLabel(device)}`} fullHeight={true}>
        <DeviceLogStream key={device.id} projectId={projectId} deviceId={device.id} maxHeight="45vh"/>
    </Section>;
};

const OpenControl: React.FC<{ to: string }> = ({to}) =>
    <Tooltip title="Open the device's live control page. You can watch it there; control is possible once the run is over.">
        <Button size="small" startIcon={<TouchAppRounded/>} component={RouterLink} to={to}>Open</Button>
    </Tooltip>;

const DevicePicker: React.FC<{ live: LiveDevices, watched: IDeviceData }> = ({live, watched}) =>
    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap={true} sx={{mb: 1.5}} role="group" aria-label="Device to watch">
        {live.devices.map(d => <Chip key={d.id} size="small" label={deviceLabel(d)}
            color={d.id === watched.id ? 'primary' : 'default'}
            variant={d.id === watched.id ? 'filled' : 'outlined'}
            onClick={() => live.watch(d.id)}/>)}
    </Stack>;

const Screen: React.FC<{ projectId: string, live: LiveDevices, watched: IDeviceData }> = ({projectId, live, watched}) =>
    <>
        {live.devices.length > 1 && <DevicePicker live={live} watched={watched}/>}
        <DeviceLiveScreen projectId={projectId} deviceId={watched.id} platformType={watched.platformType} label={deviceLabel(watched)}/>
        {live.devices.length === 1 && <Typography variant="caption" color="text.secondary" component="div" sx={{mt: 1}}>
            {deviceLabel(watched)} · {watched.os} {watched.osVersion}
        </Typography>}
    </>;

const WaitingForDevice: React.FC<{ knowsHeld: boolean }> = ({knowsHeld}) =>
    <Typography variant="body2" color="text.secondary">
        {knowsHeld ? 'Waiting for the run to take a device…' : 'The screen shows once a device has started the app.'}
    </Typography>;

// the screens of the devices the run holds, while it runs
export const LiveScreenSection: React.FC<{ projectId: string, live: LiveDevices }> = ({projectId, live}) => {
    const {watched} = live;
    const control = watched && <OpenControl to={routes.project(projectId).device(watched.id).control()}/>;
    return <Section title="Live screen" fullHeight={true} action={control}>
        {watched ? <Screen projectId={projectId} live={live} watched={watched}/> : <WaitingForDevice knowsHeld={live.knowsHeld}/>}
    </Section>;
};
