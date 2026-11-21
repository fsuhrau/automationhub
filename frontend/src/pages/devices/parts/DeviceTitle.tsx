import React from 'react';
import {Box, Chip, Stack, Tooltip, Typography} from '@mui/material';
import IDeviceData from '../../../types/device';
import {INodeData} from '../../../types/node';
import PlatformTypeIcon from '../../../components/PlatformTypeIcon';
import {Availability, DEVICE_TYPES} from '../deviceDetailsModel';

interface DeviceTitleProps {
    device: IDeviceData,
    node?: INodeData,
    avail: Availability,
}

// DeviceTitle says which device it is and whether it can be used now
export const DeviceTitle: React.FC<DeviceTitleProps> = ({device, node, avail}) =>
    <Stack direction="row" spacing={1.5} alignItems="center" sx={{minWidth: 0}}>
        <Box sx={{display: 'flex', p: 1, borderRadius: 1, border: 1, borderColor: 'divider'}}>
            <PlatformTypeIcon platformType={device.platformType}/>
        </Box>
        <Box sx={{minWidth: 0}}>
            <Stack direction="row" spacing={1.5} alignItems="center" flexWrap="wrap" useFlexGap={true}>
                <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{device.alias || device.name}</Typography>
                <Tooltip title={avail.detail}><Chip size="small" color={avail.tone} label={avail.label}/></Tooltip>
            </Stack>
            <Typography variant="body2" color="text.secondary">
                {DEVICE_TYPES[device.deviceType] ?? 'Device'} · {device.os} {device.osVersion}
                {node && <> · on {node.name}</>}
            </Typography>
        </Box>
    </Stack>;
