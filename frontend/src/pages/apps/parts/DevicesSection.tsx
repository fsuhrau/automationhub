import React from 'react';
import {Box, LinearProgress, Link, Stack, Typography} from '@mui/material';
import {CheckCircleRounded, ErrorOutlineRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {Section} from '../../../components/common';
import IDeviceData from '../../../types/device';
import {BundleInfo} from '../../../types/bundle.info';
import {deviceFit} from '../bundleDetailModel';

interface DeviceRowProps {
    info: BundleInfo,
    device: IDeviceData,
    to: string,
}

const DeviceRow: React.FC<DeviceRowProps> = ({info, device, to}) => {
    const fit = deviceFit(info, device);
    return <Stack direction="row" spacing={1} alignItems="flex-start">
        {fit.ok ? <CheckCircleRounded color="success" fontSize="small" aria-label="can install"/>
            : <ErrorOutlineRounded color="error" fontSize="small" aria-label="can't install"/>}
        <Box sx={{minWidth: 0}}>
            <Link component={RouterLink} to={to} variant="body2">{device.alias || device.name}</Link>
            <Typography component="span" variant="body2" color="text.secondary"> · {device.os} {device.osVersion}</Typography>
            {!fit.ok && <Typography variant="caption" color="error" sx={{display: 'block'}}>{fit.reasons.join('; ')}</Typography>}
        </Box>
    </Stack>;
};

interface DevicesSectionProps {
    info: BundleInfo,
    // null while loading
    devices: IDeviceData[] | null,
    deviceUrl: (id: number) => string,
}

const DeviceList: React.FC<DevicesSectionProps> = ({info, devices, deviceUrl}) => {
    if (devices === null) return <LinearProgress/>;
    if (devices.length === 0) return <Typography variant="body2" color="text.secondary">The pool has no device for this platform.</Typography>;
    return <Stack spacing={0.5}>
        {devices.map(d => <DeviceRow key={d.id} info={info} device={d} to={deviceUrl(d.id)}/>)}
    </Stack>;
};

export const DevicesSection: React.FC<DevicesSectionProps> = props =>
    <Section title="Devices" description="Which devices of the pool can install this bundle" fullHeight={true}>
        <DeviceList {...props}/>
    </Section>;
