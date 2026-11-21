import {describe, expect, it, vi} from 'vitest';
import {fireEvent, screen} from '@testing-library/react';
import DeviceShowPage from './DeviceShowPage';
import {DeviceStateType} from '../../types/deviceStateType';
import {device} from './deviceFixtures';
import {fakeApi} from '../../testing/fakeApi';
import {renderPage} from './deviceHarness';

vi.mock('../../hooks/AuthProvider', () => ({useAuth: () => ({user: {id: 3, name: 'anna'}})}));

describe('DeviceShowPage for the user who controls the device', () => {
    it('says so and offers live control in the alert', async () => {
        const held = device({status: DeviceStateType.Booted, isLocked: true, lockedBy: 'anna (user:3)'});
        renderPage(<DeviceShowPage device={held}/>, {
            api: fakeApi({'GET /p1/device/1/protocols': []}), devices: [held], route: 'device/:deviceId', url: '/project/p1/device/1',
        });
        expect(await screen.findAllByText('Controlled by you')).not.toHaveLength(0);
        const links = screen.getAllByRole('link', {name: 'Live control'});
        expect(links).toHaveLength(2);
        fireEvent.click(links[1]);
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/device/1/control');
    });
});
