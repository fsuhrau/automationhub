import React from 'react';
import {Button, Stack, Tooltip} from '@mui/material';
import {EditRounded, TouchAppRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import IDeviceData from '../../../types/device';
import {INodeData} from '../../../types/node';
import {DeviceRoutes} from '../../../lib/routes';
import {Availability} from '../deviceDetailsModel';
import {DeviceTitle} from './DeviceTitle';
import {DeviceActionsMenu} from './DeviceActionsMenu';

interface DeviceHeaderProps {
    device: IDeviceData,
    node?: INodeData,
    avail: Availability,
    links: DeviceRoutes,
    onUnlock: () => void,
    onDelete: () => void,
}

const controlHint = (avail: Availability): string =>
    avail.canControl ? 'See the screen and use the device' : `Not possible now: ${avail.label.toLowerCase()}`;

// DeviceHeader: which device, whether it can be used, and what to do with it
export const DeviceHeader: React.FC<DeviceHeaderProps> = ({device, node, avail, links, onUnlock, onDelete}) =>
    <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
        <DeviceTitle device={device} node={node} avail={avail}/>
        <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {md: 'auto'}}}>
            <Tooltip title={controlHint(avail)}>
                <span>
                    <Button variant="contained" startIcon={<TouchAppRounded/>} disabled={!avail.canControl}
                        component={RouterLink} to={links.control()}>Live control</Button>
                </span>
            </Tooltip>
            <Button variant="outlined" startIcon={<EditRounded/>} component={RouterLink} to={links.edit()}>Edit</Button>
            <DeviceActionsMenu canUnlock={avail.canUnlock} onUnlock={onUnlock} onDelete={onDelete}/>
        </Stack>
    </Stack>;
