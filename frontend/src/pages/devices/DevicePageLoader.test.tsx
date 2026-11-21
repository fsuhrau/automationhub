import {describe, expect, it} from 'vitest';
import {fireEvent, screen} from '@testing-library/react';
import DevicePageLoader from './DevicePageLoader';
import {device, fakeApi, renderPage} from './testKit';

const open = (api = fakeApi({}), edit = false): ReturnType<typeof renderPage> =>
    renderPage(<DevicePageLoader edit={edit}/>, {api, route: 'device/:deviceId', url: '/project/p1/device/1'});

const loaded = (): ReturnType<typeof fakeApi> =>
    fakeApi({'GET /p1/device/1': device({alias: 'Loaded'}), 'GET /p1/device/1/protocols': []});

describe('DevicePageLoader', () => {
    it('loads the device and shows its page', async () => {
        const api = loaded();
        open(api);
        expect(await screen.findByRole('heading', {level: 1, name: 'Loaded'})).toBeInTheDocument();
        expect(api.calls[0]).toMatchObject({method: 'GET', url: '/p1/device/1'});
    });

    it('shows the edit form for the edit route', async () => {
        open(loaded(), true);
        expect(await screen.findByRole('heading', {level: 1, name: 'Edit Loaded'})).toBeInTheDocument();
    });

    it('shows the error and loads again on retry', async () => {
        let fail = true;
        const api = fakeApi({
            'GET /p1/device/1': () => fail ? new Error('no such device') : device({alias: 'Back'}),
            'GET /p1/device/1/protocols': [],
        });
        open(api);
        expect(await screen.findByText('no such device')).toBeInTheDocument();
        fail = false;
        fireEvent.click(screen.getByRole('button', {name: 'Retry'}));
        expect(await screen.findByRole('heading', {level: 1, name: 'Back'})).toBeInTheDocument();
    });

    it('does not load without a device id', () => {
        const api = fakeApi({});
        renderPage(<DevicePageLoader edit={false}/>, {api, route: 'device', url: '/project/p1/device'});
        expect(api.calls).toHaveLength(0);
    });
});
