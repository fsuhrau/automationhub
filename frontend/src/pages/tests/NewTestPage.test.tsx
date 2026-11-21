import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, describe, expect, it} from 'vitest';
import {TestType} from '../../types/test.type.enum';
import {failure, FakeApi} from './definitions/testing/fakeApi';
import {CREATE, location, openEditor} from './definitions/testing/editSupport';

let api: FakeApi;
const open = async (routes = {}): Promise<void> => {
    api = openEditor(null, routes).api;
    await screen.findByRole('heading', {name: 'New test'});
};
const name = (): HTMLElement => screen.getByRole('textbox', {name: 'Name'});
const create = (): Promise<void> => userEvent.click(screen.getByRole('button', {name: 'Create test'}));

afterEach(() => api.restore());

describe('new test', () => {
    it('starts as a Unity test on all devices with the name in focus', async () => {
        await open();
        expect(screen.getByText('A Unity test of Hero: name it, pick what it runs and where.')).toBeInTheDocument();
        expect(name()).toHaveFocus();
        expect(screen.getByRole('radio', {name: /All tests/})).toBeChecked();
        expect(screen.getByRole('radio', {name: /All devices/})).toBeChecked();
        expect(screen.getByRole('button', {name: 'Create test'})).toBeEnabled();
    });

    it('creates the test and opens it', async () => {
        await open();
        await userEvent.type(name(), 'Smoke tests');
        await create();
        await waitFor(() => expect(location()).toBe('/project/proj/app:1/test/21'));
        expect(api.calledWith(CREATE)[0].data).toMatchObject({type: TestType.Unity, name: 'Smoke tests', allDevices: true, devices: [], testFunctions: []});
    });

    it('needs a name', async () => {
        await open();
        await create();
        expect(await screen.findAllByText('The test needs a name.')).toHaveLength(2);
        expect(api.calledWith(CREATE)).toHaveLength(0);
    });

});

describe('new test errors and leaving', () => {
    it('shows the error when creating fails', async () => {
        await open({[CREATE]: () => {
            throw failure(400, 'name taken');
        }});
        await userEvent.type(name(), 'Smoke');
        await create();
        expect(await screen.findByRole('alert')).toHaveTextContent('name taken');
        expect(location()).toBe('/project/proj/app:1/test/new');
    });

    it('goes back to the list from a clean form', async () => {
        await open();
        await userEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        expect(location()).toBe('/project/proj/app:1/tests');
    });

    it('asks before a new test is thrown away', async () => {
        await open();
        await userEvent.type(name(), 'x');
        await userEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        expect(await screen.findByText("The new test isn't created yet.")).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Discard'}));
        expect(location()).toBe('/project/proj/app:1/tests');
    });
});
