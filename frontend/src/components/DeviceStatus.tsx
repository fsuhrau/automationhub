import React from "react";
import {Box, Stack, Tooltip, Typography} from "@mui/material";
import InfoOutlined from "@mui/icons-material/InfoOutlined";
import {deviceState, DeviceStateType} from "../types/deviceStateType";
import {PlatformType} from "../types/platform.type.enum";

const iosUnauthorized = <>
    <b>This iPhone/iPad doesn't trust the computer yet, or is locked.</b>
    <ol style={{margin: '4px 0', paddingLeft: 18}}>
        <li>Unlock the device and tap "Trust" on "Trust This Computer?" (enter the passcode if asked).</li>
        <li>No prompt? Unplug and replug the cable.</li>
        <li>Still not trusted? On the device: Settings → General → Transfer or Reset → Reset → Reset Location &amp; Privacy, then reconnect.</li>
    </ol>
    The status updates shortly after.
</>;

// what a state means and how to get the device ready again
const hints: Partial<Record<DeviceStateType, React.ReactNode>> = {
    [DeviceStateType.Unauthorized]: <>
        <b>USB debugging is not authorized for this computer.</b>
        <ol style={{margin: '4px 0', paddingLeft: 18}}>
            <li>Unlock the phone and accept the prompt "Allow USB debugging?" (tick "Always allow from this computer").</li>
            <li>No prompt? Unplug and replug the cable, or on the phone choose Developer options → "Revoke USB debugging authorizations" and reconnect.</li>
            <li>Still unauthorized? Run <code>adb kill-server</code> on the node and reconnect.</li>
        </ol>
        The status updates shortly after.
    </>,
    [DeviceStateType.RemoteDisconnected]: <>
        <b>The node doesn't see this device.</b>
        <ol style={{margin: '4px 0', paddingLeft: 18}}>
            <li>Check that the phone is plugged in and unlocked, or that the emulator/simulator is running.</li>
            <li>Android: run <code>adb devices -l</code> on the node; the device should be listed as "device".</li>
            <li>iOS: the device should show up in Finder (or <code>idevice_id -l</code>) on the node.</li>
            <li>Network devices: check that the device is reachable from the node.</li>
        </ol>
    </>,
    [DeviceStateType.NodeDisconnected]: <>
        <b>The node of this device isn't connected to the master.</b> Start the node, or check its network and its
        master URL and token.
    </>,
    [DeviceStateType.Unknown]: <>The device state isn't known yet; it is updated with the next refresh.</>,
};

const DeviceStatus: React.FC<{ state: DeviceStateType, platform?: PlatformType }> = ({state, platform}) => {
    const isIOS = platform === PlatformType.iOS || platform === PlatformType.IOSSimulator;
    const hint = state === DeviceStateType.Unauthorized && isIOS ? iosUnauthorized : hints[state];
    const label = deviceState(state);
    if (!hint) {
        return <>{label}</>;
    }
    return (
        <Tooltip title={<Box sx={{maxWidth: 360}}>{hint}</Box>} arrow placement="left">
            <Stack direction="row" spacing={0.5} alignItems="center" component="span"
                   sx={{display: 'inline-flex', cursor: 'help', color: state === DeviceStateType.Unauthorized ? 'warning.main' : 'inherit'}}>
                <Typography variant="body2" component="span" color="inherit">{label}</Typography>
                <InfoOutlined fontSize="inherit"/>
            </Stack>
        </Tooltip>
    );
};

export default DeviceStatus;
