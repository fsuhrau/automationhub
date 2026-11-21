import {describe, expect, it} from 'vitest';
import {describeController} from '../../components/DeviceController';
import {
    alertSeverity, availability, connectionAddress, DeviceLiveState, mergeLiveDevice, recentTestLink, resolveNode, showsLog,
} from './deviceDetailsModel';
import {DeviceStateType} from '../../types/deviceStateType';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {PlatformType} from '../../types/platform.type.enum';
import ITestProtocolData from '../../types/test.protocol';
import {device, node} from './deviceFixtures';

const describeOwner = describeController;

describe('availability', () => {
    it.each([
        {status: DeviceStateType.Booted, label: 'Ready', tone: 'success', control: true, unlock: false},
        {status: DeviceStateType.Unauthorized, label: 'Not authorized', tone: 'warning', control: false, unlock: false},
        {status: DeviceStateType.RemoteDisconnected, label: 'Disconnected', tone: 'error', control: false, unlock: false},
        {status: DeviceStateType.NodeDisconnected, label: 'Node offline', tone: 'error', control: false, unlock: false},
        {status: DeviceStateType.Shutdown, label: 'Shut down', tone: 'default', control: false, unlock: false},
        {status: DeviceStateType.Unknown, label: 'Unknown', tone: 'default', control: false, unlock: false},
    ])('answers $label for an unheld device', ({status, label, tone, control, unlock}) => {
        expect(availability({status, isLocked: false}, describeOwner)).toMatchObject({
            label, tone, canControl: control, canUnlock: unlock,
        });
    });

    it('says a test run holds a locked device', () => {
        expect(availability({status: DeviceStateType.Booted, isLocked: true}, describeOwner))
            .toMatchObject({label: 'In use by a test run', canControl: false, canUnlock: true});
        expect(availability({status: DeviceStateType.Locked, isLocked: false}, describeOwner).label).toBe('In use by a test run');
    });

    it('names who controls the device', () => {
        const held = {status: DeviceStateType.Booted, isLocked: true, lockedBy: 'token:ci'};
        expect(availability(held, describeOwner, 3)).toMatchObject({label: 'Controlled by token ci', tone: 'warning', canControl: true});
        expect(availability({...held, status: DeviceStateType.Shutdown}, describeOwner).canControl).toBe(false);
    });

    const heldBy = (lockedBy: string): DeviceLiveState => ({status: DeviceStateType.Locked, isLocked: true, lockedBy});

    it('says the signed-in user controls it, by user id', () => {
        expect(availability(heldBy('anna (user:3)'), describeOwner, 3))
            .toMatchObject({label: 'Controlled by you', tone: 'success', canUnlock: true});
        expect(availability(heldBy('user:3'), describeOwner, 3).label).toBe('Controlled by you');
    });

    it('does not take the same name with another id for you', () => {
        expect(availability(heldBy('anna (user:4)'), describeOwner, 3))
            .toMatchObject({label: 'Controlled by anna', tone: 'warning'});
    });

    it.each(['token:3', 'anna', 'anonymous', 'mcp:anna (user:3):sess1'])('never takes %s for you', owner => {
        expect(availability(heldBy(owner), describeOwner, 3).tone).toBe('warning');
    });

    it('knows nobody when no user is signed in', () => {
        expect(availability(heldBy('anna (user:3)'), describeOwner).tone).toBe('warning');
    });
});

describe('mergeLiveDevice', () => {
    it('takes the live status and lease over the stored ones', () => {
        const stored = device({status: DeviceStateType.Shutdown, alias: 'kept'});
        const live = device({status: DeviceStateType.Booted, isLocked: true, lockedBy: 'anna', alias: 'ignored'});
        expect(mergeLiveDevice(stored, live)).toEqual({...stored, status: DeviceStateType.Booted, isLocked: true, lockedBy: 'anna'});
    });

    it('keeps the stored device without a live one', () => {
        const stored = device();
        expect(mergeLiveDevice(stored)).toBe(stored);
    });

    it('does not change its input', () => {
        const stored = device({status: DeviceStateType.Shutdown});
        mergeLiveDevice(stored, device({status: DeviceStateType.Booted}));
        expect(stored.status).toBe(DeviceStateType.Shutdown);
    });
});

describe('resolveNode', () => {
    it('prefers the node of the hub state, else the loaded one', () => {
        const loaded = node({status: 0});
        const d = device({nodeId: 5, node: loaded});
        expect(resolveNode(d, [node({status: 1})])?.status).toBe(1);
        expect(resolveNode(d, null)).toBe(loaded);
        expect(resolveNode(d, [node({id: 9})])).toBe(loaded);
        expect(resolveNode(device(), [])).toBeUndefined();
    });
});

describe('showsLog', () => {
    it.each([
        [PlatformType.Android, DeviceStateType.Booted, true],
        [PlatformType.Android, DeviceStateType.Locked, true],
        [PlatformType.Android, DeviceStateType.Shutdown, false],
        [PlatformType.iOS, DeviceStateType.Booted, false],
    ])('platform %s in state %s: %s', (platformType, status, shown) => {
        expect(showsLog(device({platformType, status}))).toBe(shown);
    });
});

describe('connectionAddress', () => {
    it.each([
        [{connectionType: DeviceConnectionType.Remote, ip: '1.2.3.4', port: 80}, '1.2.3.4:80'],
        [{connectionType: DeviceConnectionType.Remote, ip: '1.2.3.4', port: 0}, '1.2.3.4'],
        [{connectionType: DeviceConnectionType.Remote, ip: '', port: 80}, null],
        [{connectionType: DeviceConnectionType.USB, ip: '1.2.3.4', port: 80}, null],
        [undefined, null],
    ])('%j gives %s', (connection, expected) => {
        expect(connectionAddress(connection)).toBe(expected);
    });
});

describe('alertSeverity', () => {
    it.each([['error', 'error'], ['warning', 'warning'], ['default', 'info'], ['success', 'info']] as const)(
        'tone %s is %s', (tone, severity) => expect(alertSeverity(tone)).toBe(severity));
});

describe('recentTestLink', () => {
    it('links to the protocol of the run, or to nothing without a test', () => {
        const withTest = {id: 9, testRunId: 4, testRun: {testId: 3, test: {appId: 2}}} as unknown as ITestProtocolData;
        const without = {id: 9, testRunId: 4, testRun: {testId: 3}} as unknown as ITestProtocolData;
        expect(recentTestLink('p1', withTest)).toBe('/project/p1/app:2/test/3/run/4/9');
        expect(recentTestLink('p1', without)).toBeNull();
        expect(recentTestLink('p1', {id: 1} as unknown as ITestProtocolData)).toBeNull();
    });
});
