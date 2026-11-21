// Which device a replay shows and where it starts. No React.
import ITestRunData from '../../../types/test.run';
import IDeviceData from '../../../types/device';
import {ReplayRange, ReplaySegment} from '../replayModel';

// deviceIdsOf lists the devices with a recording, then those that only ran tests
export const deviceIdsOf = (run: ITestRunData, segments: Map<number, ReplaySegment[]>): number[] => {
    const ids = [...segments.keys()];
    (run.protocols ?? []).forEach(p => p.deviceId && !ids.includes(p.deviceId) && ids.push(p.deviceId));
    return ids;
};

// pickDeviceId is the requested device when the replay has it, else the first
export const pickDeviceId = (ids: number[], requested: number): number | null =>
    ids.includes(requested) ? requested : ids[0] ?? null;

const clamp = (time: number, range: ReplayRange): number => Math.max(range.start, Math.min(range.end, time));

// initialTime is the requested time within the replay, else the start of the
// device's first recording, else the start of the replay
export const initialTime = (requested: number, range: ReplayRange | null, segments: ReplaySegment[]): number => {
    const start = range?.start ?? 0;
    if (range && requested > 0) return clamp(requested, range);
    return segments[0]?.start ?? start;
};

export const deviceNameOf = (hubDevices: IDeviceData[] | null | undefined, run: ITestRunData, id: number): string => {
    const device = hubDevices?.find(x => x.id === id) ?? run.protocols?.find(p => p.deviceId === id)?.device;
    return device ? (device.alias || device.name) : `Device ${id}`;
};
