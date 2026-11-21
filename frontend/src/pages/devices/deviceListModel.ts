// How the device pool is filtered by platform and grouped by node.
import _ from 'lodash';
import IDeviceData from '../../types/device';
import {INodeData} from '../../types/node';
import {PlatformType} from '../../types/platform.type.enum';

export type PlatformChoice = PlatformType | 'all';

export type PlatformCounts = Record<number, number>;

export interface DeviceGroup {
    title: string,
    devices: IDeviceData[],
}

export const platformCounts = (devices: IDeviceData[]): PlatformCounts => _.countBy(devices, d => d.platformType);

// the platforms that have devices, in enum order
export const platformsOf = (counts: PlatformCounts): PlatformType[] =>
    Object.keys(counts).map(Number).sort((a, b) => a - b) as PlatformType[];

export const filterByPlatform = (devices: IDeviceData[], platform: PlatformChoice): IDeviceData[] =>
    devices.filter(d => platform === 'all' || d.platformType === platform);

const CONNECTED = 1;

const groupLabel = (device: IDeviceData, nodes: INodeData[] | null): string => {
    if (!device.node) return 'Master';
    const connected = nodes?.find(n => n.id === device.nodeId)?.status === CONNECTED;
    return `${device.node.name} (${connected ? 'Connected' : 'Disconnected'})`;
};

// groups by node, connected nodes first, each by name; devices without a
// node belong to the master
export const groupDevices = (devices: IDeviceData[], nodes: INodeData[] | null): DeviceGroup[] => {
    const groups = _.groupBy(devices, d => groupLabel(d, nodes));
    return _.orderBy(Object.entries(groups), [
        ([label]) => label.endsWith('(Connected)') ? 0 : 1,
        ([label]) => label.replace(/ \(.*\)$/, ''),
    ], ['asc', 'asc']).map(([label, items]) => ({title: `Node: ${label}`, devices: items}));
};
