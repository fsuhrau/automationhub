import {describe, expect, it} from 'vitest';
import {fireEvent, screen, waitFor} from '@testing-library/react';
import DeviceEditPage from './DeviceEditPage';
import IDeviceData from '../../types/device';
import {device, fakeApi, FakeApi, renderPage} from './testKit';

const stored = device({alias: 'Old', customParameter: [{id: 7, key: 'PIN', value: '1234'}]});

const open = (api: FakeApi, live?: IDeviceData[]): ReturnType<typeof renderPage> =>
    renderPage(<DeviceEditPage device={stored}/>, {api, devices: live, route: 'device/:deviceId/edit', url: '/project/p1/device/1/edit'});

const save = (): void => {
    fireEvent.click(screen.getByRole('button', {name: 'Save'}));
};

describe('DeviceEditPage', () => {
    it('loads the form with the alias and the custom parameters', () => {
        open(fakeApi({}));
        expect(screen.getByRole('heading', {level: 1, name: 'Edit Old'})).toBeInTheDocument();
        expect(screen.getByLabelText('Alias')).toHaveValue('Old');
        expect(screen.getByLabelText('Key of parameter 1')).toHaveValue('PIN');
        expect(screen.getByLabelText('Value of PIN')).toHaveValue('1234');
        expect(screen.getByText('No changes yet')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Save'})).toBeDisabled();
    });

    it('says when the device has no custom parameters', () => {
        renderPage(<DeviceEditPage device={device()}/>, {route: 'device/:deviceId/edit', url: '/project/p1/device/1/edit'});
        expect(screen.getByText('No custom parameters.')).toBeInTheDocument();
        expect(screen.getByRole('heading', {level: 1, name: 'Edit Pixel'})).toBeInTheDocument();
    });

    it('rejects a parameter without a key and a key that is there already', () => {
        open(fakeApi({}));
        fireEvent.click(screen.getByRole('button', {name: 'Add parameter'}));
        expect(screen.getByText('A key is needed')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Save'})).toBeDisabled();
        fireEvent.change(screen.getByLabelText('Key of parameter 2'), {target: {value: 'PIN'}});
        expect(screen.getByText('This key is already there')).toBeInTheDocument();
        fireEvent.change(screen.getByLabelText('Key of parameter 2'), {target: {value: 'LANG'}});
        expect(screen.queryByText('This key is already there')).toBeNull();
        expect(screen.getByRole('button', {name: 'Save'})).toBeEnabled();
    });

});

describe('DeviceEditPage saving', () => {
    it('saves the changes and opens the device', async () => {
        const saved = {...stored, alias: 'New', customParameter: [{id: 7, key: 'PIN', value: '9999'}]};
        const api = fakeApi({'PUT /p1/device/1': saved});
        open(api, [{...stored, isLocked: true}]);
        fireEvent.change(screen.getByLabelText('Alias'), {target: {value: ' New '}});
        fireEvent.change(screen.getByLabelText('Value of PIN'), {target: {value: '9999'}});
        expect(screen.getByText('Unsaved changes')).toBeInTheDocument();
        save();
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/device/1');
        expect(api.calls.at(-1)).toMatchObject({
            method: 'PUT', url: '/p1/device/1', data: {alias: 'New', customParameter: [{id: 7, key: 'PIN', value: '9999'}]},
        });
    });

    it('saves with Ctrl+S, also when the device is not in the list yet', async () => {
        const api = fakeApi({'PUT /p1/device/1': {...stored, alias: 'Short'}});
        open(api);
        fireEvent.change(screen.getByLabelText('Alias'), {target: {value: 'Short'}});
        fireEvent.keyDown(screen.getByLabelText('Alias'), {key: 'x', ctrlKey: true});
        fireEvent.keyDown(screen.getByLabelText('Alias'), {key: 's'});
        expect(api.calls).toHaveLength(0);
        fireEvent.keyDown(screen.getByLabelText('Alias'), {key: 's', ctrlKey: true});
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/device/1');
    });

    it('does not save an unchanged or invalid form', () => {
        const api = fakeApi({});
        open(api);
        fireEvent.keyDown(screen.getByLabelText('Alias'), {key: 's', metaKey: true});
        fireEvent.click(screen.getByRole('button', {name: 'Add parameter'}));
        fireEvent.keyDown(screen.getByLabelText('Alias'), {key: 's', metaKey: true});
        expect(api.calls).toHaveLength(0);
    });

});

describe('DeviceEditPage save errors', () => {
    it('shows why saving failed and stays on the form', async () => {
        open(fakeApi({'PUT /p1/device/1': new Error('alias taken')}));
        fireEvent.change(screen.getByLabelText('Alias'), {target: {value: 'Dup'}});
        save();
        expect(await screen.findByText('alias taken')).toBeInTheDocument();
        expect(screen.queryByTestId('location')).toBeNull();
        await waitFor(() => expect(screen.getByRole('button', {name: 'Save'})).toBeEnabled());
    });

});

describe('DeviceEditPage leaving', () => {
    it('removes a parameter', () => {
        open(fakeApi({}));
        fireEvent.click(screen.getByRole('button', {name: 'Remove PIN'}));
        expect(screen.getByText('No custom parameters.')).toBeInTheDocument();
        expect(screen.getByText('Unsaved changes')).toBeInTheDocument();
    });

    it('leaves without asking when nothing changed', async () => {
        open(fakeApi({}));
        fireEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/device/1');
    });

});

describe('DeviceEditPage discarding', () => {
    it('asks before throwing changes away', async () => {
        open(fakeApi({}));
        fireEvent.change(screen.getByLabelText('Alias'), {target: {value: 'Changed'}});
        fireEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        expect(screen.getByText('Discard your changes?')).toBeInTheDocument();
        fireEvent.click(screen.getAllByRole('button', {name: 'Cancel'}).at(-1) as HTMLElement);
        await waitFor(() => expect(screen.queryByText('Discard your changes?')).toBeNull());
        fireEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        fireEvent.click(screen.getByRole('button', {name: 'Discard'}));
        expect(await screen.findByTestId('location')).toHaveTextContent('/project/p1/device/1');
    });

    it('warns before the tab closes with unsaved changes', () => {
        open(fakeApi({}));
        const clean = new Event('beforeunload', {cancelable: true});
        window.dispatchEvent(clean);
        expect(clean.defaultPrevented).toBe(false);
        fireEvent.change(screen.getByLabelText('Alias'), {target: {value: 'Changed'}});
        const dirty = new Event('beforeunload', {cancelable: true});
        window.dispatchEvent(dirty);
        expect(dirty.defaultPrevented).toBe(true);
    });
});
