import {expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import dayjs, {Dayjs} from 'dayjs';
import SettingsPage from './SettingsPage';
import {accessToken, apiError, createFakeApi, Handler, Opened, renderInProject} from './testKit';

const expired = accessToken({id: 6, name: 'Old', token: 'zzzzzzzzzzzz', expiresAt: new Date('2020-01-01')});
const valid = accessToken({expiresAt: new Date('2999-01-01')});

const open = async (handlers: Record<string, Handler> = {}): Promise<Opened> => {
    const api = createFakeApi({'GET /p1/settings/access_tokens': () => [valid, expired], ...handlers});
    renderInProject(<SettingsPage/>, {api, url: '/project/p1/settings?tab=tokens'});
    await screen.findByRole('table', {name: 'Access tokens'});
    return {api, user: userEvent.setup()};
};

it('Settings > Access tokens: lists the tokens masked, with their expiry', async () => {
    await open();
    const rows = within(screen.getByRole('table', {name: 'Access tokens'})).getAllByRole('row');
    expect(rows[1]).toHaveTextContent('CI nightly••••456789');
    expect(rows[1]).toHaveTextContent('2024/02/03');
    expect(rows[1]).toHaveTextContent('2999/01/01');
    expect(rows[2]).toHaveTextContent('expired');
    expect(screen.getByRole('link', {name: 'profile'})).toHaveAttribute('href', '/profile');
});

it('Settings > Access tokens: shows and hides a token and copies it', async () => {
    const {user} = await open();
    const [show] = screen.getAllByRole('button', {name: 'Show the token'});
    await user.click(show);
    expect(screen.getByText('abcdef123456789')).toBeInTheDocument();
    await user.click(screen.getAllByRole('button', {name: 'Copy the token'})[0]);
    expect(await navigator.clipboard.readText()).toBe('abcdef123456789');
    await user.click(screen.getByRole('button', {name: 'Hide the token'}));
    expect(screen.queryByText('abcdef123456789')).toBeNull();
});

it('Settings > Access tokens: says when there are no tokens', async () => {
    const api = createFakeApi({'GET /p1/settings/access_tokens': () => []});
    renderInProject(<SettingsPage/>, {api, url: '/project/p1/settings?tab=tokens'});
    expect(await screen.findByText('No tokens yet.')).toBeInTheDocument();
});

it('Settings > Access tokens: shows the error when the tokens do not load', async () => {
    const api = createFakeApi({'GET /p1/settings/access_tokens': () => {
        throw apiError('no tokens for you');
    }});
    renderInProject(<SettingsPage/>, {api, url: '/project/p1/settings?tab=tokens'});
    expect(await screen.findByRole('alert')).toHaveTextContent('no tokens for you');
});

it('Settings > Access tokens > create: asks for a name first', async () => {
    const {api, user} = await open();
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    expect(screen.getByText('A name is needed.')).toBeInTheDocument();
    expect(api.bodies('POST /p1/settings/access_token')).toEqual([]);
});

it('Settings > Access tokens > create: creates a token, shows the command with it once and empties the form', async () => {
    const created = accessToken({id: 9, name: 'Nightly', token: 'secret-token-value'});
    const {api, user} = await open({'POST /p1/settings/access_token': () => created});
    await user.type(screen.getByLabelText('Name'), ' Nightly ');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    expect(await screen.findByText(/is ready/)).toBeInTheDocument();
    expect(api.bodies('POST /p1/settings/access_token')).toEqual([{name: 'Nightly', expiresAt: null}]);
    const command = `curl -H "X-Auth-Token: secret-token-value" ${window.location.origin}/api/p1/devices`;
    expect(screen.getByText(command)).toBeInTheDocument();
    expect(screen.getByLabelText('Name')).toHaveValue('');
    await user.click(screen.getByRole('button', {name: 'Copy the command'}));
    expect(await navigator.clipboard.readText()).toBe(command);
    expect(within(screen.getByRole('table', {name: 'Access tokens'})).getByText('Nightly')).toBeInTheDocument();
});

it('Settings > Access tokens > create: sends the expiry that was picked', async () => {
    const {api, user} = await open({'POST /p1/settings/access_token': () => accessToken({id: 9})});
    await user.type(screen.getByLabelText('Name'), 'Dated');
    await user.type(screen.getByLabelText(/Expires/), '12/31/2099');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    await waitFor(() => expect(api.bodies('POST /p1/settings/access_token')).toHaveLength(1));
    const body = api.bodies('POST /p1/settings/access_token')[0] as { expiresAt: Dayjs };
    expect(dayjs.isDayjs(body.expiresAt) && body.expiresAt.format('YYYY-MM-DD')).toBe('2099-12-31');
});

it('Settings > Access tokens > create: shows the error and keeps the form when creating fails', async () => {
    const {user} = await open({'POST /p1/settings/access_token': () => {
        throw apiError('quota reached');
    }});
    await user.type(screen.getByLabelText('Name'), 'X');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    expect(await screen.findByRole('alert')).toHaveTextContent('quota reached');
    expect(screen.getByLabelText('Name')).toHaveValue('X');
});

it('Settings > Access tokens > revoke: revokes a token after a confirmation and drops its command', async () => {
    const created = accessToken({id: 9, name: 'Nightly', token: 'secret-token-value'});
    const {api, user} = await open({
        'POST /p1/settings/access_token': () => created,
        'DELETE /p1/settings/access_token/9': () => undefined,
    });
    await user.type(screen.getByLabelText('Name'), 'Nightly');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    await screen.findByText(/is ready/);
    await user.click(screen.getByRole('button', {name: 'Revoke Nightly'}));
    const dialog = await screen.findByRole('dialog', {name: 'Revoke Nightly?'});
    await user.click(within(dialog).getByRole('button', {name: 'Revoke token'}));
    await waitFor(() => expect(screen.queryByText(/is ready/)).toBeNull());
    expect(api.calls.map(c => c.key)).toContain('DELETE /p1/settings/access_token/9');
    expect(screen.queryByRole('button', {name: 'Revoke Nightly'})).toBeNull();
});

it('Settings > Access tokens > revoke: keeps the command of another token when one is revoked', async () => {
    const created = accessToken({id: 9, name: 'Nightly', token: 'secret-token-value'});
    const {user} = await open({
        'POST /p1/settings/access_token': () => created,
        'DELETE /p1/settings/access_token/6': () => undefined,
    });
    await user.type(screen.getByLabelText('Name'), 'Nightly');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    await screen.findByText(/is ready/);
    await user.click(screen.getByRole('button', {name: 'Revoke Old'}));
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', {name: 'Revoke token'}));
    await waitFor(() => expect(screen.queryByRole('button', {name: 'Revoke Old'})).toBeNull());
    expect(screen.getByText(/is ready/)).toBeInTheDocument();
});

it('Settings > Access tokens > revoke: can be cancelled, and shows the error when it fails', async () => {
    const {user} = await open({'DELETE /p1/settings/access_token/5': () => {
        throw apiError('in use');
    }});
    await user.click(screen.getAllByRole('button', {name: 'Revoke CI nightly'})[0]);
    let dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('button', {name: 'Cancel'}));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    await user.click(screen.getAllByRole('button', {name: 'Revoke CI nightly'})[0]);
    dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('button', {name: 'Revoke token'}));
    expect(await screen.findByRole('alert', {hidden: true})).toHaveTextContent('in use');
    expect(screen.getByRole('dialog')).toBeInTheDocument();
});
