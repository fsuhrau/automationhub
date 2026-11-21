// The devices whose screens the run page can show while the run executes.
import IDeviceData from '../../../types/device';
import ITestProtocolData from '../../../types/test.protocol';
import ITesRunDeviceStatusData from '../../../types/test.run.device.status';

export interface LiveDevicesInput {
    // undefined on older masters, which don't say which devices a run holds
    held: number[] | undefined,
    protocols: ITestProtocolData[],
    deviceStatus: ITesRunDeviceStatusData[],
    hubDevices: IDeviceData[] | null | undefined,
}

// knownDevices has every device the run page heard of, the hub's state last
const knownDevices = ({deviceStatus, protocols, hubDevices}: LiveDevicesInput): Map<number, IDeviceData> => {
    const known = new Map<number, IDeviceData>();
    deviceStatus.forEach(d => d.device && known.set(d.deviceId, d.device));
    protocols.forEach(p => p.device && p.deviceId && known.set(p.deviceId, p.device));
    hubDevices?.forEach(d => known.set(d.id, d));
    return known;
};

// without the held devices, the ones that started the app or ran a test
const guessedIds = ({deviceStatus, protocols}: LiveDevicesInput): number[] => [...new Set([
    ...deviceStatus.map(d => d.deviceId),
    ...protocols.map(p => p.deviceId as number).filter(id => id),
])];

export const liveDevices = (input: LiveDevicesInput): IDeviceData[] => {
    const known = knownDevices(input);
    const ids = input.held === undefined ? guessedIds(input) : input.held;
    return ids.map(id => known.get(id)).filter((d): d is IDeviceData => !!d);
};

// how many devices the run used
export const deviceCount = (protocols: ITestProtocolData[], deviceStatus: ITesRunDeviceStatusData[]): number =>
    new Set(protocols.map(p => p.deviceId).filter(id => id)).size || deviceStatus.length;
