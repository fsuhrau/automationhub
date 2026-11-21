import React from 'react';
import {Stack} from '@mui/material';
import Grid from '@mui/material/Grid';
import IDeviceData from '../../../types/device';
import {INodeData} from '../../../types/node';
import {Section} from '../../../components/common';
import DeviceLogStream from '../../../components/DeviceLogStream';
import {SseSourceFactory} from '../../../hooks/useSseLog';
import {DeviceRoutes} from '../../../lib/routes';
import {showsLog} from '../deviceDetailsModel';
import {RecentTests} from './RecentTests';
import {ConnectionSection, DeviceSection} from './DeviceInfoSections';
import {CustomSection, ReportedSection} from './ParameterSections';

interface DeviceDetailsColumnProps {
    projectId: string,
    device: IDeviceData,
    node?: INodeData,
    links: DeviceRoutes,
    createLogSource?: SseSourceFactory,
}

// DeviceDetailsColumn: what ran on the device, its log, and how it is set up
export const DeviceDetailsColumn: React.FC<DeviceDetailsColumnProps> = ({projectId, device, node, links, createLogSource}) =>
    <Stack spacing={2}>
        <RecentTests projectId={projectId} deviceId={device.id}/>
        {showsLog(device) && <Section title="Device log">
            <DeviceLogStream projectId={projectId} deviceId={device.id} createSource={createLogSource}/>
        </Section>}
        <Grid container={true} spacing={2}>
            <Grid size={{xs: 12, lg: 6}}><DeviceSection device={device}/></Grid>
            <Grid size={{xs: 12, lg: 6}}><ConnectionSection device={device} node={node}/></Grid>
            <Grid size={{xs: 12, lg: 6}}><ReportedSection parameters={device.deviceParameter ?? []}/></Grid>
            <Grid size={{xs: 12, lg: 6}}>
                <CustomSection parameters={device.customParameter ?? []} editUrl={links.edit()}/>
            </Grid>
        </Grid>
    </Stack>;
