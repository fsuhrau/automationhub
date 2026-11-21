import React from 'react';
import {Button, IconButton, Stack, Tooltip} from '@mui/material';
import {ArrowForward, LockOpen, PlayArrow, ScreenShare} from '@mui/icons-material';
import IDeviceData from '../../../types/device';

export interface RowHandlers {
    onSelectForRun: (deviceId: number) => void,
    onOpenDeviceDetails: (deviceId: number) => void,
    onOpenControl: (deviceId: number) => void,
    onUnlockDevice: (deviceId: number) => void,
}

interface ActionProps {
    title: string,
    label: string,
    color: 'primary' | 'warning' | 'default',
    icon: React.ReactElement,
    onClick: () => void,
}

// every action is labelled; the row itself opens the device details, so
// a click on an action must not reach the row
const ActionButton: React.FC<ActionProps> = ({title, label, color, icon, onClick}) =>
    <Tooltip title={title}>
        <IconButton size="small" color={color} aria-label={label} onClick={e => {
            e.stopPropagation();
            onClick();
        }}>
            {icon}
        </IconButton>
    </Tooltip>;

const RunButton: React.FC<{ onClick: () => void }> = ({onClick}) =>
    <Button color="primary" size="small" variant="outlined" startIcon={<PlayArrow/>}
        onClick={e => {
                e.stopPropagation();
                onClick();
            }}>
        Run
    </Button>;

// the actions of one device row
export const DeviceRowActions: React.FC<{ device: IDeviceData } & RowHandlers> = ({device, ...on}) =>
    <Stack direction="row" spacing={0.5} justifyContent="flex-end" alignItems="center" sx={{height: '100%'}}>
        {device.dev && <ActionButton title="Live control: see the screen and use the device" label="Live control"
            color="primary" icon={<ScreenShare fontSize="small"/>} onClick={() => on.onOpenControl(device.id)}/>}
        {device.isLocked && <ActionButton title="Unlock: end the test run or control session that holds the device"
            label="Unlock device" color="warning" icon={<LockOpen fontSize="small"/>}
            onClick={() => on.onUnlockDevice(device.id)}/>}
        {device.connection && <RunButton onClick={() => on.onSelectForRun(device.id)}/>}
        <ActionButton title="Details" label="Device details" color="default" icon={<ArrowForward fontSize="small"/>}
            onClick={() => on.onOpenDeviceDetails(device.id)}/>
    </Stack>;
