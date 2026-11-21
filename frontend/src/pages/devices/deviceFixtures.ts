import IDeviceData from '../../types/device';
import {INodeData} from '../../types/node';
import {DeviceStateType} from '../../types/deviceStateType';
import {DeviceType} from '../../types/device.type.enum';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {PlatformType} from '../../types/platform.type.enum';

export const device = (change: Partial<IDeviceData> = {}): IDeviceData => ({
    id: 1, nodeId: 0, companyId: 1, deviceIdentifier: 'serial-1', deviceType: DeviceType.Phone,
    platformType: PlatformType.iOS, name: 'Pixel', alias: '', os: 'Android', osVersion: '14',
    status: DeviceStateType.Booted, isLocked: false, manager: '', isAcknowledged: true,
    connectionType: DeviceConnectionType.USB,
    connectionParameter: {connectionType: DeviceConnectionType.USB, ip: '', port: 0},
    customParameter: [], deviceParameter: [],
    ...change,
});

export const node = (change: Partial<INodeData> = {}): INodeData =>
    ({id: 5, identifier: 'node-a-id', name: 'node-a', status: 1, ...change});
