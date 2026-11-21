import {expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import SettingsPage from './SettingsPage';
import {node} from './settingsFixtures';
import {Reply, failure, fakeApi} from '../../testing/fakeApi';
import {Opened, User} from '../../testing/opened';
import {renderInProject} from './settingsHarness';

const offline = node({id: 4, identifier: 'node-2', name: 'Linux box', status: 0});

const open = async (handlers: Record<string, Reply> = {}): Promise<Opened> => {
    const api = fakeApi({'GET /p1/settings/nodes': () => [node(), offline], ...handlers});
    renderInProject(<SettingsPage/>, {api, url: '/project/p1/settings?tab=nodes'});
    await screen.findByRole('table', {name: 'Nodes'});
    return {api, user: userEvent.setup()};
};

const config = (token: string): string =>
    `master_url: ${window.location.host}\nidentifier: node-3\nauth:\n  token:\n    auth_token: ${token}`;

// the config is a multi-line <pre>, which getByText would normalize
const pre = (text: string): Promise<HTMLElement> => screen.findByText((_, el) => el?.tagName === 'PRE' && el.textContent === text);

// confirms the open dialog and waits for it to be gone
const confirm = async (user: User, button: string): Promise<void> => {
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', {name: button}));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
};

it('Settings > Nodes: lists the nodes with their status and copies an identifier', async () => {
    const {user} = await open();
    const rows = within(screen.getByRole('table', {name: 'Nodes'})).getAllByRole('row');
    expect(rows[1]).toHaveTextContent('Mac mini QAConnectednode-1');
    expect(rows[2]).toHaveTextContent('Linux boxOfflinenode-2');
    await user.click(screen.getAllByRole('button', {name: 'Copy the identifier'})[0]);
    expect(await navigator.clipboard.readText()).toBe('node-1');
});

it('Settings > Nodes: says when there are no nodes', async () => {
    const api = fakeApi({'GET /p1/settings/nodes': () => []});
    renderInProject(<SettingsPage/>, {api, url: '/project/p1/settings?tab=nodes'});
    expect(await screen.findByText(/No nodes yet/)).toBeInTheDocument();
});

it('Settings > Nodes: shows the error when the nodes do not load', async () => {
    const api = fakeApi({'GET /p1/settings/nodes': () => {
        throw failure('nodes down');
    }});
    renderInProject(<SettingsPage/>, {api, url: '/project/p1/settings?tab=nodes'});
    expect(await screen.findByRole('alert')).toHaveTextContent('nodes down');
});

it('Settings > Nodes: asks for a name, then adds the node and shows its config once', async () => {
    const created = {...node({id: 8, identifier: 'node-3', name: 'New'}), token: 'node-secret'};
    const {api, user} = await open({'POST /p1/settings/node': () => created});
    await user.click(screen.getByRole('button', {name: 'Add node'}));
    expect(screen.getByText('A name is needed.')).toBeInTheDocument();
    await user.type(screen.getByLabelText('Name'), ' New ');
    await user.click(screen.getByRole('button', {name: 'Add node'}));
    expect(await pre(config('node-secret'))).toBeInTheDocument();
    expect(api.bodies('POST /p1/settings/node')).toEqual([{name: 'New'}]);
    expect(screen.getByLabelText('Name')).toHaveValue('');
    await user.click(screen.getByRole('button', {name: 'Copy the config'}));
    expect(await navigator.clipboard.readText()).toBe(config('node-secret'));
    await user.click(screen.getByRole('button', {name: 'Close'}));
    expect(screen.queryByText(/auth_token/)).toBeNull();
});

it('Settings > Nodes: shows the error when adding fails', async () => {
    const {user} = await open({'POST /p1/settings/node': () => {
        throw failure('duplicate');
    }});
    await user.type(screen.getByLabelText('Name'), 'X');
    await user.click(screen.getByRole('button', {name: 'Add node'}));
    expect(await screen.findByRole('alert')).toHaveTextContent('duplicate');
});

it('Settings > Nodes > token and delete: issues a new token after a confirmation and shows the config', async () => {
    const {api, user} = await open({'POST /p1/settings/nodes/3/token': () => ({...node(), token: 'fresh'})});
    await user.click(screen.getByRole('button', {name: 'New token for Mac mini QA'}));
    const dialog = await screen.findByRole('dialog', {name: 'New token for Mac mini QA?'});
    await user.click(within(dialog).getByRole('button', {name: 'Issue new token'}));
    expect(await screen.findByText(/auth_token: fresh/)).toBeInTheDocument();
    expect(api.calls.map(c => c.key)).toContain('POST /p1/settings/nodes/3/token');
});

it('Settings > Nodes > token and delete: shows the error when issuing a token fails', async () => {
    const {user} = await open({'POST /p1/settings/nodes/3/token': () => {
        throw failure('cannot issue');
    }});
    await user.click(screen.getByRole('button', {name: 'New token for Mac mini QA'}));
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', {name: 'Issue new token'}));
    expect(await screen.findByRole('alert', {hidden: true})).toHaveTextContent('cannot issue');
});

it('Settings > Nodes > token and delete: deletes a node after a confirmation and drops its shown config', async () => {
    const created = {...node({id: 3, identifier: 'node-3', name: 'Mac mini QA'}), token: 'node-secret'};
    const {user} = await open({'POST /p1/settings/nodes/3/token': () => created, 'DELETE /p1/settings/nodes/3': () => undefined});
    await user.click(screen.getByRole('button', {name: 'New token for Mac mini QA'}));
    await confirm(user, 'Issue new token');
    await screen.findByText(/auth_token: node-secret/);
    await user.click(screen.getByRole('button', {name: 'Delete Mac mini QA'}));
    await confirm(user, 'Delete node');
    await waitFor(() => expect(screen.queryByText(/auth_token/)).toBeNull());
    expect(screen.queryByRole('button', {name: 'Delete Mac mini QA'})).toBeNull();
});

it('Settings > Nodes > token and delete: keeps the config of another node when one is deleted', async () => {
    const created = {...node({id: 3}), token: 'node-secret'};
    const {user} = await open({'POST /p1/settings/nodes/3/token': () => created, 'DELETE /p1/settings/nodes/4': () => undefined});
    await user.click(screen.getByRole('button', {name: 'New token for Mac mini QA'}));
    await confirm(user, 'Issue new token');
    await screen.findByText(/auth_token: node-secret/);
    await user.click(screen.getByRole('button', {name: 'Delete Linux box'}));
    await confirm(user, 'Delete node');
    await waitFor(() => expect(screen.queryByRole('button', {name: 'Delete Linux box'})).toBeNull());
    expect(screen.getByText(/auth_token: node-secret/)).toBeInTheDocument();
});
