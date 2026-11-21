import {describe, expect, it} from 'vitest';
import {fireEvent, screen, waitFor, within} from '@testing-library/react';
import {device} from './deviceFixtures';
import {open, withHistory} from './deviceShowKit';

describe('DeviceShowPage unlocking', () => {
    it('unlocks the device after a confirmation', async () => {
        const api = withHistory({'POST /p1/device/1/unlock': {}});
        const stored = device({isLocked: true, lockedBy: 'anna'});
        open(stored, api, [stored]);
        await screen.findByText('Controlled by anna');
        fireEvent.click(screen.getByRole('button', {name: 'Unlock'}));
        expect(screen.getByText('This ends the live control of anna.')).toBeInTheDocument();
        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', {name: 'Unlock'}));
        await waitFor(() => expect(api.calls.some(c => c.url === '/p1/device/1/unlock')).toBe(true));
        await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
        expect(await screen.findByText('Ready')).toBeInTheDocument();
    });

    it('warns that unlocking a test run fails it, and unlocks through the menu', async () => {
        const api = withHistory({'POST /p1/device/1/unlock': {}});
        open(device({isLocked: true}), api);
        await screen.findByText('In use by a test run');
        fireEvent.click(screen.getByRole('button', {name: 'More actions'}));
        fireEvent.click(screen.getByRole('menuitem', {name: /Unlock/}));
        expect(screen.getByText(/loses it and fails/)).toBeInTheDocument();
        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', {name: 'Cancel'}));
        await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    });

    it('closes the menu with Escape', async () => {
        open(device(), withHistory());
        await screen.findByText('Ready');
        fireEvent.click(screen.getByRole('button', {name: 'More actions'}));
        fireEvent.keyDown(screen.getByRole('menu'), {key: 'Escape'});
        await waitFor(() => expect(screen.queryByRole('menu')).toBeNull());
    });

    it('shows why unlocking failed', async () => {
        open(device({isLocked: true}), withHistory({'POST /p1/device/1/unlock': new Error('still held')}));
        fireEvent.click(await screen.findByRole('button', {name: 'Unlock'}));
        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', {name: 'Unlock'}));
        expect(await screen.findByText('still held')).toBeInTheDocument();
        await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    });
});

describe('DeviceShowPage deleting', () => {
    it('deletes the device and goes back to the device list', async () => {
        const api = withHistory({'DELETE /p1/device/1': undefined});
        open(device({alias: 'Old phone'}), api);
        await screen.findByText('Ready');
        fireEvent.click(screen.getByRole('button', {name: 'More actions'}));
        fireEvent.click(screen.getByRole('menuitem', {name: 'Delete device'}));
        expect(screen.getByText('Delete Old phone?')).toBeInTheDocument();
        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', {name: 'Delete device'}));
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/devices');
    });

    it('keeps the device when deleting fails', async () => {
        open(device(), withHistory({'DELETE /p1/device/1': new Error('cannot delete')}));
        await screen.findByText('Ready');
        fireEvent.click(screen.getByRole('button', {name: 'More actions'}));
        fireEvent.click(screen.getByRole('menuitem', {name: 'Delete device'}));
        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', {name: 'Delete device'}));
        expect(await screen.findByText('cannot delete')).toBeInTheDocument();
        await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
        expect(screen.getByRole('heading', {level: 1, name: 'Pixel'})).toBeInTheDocument();
    });

    it('links to the edit page', async () => {
        open(device(), withHistory());
        const links = await screen.findAllByRole('link', {name: 'Edit'});
        links.forEach(link => expect(link).toHaveAttribute('href', '/project/p1/device/1/edit'));
    });
});
