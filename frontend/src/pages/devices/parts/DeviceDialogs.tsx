import React from 'react';
import IDeviceData from '../../../types/device';
import {ConfirmDialog} from '../../../components/common';
import {describeController} from '../../../components/DeviceController';

export type DeviceConfirm = 'delete' | 'unlock';

interface DeviceDialogsProps {
    device: IDeviceData,
    confirm: DeviceConfirm | null,
    busy: boolean,
    onConfirm: (what: DeviceConfirm) => void,
    onClose: () => void,
}

const unlockText = (device: IDeviceData): string => device.lockedBy
    ? `This ends the live control of ${describeController(device.lockedBy).label}.`
    : 'A test run still on this device loses it and fails. Unlock only when the run is stuck.';

// the questions before deleting or unlocking the device
export const DeviceDialogs: React.FC<DeviceDialogsProps> = ({device, confirm, busy, onConfirm, onClose}) => {
    const name = device.alias || device.name;
    return <>
        <ConfirmDialog open={confirm === 'delete'} title={`Delete ${name}?`} confirm="Delete device" busy={busy}
            onConfirm={() => onConfirm('delete')} onClose={onClose}>
            The device leaves the device pool, with its alias and custom parameters. Its test results stay.
            A device that is still connected can be detected again.
        </ConfirmDialog>
        <ConfirmDialog open={confirm === 'unlock'} title={`Unlock ${name}?`} confirm="Unlock" color="warning" busy={busy}
            onConfirm={() => onConfirm('unlock')} onClose={onClose}>
            {unlockText(device)}
        </ConfirmDialog>
    </>;
};
