import React from 'react';
import {Alert, Button, Stack} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import IDeviceData from '../../../types/device';
import {DeviceRoutes} from '../../../lib/routes';
import DeviceStatus from '../../../components/DeviceStatus';
import {alertSeverity, Availability} from '../deviceDetailsModel';

interface DeviceAlertsProps {
    device: IDeviceData,
    avail: Availability,
    links: DeviceRoutes,
    onUnlock: () => void,
}

const ControlledByYou: React.FC<{ detail: string, links: DeviceRoutes }> = ({detail, links}) =>
    <Alert severity="info" sx={{mb: 2}}
        action={<Button color="inherit" size="small" component={RouterLink} to={links.control()}>Live control</Button>}>
        {detail}
    </Alert>;

const Problem: React.FC<Omit<DeviceAlertsProps, 'links'>> = ({device, avail, onUnlock}) =>
    <Alert severity={alertSeverity(avail.tone)} sx={{mb: 2}}
        action={avail.canUnlock && <Button color="inherit" size="small" onClick={onUnlock}>Unlock</Button>}>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true}>
            <span>{avail.detail}</span>
            {!avail.canUnlock && <DeviceStatus state={device.status} platform={device.platformType} label="How to fix"/>}
        </Stack>
    </Alert>;

// DeviceAlerts explains why a device can't be used, or that you control it
export const DeviceAlerts: React.FC<DeviceAlertsProps> = ({device, avail, links, onUnlock}) =>
    <>
        {avail.label === 'Controlled by you' && <ControlledByYou detail={avail.detail} links={links}/>}
        {avail.tone !== 'success' && <Problem device={device} avail={avail} onUnlock={onUnlock}/>}
    </>;
