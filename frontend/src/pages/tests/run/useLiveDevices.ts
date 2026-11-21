import {useMemo, useState} from 'react';
import {useHubState} from '../../../hooks/HubStateProvider';
import IDeviceData from '../../../types/device';
import ITestProtocolData from '../../../types/test.protocol';
import ITestRunData from '../../../types/test.run';
import {TestRef} from '../../../types/test.run.refs';
import {liveDevices} from './liveDevicesModel';
import {useHeldDevices} from './useHeldDevices';

export interface LiveDevices {
    devices: IDeviceData[],
    // the device whose screen and log are shown
    watched: IDeviceData | undefined,
    watch: (deviceId: number) => void,
    // the master says which devices the run holds
    knowsHeld: boolean,
}

export interface LiveDevicesOptions {
    run: ITestRunData,
    protocols: ITestProtocolData[],
    running: boolean,
    test: TestRef,
    // the devices the run held when the page loaded; undefined on older masters
    initialHeld?: number[],
}

// useLiveDevices lists the devices a running run holds and which of them is
// watched; none while the run is not running
export const useLiveDevices = ({run, protocols, running, test, initialHeld}: LiveDevicesOptions): LiveDevices => {
    const {state} = useHubState();
    const held = useHeldDevices(initialHeld, running, {...test, runId: run.id as number});
    const [picked, setPicked] = useState<number | null>(null);
    const devices = useMemo(() => running
        ? liveDevices({held, protocols, deviceStatus: run.deviceStatus, hubDevices: state.devices})
        : [], [running, held, protocols, run.deviceStatus, state.devices]);
    return {devices, watched: devices.find(d => d.id === picked) ?? devices[0], watch: setPicked, knowsHeld: initialHeld !== undefined};
};
