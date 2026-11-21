import React from 'react';
import {logStreamUrl} from '../services/device.control.service';
import {SseSourceFactory} from '../hooks/useSseLog';
import {DeviceLogPanel} from '../pages/devices/parts/DeviceLogPanel';

interface DeviceLogStreamProps {
    projectId: string,
    deviceId: number,
    maxHeight?: number | string,
    // opens the stream; tests inject a fake EventSource
    createSource?: SseSourceFactory,
}

// the most lines the device sends first
const FIRST_LINES = 300;

// DeviceLogStream follows the system log of a device (Android logcat) live.
// It needs no lease, so it also works while a test run holds the device.
const DeviceLogStream: React.FC<DeviceLogStreamProps> = ({projectId, deviceId, maxHeight = 420, createSource}) => {
    const url = logStreamUrl(projectId, deviceId, FIRST_LINES);
    // a panel per stream: a new device starts with a fresh view
    return <DeviceLogPanel key={url} url={url} maxHeight={maxHeight} createSource={createSource}/>;
};

export default DeviceLogStream;
