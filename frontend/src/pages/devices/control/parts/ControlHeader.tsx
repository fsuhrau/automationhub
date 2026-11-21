import React from 'react';
import {Link as RouterLink} from 'react-router-dom';
import {Box, Chip, Link, Stack, Typography} from '@mui/material';
import IDeviceData from '../../../../types/device';
import {routes} from '../../../../lib/routes';
import {DeviceTarget} from '../../../../types/device.control.api';
import {ControlStatus} from '../statusModel';
import {ControlSession} from '../useControlSession';
import {ReleaseButton, TakeControlButton} from './LeaseButtons';

const DeviceSubtitle: React.FC<{ device: IDeviceData | null, target: DeviceTarget }> = ({device, target}) => (
    <Typography variant="body2" color="text.secondary">
        {device ? <>{device.os} {device.osVersion}
            {device.alias && device.alias !== device.name && <> · {device.name}</>}
            {device.node && <> · on {device.node.name}</>}</> : ' '}
        {' · '}
        <Link component={RouterLink} to={routes.project(target.projectId).device(target.deviceId).show()}
            underline="hover" color="inherit">Device details</Link>
    </Typography>
);

const StatusChip: React.FC<{ status: ControlStatus }> = ({status}) => (
    <Box sx={{textAlign: {md: 'right'}}}>
        <Chip size="small" color={status.color} label={status.label}/>
        {status.note && <Typography variant="caption" color="text.secondary" component="div">{status.note}</Typography>}
    </Box>
);

// which device, who controls it, and taking or giving back control
const ControlHeader: React.FC<{ session: ControlSession }> = ({session}) => {
    const {device, target, status, busy, name} = session;
    return (
        <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
            <Box sx={{minWidth: 0}}>
                <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{name}</Typography>
                <DeviceSubtitle device={device} target={target}/>
            </Box>
            <Stack direction="row" spacing={1.5} alignItems="center" sx={{ml: {md: 'auto'}}} useFlexGap={true} flexWrap="wrap">
                <StatusChip status={status}/>
                {status.own
                    ? <ReleaseButton onClick={session.onRelease} disabled={busy}/>
                    : <TakeControlButton onClick={session.onLock} disabled={busy || !!session.lease.status?.locked}/>}
            </Stack>
        </Stack>
    );
};

export default ControlHeader;
