import {describe, expect, it} from 'vitest';
import {fireEvent, screen} from '@testing-library/react';
import {DeviceConnectionType} from '../../types/device.connection.type.enum';
import {DeviceStateType} from '../../types/deviceStateType';
import {device} from './deviceFixtures';
import {fakeApi} from '../../testing/fakeApi';
import {open, withHistory} from './deviceShowKit';

describe('DeviceShowPage', () => {
    it('shows the device with its title, state and setup', async () => {
        const stored = device({
            alias: 'Test phone', nodeId: 5, createdAt: new Date('2024-05-06'), deviceParameter: [{key: 'abi', value: 'arm64'}],
            customParameter: [{key: 'PIN', value: '1234'}],
            connectionParameter: {connectionType: DeviceConnectionType.Remote, ip: '10.0.0.2', port: 5555},
            connection: {version: '2.1'} as never,
        });
        open(stored, withHistory());
        expect(await screen.findByRole('heading', {level: 1, name: 'Test phone'})).toBeInTheDocument();
        expect(screen.getByText('Ready')).toBeInTheDocument();
        expect(screen.getByText(/Phone · Android 14 · on node-a/)).toBeInTheDocument();
        expect(screen.getByText('10.0.0.2:5555')).toBeInTheDocument();
        expect(screen.getByText('version 2.1')).toBeInTheDocument();
        expect(screen.getByText('arm64')).toBeInTheDocument();
        expect(screen.getByText('1234')).toBeInTheDocument();
        expect(screen.getByText('(connected)')).toBeInTheDocument();
        expect(screen.getByText('2024/05/06')).toBeInTheDocument();
    });

    it('shows the live status of the device list instead of the loaded one', async () => {
        const stored = device({status: DeviceStateType.Shutdown});
        open(stored, withHistory(), [{...stored, status: DeviceStateType.Booted, isLocked: true, lockedBy: 'anna'}]);
        expect(await screen.findByText('Controlled by anna')).toBeInTheDocument();
        expect(screen.queryByText('Shut down')).toBeNull();
        expect(screen.getByRole('link', {name: 'Live control'})).toHaveAttribute('href', '/project/p1/device/1/control');
    });

    it('explains a device that cannot be used and offers no live control', async () => {
        open(device({status: DeviceStateType.RemoteDisconnected, deviceType: 99 as never, createdAt: undefined}), withHistory());
        expect(await screen.findByText("The node doesn't see the device.")).toBeInTheDocument();
        expect(screen.getByText('No screen while the device is disconnected.')).toBeInTheDocument();
        expect(screen.getByRole('link', {name: 'Live control'})).toHaveAttribute('aria-disabled', 'true');
    });

    it('shows the master as node of a device without one', async () => {
        open(device({nodeId: 0, connectionParameter: undefined as never, connection: undefined}), withHistory());
        expect(await screen.findByText('the master itself')).toBeInTheDocument();
    });
});

describe('DeviceShowPage history', () => {
    it('lists the recent tests and opens a test run', async () => {
        open(device(), withHistory());
        expect(await screen.findByText('Nightly · run 4')).toBeInTheDocument();
        fireEvent.click(screen.getByText('Login'));
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/app:2/test/3/run/4/9');
    });

    it('says when no test ran or the history is unavailable', async () => {
        const none = open(device(), fakeApi({'GET /p1/device/1/protocols': []}));
        expect(await screen.findByText('No test of this project ran on this device yet.')).toBeInTheDocument();
        none.unmount();
        open(device(), fakeApi({}));
        expect(await screen.findByText("The test history couldn't be loaded.")).toBeInTheDocument();
    });
});
