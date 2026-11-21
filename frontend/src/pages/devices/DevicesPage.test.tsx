import {describe, expect, it} from 'vitest';
import {fireEvent, screen, waitFor, within} from '@testing-library/react';
import DevicesPage from './DevicesPage';
import {DeviceStateType} from '../../types/deviceStateType';
import {PlatformType} from '../../types/platform.type.enum';
import {device, node} from './deviceFixtures';
import {fakeApi, FakeApi} from '../../testing/fakeApi';
import {renderPage} from './deviceHarness';

const open = (api: FakeApi, nodes = [node()]): ReturnType<typeof renderPage> =>
    renderPage(<DevicesPage/>, {api, nodes, route: 'devices', url: '/project/p1/devices'});

const master = device({id: 1, name: 'Mac mini', alias: 'Build mac', status: DeviceStateType.Booted});
const remote = device({id: 2, name: 'Galaxy', platformType: PlatformType.Android, nodeId: 5, node: node(), status: DeviceStateType.Shutdown});

describe('DevicesPage', () => {
    it('shows the devices grouped by node with their status', async () => {
        open(fakeApi({'GET /p1/devices': [master, remote]}));
        expect(await screen.findByText('Build mac')).toBeInTheDocument();
        expect(screen.getByText('Node: Master')).toBeInTheDocument();
        expect(screen.getByText('Node: node-a (Connected)')).toBeInTheDocument();
        expect(screen.getByText('Booted')).toBeInTheDocument();
        expect(screen.getByText('Shutdown')).toBeInTheDocument();
        expect(screen.getByText('Galaxy')).toBeInTheDocument();
    });

    it('marks the nodes that are offline and puts them last', async () => {
        open(fakeApi({'GET /p1/devices': [master, remote]}), [node({status: 0})]);
        await screen.findByText('Build mac');
        const titles = screen.getAllByRole('heading', {level: 2}).map(h => h.textContent);
        expect(titles).toEqual(['Device Pool', 'Node: Master', 'Node: node-a (Disconnected)']);
    });

    it('shows who controls a device', async () => {
        open(fakeApi({'GET /p1/devices': [device({lockedBy: 'token:ci'})]}));
        expect(await screen.findByLabelText('Controlled by token ci')).toBeInTheDocument();
    });

});

describe('DevicesPage filters and rows', () => {
    it('filters by platform once there are several', async () => {
        open(fakeApi({'GET /p1/devices': [master, remote]}));
        await screen.findByText('Build mac');
        fireEvent.click(screen.getByRole('button', {name: /Android \(1\)/}));
        expect(screen.queryByText('Build mac')).toBeNull();
        expect(screen.getByText('Galaxy')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', {name: 'All (2)'}));
        expect(screen.getByText('Build mac')).toBeInTheDocument();
    });

    it('offers no platform filter for a single platform', async () => {
        open(fakeApi({'GET /p1/devices': [master]}));
        await screen.findByText('Build mac');
        expect(screen.queryByRole('group', {name: 'Filter by platform'})).toBeNull();
    });

    it('says when there are no devices', async () => {
        open(fakeApi({'GET /p1/devices': []}));
        expect(await screen.findByText(/No devices yet/)).toBeInTheDocument();
    });

    it('shows the load error and loads again on retry', async () => {
        let fail = true;
        const api = fakeApi({'GET /p1/devices': () => fail ? new Error('master is down') : [master]});
        open(api);
        expect(await screen.findByText('master is down')).toBeInTheDocument();
        fail = false;
        fireEvent.click(screen.getByRole('button', {name: 'Retry'}));
        expect(await screen.findByText('Build mac')).toBeInTheDocument();
        expect(screen.queryByText('master is down')).toBeNull();
    });

});

describe('DevicesPage navigation', () => {
    it('opens the device on a click on its row', async () => {
        open(fakeApi({'GET /p1/devices': [master]}));
        fireEvent.click(await screen.findByText('Build mac'));
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/device/1');
    });

    it('opens the details from the details button without the row click', async () => {
        open(fakeApi({'GET /p1/devices': [master]}));
        fireEvent.click(await screen.findByRole('button', {name: 'Device details'}));
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/device/1');
    });

    it('opens live control of a device that has a real device', async () => {
        const real = {companyId: '1', deviceIp: '', connection: 0};
        open(fakeApi({'GET /p1/devices': [device({dev: real})]}));
        fireEvent.click(await screen.findByRole('button', {name: 'Live control'}));
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/device/1/control');
    });

});

describe('DevicesPage actions', () => {
    it('unlocks a locked device and shows it free', async () => {
        const api = fakeApi({
            'GET /p1/devices': [device({isLocked: true, lockedBy: 'anna'})],
            'POST /p1/device/1/unlock': {},
        });
        open(api);
        fireEvent.click(await screen.findByRole('button', {name: 'Unlock device'}));
        await waitFor(() => expect(screen.queryByRole('button', {name: 'Unlock device'})).toBeNull());
        expect(api.calls.some(c => c.method === 'POST' && c.url === '/p1/device/1/unlock')).toBe(true);
    });

    it('shows why unlocking failed', async () => {
        const api = fakeApi({
            'GET /p1/devices': [device({isLocked: true})],
            'POST /p1/device/1/unlock': new Error('not allowed'),
        });
        open(api);
        fireEvent.click(await screen.findByRole('button', {name: 'Unlock device'}));
        expect(await screen.findByText('not allowed')).toBeInTheDocument();
    });

});

describe('DevicesPage running tests', () => {
    it('starts a test on a device', async () => {
        const api = fakeApi({
            'GET /p1/devices': [device({connection: {id: 1, version: '1'} as never})],
            'POST /p1/device/1/tests': {},
        });
        open(api);
        fireEvent.click(await screen.findByRole('button', {name: 'Run'}));
        fireEvent.change(screen.getByLabelText('Test name'), {target: {value: 'Suite.Smoke'}});
        fireEvent.change(screen.getByLabelText('Parameters'), {target: {value: 'user=a'}});
        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', {name: 'Start'}));
        await waitFor(() => expect(api.calls.at(-1)).toMatchObject({
            method: 'POST', url: '/p1/device/1/tests', data: {testName: 'Suite.Smoke', env: 'user=a'},
        }));
        expect(screen.queryByRole('dialog')).toBeNull();
    });

    it('shows why starting a test failed and closes the dialog on cancel', async () => {
        const api = fakeApi({
            'GET /p1/devices': [device({connection: {id: 1, version: '1'} as never})],
            'POST /p1/device/1/tests': new Error('no such test'),
        });
        open(api);
        fireEvent.click(await screen.findByRole('button', {name: 'Run'}));
        fireEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
        fireEvent.click(screen.getByRole('button', {name: 'Run'}));
        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', {name: 'Start'}));
        expect(await screen.findByText('no such test')).toBeInTheDocument();
    });
});
