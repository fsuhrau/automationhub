import {useCallback, useMemo} from 'react';
import {useSearchParams} from 'react-router-dom';
import ITestRunData from '../../../types/test.run';
import {useHubState} from '../../../hooks/HubStateProvider';
import {buildTimeline, replayRange, ReplayEvent, ReplayRange, ReplaySegment, segmentsOf} from '../replayModel';
import {deviceIdsOf, deviceNameOf, initialTime, pickDeviceId} from './replayDeviceModel';

export interface ReplayDevice {
    deviceIds: number[],
    deviceId: number | null,
    deviceSegments: ReplaySegment[],
    events: ReplayEvent[],
    range: ReplayRange | null,
    initial: number,
    nameOf: (id: number) => string,
    // switch the device, keeping the time
    pick: (id: number, time: number) => void,
}

// useReplayDevice picks the device of a run's replay from the URL ("device",
// "t") and builds what the replay shows of it.
export const useReplayDevice = (run: ITestRunData): ReplayDevice => {
    const {state: hub} = useHubState();
    const [params, setParams] = useSearchParams();
    const segments = useMemo(() => segmentsOf(run.recordings), [run.recordings]);
    const deviceIds = useMemo(() => deviceIdsOf(run, segments), [run, segments]);
    const deviceId = pickDeviceId(deviceIds, Number(params.get('device')));
    const deviceSegments = useMemo(() => (deviceId !== null ? segments.get(deviceId) : undefined) ?? [], [segments, deviceId]);
    const events = useMemo(() => buildTimeline(run, deviceId), [run, deviceId]);
    const range = useMemo(() => replayRange(deviceSegments, events), [deviceSegments, events]);
    const nameOf = useCallback((id: number) => deviceNameOf(hub.devices, run, id), [hub.devices, run]);
    const pick = useCallback((id: number, time: number) => setParams(p => {
        const next = new URLSearchParams(p);
        next.set('device', String(id));
        next.set('t', String(Math.round(time)));
        return next;
    }, {replace: true}), [setParams]);
    const initial = initialTime(Number(params.get('t')), range, deviceSegments);
    return {deviceIds, deviceId, deviceSegments, events, range, initial, nameOf, pick};
};
