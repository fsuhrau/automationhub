import {describe, expect, it} from 'vitest';
import {device, protocol} from './harness/fixtures';
import {deviceCount, liveDevices} from './liveDevicesModel';

const status = (deviceId: number, withDevice = true): never =>
    ({id: deviceId, testRunId: 3, deviceId, device: withDevice ? device({id: deviceId, name: `from status ${deviceId}`}) : undefined as never,
        startupTime: 1000, histAvgStartupTime: 0} as never);

describe('liveDevices', () => {
    it('lists the held devices, the hub state winning over older data', () => {
        const found = liveDevices({
            held: [1, 2, 3], deviceStatus: [status(1), status(2, false)], hubDevices: [device({id: 1, name: 'from hub'})],
            protocols: [protocol({deviceId: 3, device: device({id: 3, name: 'from protocol'})}), protocol({id: 2, deviceId: 4, device: null})],
        });
        expect(found.map(d => d.name)).toEqual(['from hub', 'from protocol']);
    });

    it('guesses the devices that started the app or ran a test when the held ones are unknown', () => {
        const found = liveDevices({
            held: undefined, deviceStatus: [status(1)], hubDevices: null,
            protocols: [protocol({deviceId: 2, device: device({id: 2})}), protocol({id: 2, deviceId: 2, device: null}), protocol({id: 3, deviceId: null})],
        });
        expect(found.map(d => d.id)).toEqual([1, 2]);
    });

    it('is empty when no device is held', () => {
        expect(liveDevices({held: [], deviceStatus: [status(1)], hubDevices: undefined, protocols: []})).toEqual([]);
    });
});

describe('deviceCount', () => {
    it('counts the devices of the protocols', () => {
        expect(deviceCount([protocol({deviceId: 1}), protocol({id: 2, deviceId: 1}), protocol({id: 3, deviceId: 2})], [])).toBe(2);
    });

    it('falls back to the devices that started the app', () => {
        expect(deviceCount([protocol({deviceId: null})], [status(1), status(2)])).toBe(2);
    });
});
