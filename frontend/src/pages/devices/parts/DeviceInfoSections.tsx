import React from 'react';
import {Box, Stack, Typography} from '@mui/material';
import IDeviceData from '../../../types/device';
import {INodeData} from '../../../types/node';
import {getPlatformName} from '../../../types/platform.type.enum';
import {DeviceConnectionType} from '../../../types/device.connection.type.enum';
import {CopyButton, KeyValues, Section} from '../../../components/common';
import DeviceStatus from '../../../components/DeviceStatus';
import {formatDateTime} from '../../../components/DateTime';
import {CONNECTIONS, connectionAddress, DEVICE_TYPES} from '../deviceDetailsModel';

type Row = [string, React.ReactNode];

const Identifier: React.FC<{ value: string }> = ({value}) =>
    <Stack direction="row" alignItems="center" spacing={0.5} component="span">
        <Box component="span" sx={{fontFamily: 'monospace', wordBreak: 'break-all'}}>{value}</Box>
        <CopyButton text={value} label="Copy identifier"/>
    </Stack>;

const deviceRows = (device: IDeviceData): Row[] => [
    ['Name', device.name],
    ...(device.alias ? [['Alias', device.alias] as Row] : []),
    ['Identifier', <Identifier value={device.deviceIdentifier}/>],
    ['Platform', getPlatformName(device.platformType)],
    ['System', `${device.os} ${device.osVersion}`],
    ['Type', DEVICE_TYPES[device.deviceType] ?? '–'],
    ['Status', <DeviceStatus state={device.status} platform={device.platformType}/>],
    ['Hub id', String(device.id)],
    ...(device.createdAt ? [['Added', formatDateTime(device.createdAt, 'YYYY/MM/DD')] as Row] : []),
];

const NodeInfo: React.FC<{ node?: INodeData }> = ({node}) => node
    ? <>{node.name} <Typography component="span" variant="caption" color="text.secondary">
        ({node.status === 1 ? 'connected' : 'offline'})
    </Typography></>
    : <>the master itself</>;

// devices found by a node have no connection settings of their own
const connectionRows = (device: IDeviceData, node?: INodeData): Row[] => {
    const connection = device.connectionParameter;
    const address = connectionAddress(connection);
    return [
        ...(connection ? [['Connected', CONNECTIONS[connection.connectionType as DeviceConnectionType] ?? '–'] as Row] : []),
        ['Node', <NodeInfo node={node}/>],
        ...(address ? [['Address', address] as Row] : []),
        ...(device.connection?.version ? [['App connected', `version ${device.connection.version}`] as Row] : []),
    ];
};

export const DeviceSection: React.FC<{ device: IDeviceData }> = ({device}) =>
    <Section title="Device"><KeyValues rows={deviceRows(device)}/></Section>;

export const ConnectionSection: React.FC<{ device: IDeviceData, node?: INodeData }> = ({device, node}) =>
    <Section title="Connection"><KeyValues rows={connectionRows(device, node)}/></Section>;
