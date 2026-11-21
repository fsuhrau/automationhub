import {describe, expect, it} from 'vitest';
import {act, fireEvent, screen} from '@testing-library/react';
import {heldBy} from './control/fakes';
import {BASE, loaded, renderPage, takeControl} from './control/controlHarness';

describe('the control page: the device', () => {
    it('shows the device, where it runs and a link to its details', async () => {
        renderPage();
        await loaded();
        expect(screen.getByText(/Android 14/)).toHaveTextContent('· Pixel 8 · on node-1');
        expect(screen.getByRole('link', {name: 'Device details'})).toHaveAttribute('href', '/project/p1/device/7');
        expect(await screen.findByText('Take control to use the device.')).toBeInTheDocument();
    });

    it('shows a failure to load the device', async () => {
        renderPage({answers: {[`GET ${BASE}`]: new Error('no such device')}});
        expect(await screen.findByRole('alert')).toHaveTextContent('no such device');
        expect(screen.getByRole('heading', {level: 1})).toHaveTextContent('Device 7');
    });
});

describe('the control page: the lease', () => {
    it('takes control and gives the device back', async () => {
        const page = renderPage();
        await loaded();
        await takeControl(page);
        expect(page.api.called(`POST ${BASE}/lock`)).toHaveLength(1);
        expect(screen.getByRole('button', {name: 'Back'})).toBeEnabled();
        fireEvent.click(screen.getByRole('button', {name: /Give the device back/}));
        await screen.findByText('Take control to use the device.');
        expect(page.api.called(`DELETE ${BASE}/lock`)[0].config?.headers).toEqual({'X-Device-Lease': 'tok-1'});
        expect(screen.getByRole('button', {name: 'Back'})).toBeDisabled();
    });

    it('shows why taking control failed', async () => {
        renderPage({answers: {[`POST ${BASE}/lock`]: new Error('device is busy')}});
        await loaded();
        fireEvent.click(screen.getAllByRole('button', {name: 'Take control'})[0]);
        expect(await screen.findByRole('alert')).toHaveTextContent('device is busy');
    });

    it('shows who else controls the device and offers no control', async () => {
        renderPage({lock: heldBy({owner: 'token:ci'})});
        expect(await screen.findAllByText('Controlled by token ci')).toHaveLength(2);
        expect(screen.getByText(/Control is possible once the device is free/)).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Take control'})).toBeDisabled();
    });

    it('says a test run is using the device', async () => {
        renderPage({lock: {locked: true, lockedByTest: true}});
        expect(await screen.findAllByText('A test run is using this device')).toHaveLength(2);
    });

    it('keeps the lease of the tab and asks the hub again after an input failed', async () => {
        const page = renderPage({answers: {[`POST ${BASE}/key`]: new Error('lease expired')}});
        await loaded();
        await takeControl(page);
        fireEvent.click(screen.getByText('Screenshots'));
        const before = page.api.called(`GET ${BASE}/lock`).length;
        fireEvent.click(screen.getByRole('button', {name: 'Home'}));
        expect(await screen.findByText('lease expired')).toBeInTheDocument();
        await act(async () => undefined);
        expect(page.api.called(`GET ${BASE}/lock`).length).toBeGreaterThan(before);
    });
});
