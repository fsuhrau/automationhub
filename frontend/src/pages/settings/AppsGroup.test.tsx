import {expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import SettingsPage from './SettingsPage';
import {app, parameter} from './settingsFixtures';
import {Reply, failure, fakeApi} from '../../testing/fakeApi';
import {Opened} from '../../testing/opened';
import {renderInProject} from './settingsHarness';

const URL = '/project/p1/settings?tab=apps';
const android = app({parameter: [parameter(), parameter({id: 11, name: 'Mode', type: {type: 'option', options: ['a', 'b'], defaultValue: 'a'}})]});
const web = app({id: 2, name: 'Game Web', identifier: '', platform: 5});

const open = async (handlers: Record<string, Reply> = {}): Promise<Opened> => {
    const api = fakeApi({'GET /p1/apps': () => [web, android], ...handlers});
    renderInProject(<SettingsPage/>, {api, url: URL});
    await screen.findByRole('list', {name: 'Apps'});
    return {api, user: userEvent.setup()};
};

it('Settings > Apps: lists the apps by platform and shows the first one with its parameters', async () => {
    await open();
    const items = within(screen.getByRole('list', {name: 'Apps'})).getAllByRole('button');
    expect(items.map(i => i.textContent)).toEqual(['Game Androidcom.game', 'Game WebWeb']);
    const table = screen.getByRole('table', {name: 'Parameters'});
    expect(within(table).getByText('MarketId')).toBeInTheDocument();
    expect(within(table).getByText('Choice of 2')).toBeInTheDocument();
    expect(screen.getByRole('link', {name: 'Tests'})).toHaveAttribute('href', '/project/p1/app:1/tests');
});

it('Settings > Apps: selects another app and says when it has no parameters', async () => {
    const {user} = await open();
    await user.click(screen.getByRole('button', {name: /Game Web/}));
    expect(screen.getByText('None.')).toBeInTheDocument();
    expect(screen.queryByText('Package name')).toBeNull();
});

it('Settings > Apps: says when the project has no apps', async () => {
    const api = fakeApi({'GET /p1/apps': () => []});
    renderInProject(<SettingsPage/>, {api, url: URL});
    expect(await screen.findByText(/No apps yet/)).toBeInTheDocument();
});

it('Settings > Apps: shows the error when the apps do not load', async () => {
    const api = fakeApi({'GET /p1/apps': () => {
        throw failure('apps are gone');
    }});
    renderInProject(<SettingsPage/>, {api, url: URL});
    expect(await screen.findByRole('alert')).toHaveTextContent('apps are gone');
});

it('Settings > Apps: renames the app and can undo a draft', async () => {
    const {api, user} = await open({'PUT /p1/app/1': (body: unknown) => body});
    const name = screen.getByLabelText('Name');
    await user.clear(name);
    await user.type(name, 'Renamed');
    await user.click(screen.getByRole('button', {name: 'Undo'}));
    expect(name).toHaveValue('Game Android');
    await user.type(name, '2');
    await user.click(screen.getByRole('button', {name: 'Save'}));
    await waitFor(() => expect(screen.getByRole('button', {name: /Game Android2/})).toBeInTheDocument());
    expect(api.bodies('PUT /p1/app/1')[0]).toMatchObject({id: 1, name: 'Game Android2'});
});

it('Settings > Apps: shows a failed rename in the error toast', async () => {
    const {user} = await open({'PUT /p1/app/1': () => {
        throw failure('nope');
    }});
    await user.type(screen.getByLabelText('Name'), 'x');
    await user.click(screen.getByRole('button', {name: 'Save'}));
    expect(await screen.findByRole('alert')).toHaveTextContent('nope');
});

it('Settings > Apps: does not offer saving an empty name', async () => {
    const {user} = await open();
    await user.clear(screen.getByLabelText('Name'));
    expect(screen.getByText('A name is needed.')).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: 'Save'})).toBeNull();
});

it('Settings > Apps > add an app: checks the form, creates the app and selects it', async () => {
    const created = app({id: 7, name: 'New iOS', identifier: 'com.ios', platform: 0});
    const {api, user} = await open({'POST /p1/app': () => created});
    await user.click(screen.getByRole('button', {name: 'Add app'}));
    const dialog = await screen.findByRole('dialog', {name: 'Add an app'});
    await user.click(within(dialog).getByRole('button', {name: 'Add app'}));
    expect(within(dialog).getByText('A name is needed.')).toBeInTheDocument();
    expect(within(dialog).getByText(/package name is needed/)).toBeInTheDocument();

    await user.type(within(dialog).getByLabelText('Name'), 'New iOS');
    await user.click(within(dialog).getByRole('combobox', {name: /Platform/}));
    await user.click(screen.getByRole('option', {name: 'iOS'}));
    await user.type(within(dialog).getByLabelText('Bundle identifier'), 'com.ios');
    await user.click(within(dialog).getByRole('button', {name: 'Add app'}));

    await waitFor(() => expect(screen.queryByRole('dialog', {name: 'Add an app'})).toBeNull());
    expect(api.bodies('POST /p1/app')).toEqual([{name: 'New iOS', platform: 0, identifier: 'com.ios'}]);
    expect(screen.getByRole('button', {name: /New iOS/})).toHaveAttribute('class', expect.stringContaining('Mui-selected'));
});

it.each([
    ['Web', 'Web apps run in a browser'],
    ['Unity editor', 'Unity editor apps run in the editor'],
])('Settings > Apps > add an app: needs no identifier for %s', async (platform, note) => {
    const {user} = await open();
    await user.click(screen.getByRole('button', {name: 'Add app'}));
    const dialog = await screen.findByRole('dialog', {name: 'Add an app'});
    await user.click(within(dialog).getByRole('combobox', {name: /Platform/}));
    await user.click(screen.getByRole('option', {name: platform}));
    expect(within(dialog).getByText(new RegExp(note))).toBeInTheDocument();
});

it('Settings > Apps > add an app: shows the master\'s error in the dialog and keeps it open', async () => {
    const {user} = await open({'POST /p1/app': () => {
        throw failure('name taken');
    }});
    await user.click(screen.getByRole('button', {name: 'Add app'}));
    const dialog = await screen.findByRole('dialog', {name: 'Add an app'});
    await user.type(within(dialog).getByLabelText('Name'), 'X');
    await user.type(within(dialog).getByLabelText('Package name'), 'a.b');
    await user.click(within(dialog).getByRole('button', {name: 'Add app'}));
    expect(await within(dialog).findByText('name taken')).toBeInTheDocument();
    await user.click(within(dialog).getByRole('button', {name: 'Cancel'}));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
});
