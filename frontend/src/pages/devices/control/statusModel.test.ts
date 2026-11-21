import {describe, expect, it} from 'vitest';
import {DeviceLease} from '../../../types/device.control.lease';
import {controlStatus} from './statusModel';

const now = Date.parse('2026-01-01T10:00:00Z');
const lease = (over = {}): DeviceLease => ({
    deviceId: '1', owner: 'anna', acquiredAt: '2026-01-01T09:00:00Z', lastUsed: '', expiresAt: '2026-01-01T10:10:00Z', ...over,
});

describe('controlStatus', () => {
    it('says the user controls the device, and for how long', () => {
        expect(controlStatus(true, {locked: true, lockedByTest: false, lease: lease()}, now)).toEqual({
            color: 'success', label: 'You control this device', own: true, taken: false,
            note: 'ends in 10 minutes unless you use the device',
        });
    });

    it('names somebody else and since when', () => {
        const status = controlStatus(false, {locked: true, lockedByTest: false, lease: lease({owner: 'token:ci'})}, now);
        expect(status).toMatchObject({color: 'warning', label: 'Controlled by token ci', taken: true, own: false});
        expect(status.note).toMatch(/^since \d\d:\d\d$/);
    });

    it('has no note when the start is unknown', () => {
        expect(controlStatus(false, {locked: true, lockedByTest: false, lease: lease({acquiredAt: ''})}, now).note).toBeNull();
    });

    it('treats a lease the hub dropped as not ours', () => {
        expect(controlStatus(true, {locked: true, lockedByTest: false, lease: lease()}, now).own).toBe(true);
        expect(controlStatus(true, {locked: false, lockedByTest: false}, now).label).toBe('Free');
    });

    it('says a test run uses the device', () => {
        expect(controlStatus(false, {locked: true, lockedByTest: true}, now))
            .toMatchObject({label: 'A test run is using this device', taken: true});
    });

    it('is free without a lock status', () => {
        expect(controlStatus(false, null, now)).toEqual({
            color: 'default', label: 'Free', note: 'Take control to use the device.', own: false, taken: false,
        });
    });
});
