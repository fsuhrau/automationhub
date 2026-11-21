import {expect, it} from 'vitest';
import IDeviceData from '../../../types/device';
import {segmentsOf} from '../replayModel';
import {deviceIdsOf, deviceNameOf, initialTime, pickDeviceId} from './replayDeviceModel';
import {protocol, recording, run, T0} from './testFixtures';

it('deviceIdsOf: lists recorded devices first, then those that only ran tests, once each', () => {
    const r = run({protocols: [protocol({deviceId: 5}), protocol({id: 12, deviceId: 6}), protocol({id: 13, deviceId: null})]});
    expect(deviceIdsOf(r, segmentsOf([recording({deviceId: 6}), recording({id: 2, deviceId: 4})]))).toEqual([6, 4, 5]);
});

it('deviceIdsOf: handles a run without protocols', () => {
    expect(deviceIdsOf(run({protocols: undefined}), new Map())).toEqual([]);
});

it.each([[[5, 6], 6, 6], [[5, 6], 9, 5], [[], 0, null]])('pickDeviceId: %j requested %d gives %s', (ids, requested, expected) => {
    expect(pickDeviceId(ids, requested)).toBe(expected);
});

const range = {start: T0, end: T0 + 10000};
const segments = segmentsOf([recording({startedAt: new Date(T0 + 2000).toISOString()})]).get(5) ?? [];

it('initialTime: clamps a requested time into the replay', () => {
    expect(initialTime(T0 + 4000, range, segments)).toBe(T0 + 4000);
    expect(initialTime(T0 + 99999, range, segments)).toBe(range.end);
    expect(initialTime(1, range, segments)).toBe(range.start);
});

it('initialTime: starts at the first recording, else the start of the replay', () => {
    expect(initialTime(0, range, segments)).toBe(T0 + 2000);
    expect(initialTime(0, range, [])).toBe(T0);
    expect(initialTime(5, null, [])).toBe(0);
});

const hub = [{id: 5, name: 'pixel', alias: 'Lab phone'}, {id: 6, name: 'iphone', alias: ''}] as IDeviceData[];

it('deviceNameOf: prefers the alias, then the name', () => {
    expect(deviceNameOf(hub, run(), 5)).toBe('Lab phone');
    expect(deviceNameOf(hub, run(), 6)).toBe('iphone');
});

it('deviceNameOf: falls back to the device of a protocol, then to the id', () => {
    const withDevice = run({protocols: [protocol({deviceId: 8, device: {id: 8, name: 'tab', alias: ''} as IDeviceData})]});
    expect(deviceNameOf(null, withDevice, 8)).toBe('tab');
    expect(deviceNameOf(undefined, run({protocols: undefined}), 9)).toBe('Device 9');
});
