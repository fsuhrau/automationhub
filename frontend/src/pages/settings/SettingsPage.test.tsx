import {expect, it} from 'vitest';
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import SettingsPage from './SettingsPage';
import {apiError, createFakeApi, Handler, Opened, project, renderInProject} from './testKit';

const open = (url: string, handlers: Record<string, Handler> = {}): Opened => {
    const api = createFakeApi({
        'GET /p1/apps': () => [], 'GET /p1/settings/access_tokens': () => [],
        'GET /p1/settings/nodes': () => [], 'GET /p1/settings/users': () => [], ...handlers,
    });
    renderInProject(<SettingsPage/>, {api, url});
    return {api, user: userEvent.setup()};
};

it('SettingsPage tabs: opens the project tab by default and describes the project', async () => {
    open('/project/p1/settings');
    expect(await screen.findByRole('heading', {name: 'Settings'})).toBeInTheDocument();
    expect(screen.getByText(/How Heroes is set up/)).toBeInTheDocument();
    expect(screen.getByRole('tab', {name: 'Project', selected: true})).toBeInTheDocument();
    expect(screen.getByText('p1')).toBeInTheDocument();
});

it('SettingsPage tabs: falls back to the project tab for an unknown one', async () => {
    open('/project/p1/settings?tab=bogus');
    expect(await screen.findByRole('tab', {name: 'Project', selected: true})).toBeInTheDocument();
});

it.each([
    ['Apps', /The builds this project tests/],
    ['Access tokens', /Tokens for scripts, CI and agents/],
    ['Nodes', /Computers that connect phones/],
    ['People', /Everyone with an account/],
    ['Project', /How the project is called/],
])('SettingsPage tabs: loads the %s tab when it is picked', async (tab, text) => {
    const {user} = open('/project/p1/settings');
    await user.click(await screen.findByRole('tab', {name: tab}));
    expect(await screen.findByText(text)).toBeInTheDocument();
    expect(screen.getByRole('tab', {name: tab, selected: true})).toBeInTheDocument();
});

it('SettingsPage tabs: opens the tab the URL names', async () => {
    const {api} = open('/project/p1/settings?tab=nodes');
    expect(await screen.findByRole('tab', {name: 'Nodes', selected: true})).toBeInTheDocument();
    await waitFor(() => expect(api.calls.map(c => c.key)).toContain('GET /p1/settings/nodes'));
});

it('Settings > Project: copies the identifier', async () => {
    const {user} = open('/project/p1/settings');
    await user.click(await screen.findByRole('button', {name: 'Copy the identifier'}));
    expect(await navigator.clipboard.readText()).toBe('p1');
});

it('Settings > Project: renames the project and shows the new name', async () => {
    const {api, user} = open('/project/p1/settings', {'PUT /project/p1': (body: unknown) => body});
    const name = await screen.findByLabelText('Name');
    expect(screen.queryByRole('button', {name: 'Save'})).toBeNull();
    await user.clear(name);
    await user.type(name, ' Renamed ');
    await user.click(screen.getByRole('button', {name: 'Save'}));
    await waitFor(() => expect(screen.getByText(/How Renamed is set up/)).toBeInTheDocument());
    expect(api.bodies('PUT /project/p1')[0]).toMatchObject({...project(), name: 'Renamed'});
    expect(screen.queryByRole('button', {name: 'Save'})).toBeNull();
    expect(screen.getByLabelText('Name')).toHaveValue('Renamed');
});

it('Settings > Project: saves with Enter and can undo a draft', async () => {
    const {api, user} = open('/project/p1/settings', {'PUT /project/p1': (body: unknown) => body});
    const name = await screen.findByLabelText('Name');
    await user.type(name, 'x');
    await user.click(screen.getByRole('button', {name: 'Undo'}));
    expect(name).toHaveValue('Heroes');
    await user.type(name, 'y{Enter}');
    await waitFor(() => expect(api.bodies('PUT /project/p1')).toHaveLength(1));
});

it('Settings > Project: does not offer saving an empty name', async () => {
    const {user} = open('/project/p1/settings');
    await user.clear(await screen.findByLabelText('Name'));
    expect(screen.getByText('A name is needed.')).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: 'Save'})).toBeNull();
});

it('Settings > Project: does nothing when the form is submitted unchanged', async () => {
    const {api, user} = open('/project/p1/settings');
    await user.type(await screen.findByLabelText('Name'), '{Enter}');
    expect(api.bodies('PUT /project/p1')).toEqual([]);
});

it('Settings > Project: shows the master\'s error and keeps the draft when renaming fails', async () => {
    const {user} = open('/project/p1/settings', {'PUT /project/p1': () => {
        throw apiError('name taken');
    }});
    const name = await screen.findByLabelText('Name');
    await user.type(name, '2');
    await user.click(screen.getByRole('button', {name: 'Save'}));
    expect(await screen.findByRole('alert')).toHaveTextContent('name taken');
    expect(name).toHaveValue('Heroes2');
});
