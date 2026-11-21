import {beforeEach, expect, it, vi} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ProfilePage from './ProfilePage';
import {apiError, createFakeApi, Handler, Opened, project, renderOutsideProject} from '../settings/testKit';

vi.mock('../../components/TopBar', () => ({default: () => <header>top bar</header>}));

const auth = vi.hoisted(() => ({
    value: {} as { authState: string, user: object | null, logOut: () => void },
}));
vi.mock('../../hooks/AuthProvider', () => ({useAuth: () => auth.value}));

const logOut = vi.fn();
const profile = {id: 1, name: 'Anna Admin', login: 'anna', email: 'anna@x.io'};
const token = {id: 5, name: 'claude-code', hint: 'a1b2', createdAt: '2024-02-03T10:00:00Z', expiresAt: null, lastUsedAt: null};
const used = {...token, id: 6, name: 'ci', hint: 'zz99', expiresAt: '2020-01-01T00:00:00Z', lastUsedAt: '2024-02-04T10:00:00Z'};

beforeEach(() => {
    logOut.mockClear();
    auth.value = {authState: 'authenticated', user: {name: 'Anna', login: 'anna', email: 'a@x.io', avatar: ''}, logOut};
});

const open = async (handlers: Record<string, Handler> = {}): Promise<Opened> => {
    const api = createFakeApi({
        'GET /me': () => profile, 'GET /me/tokens': () => [token, used],
        'GET /projects': () => [project(), project({id: 2, identifier: 'p2', name: 'Second'})], ...handlers,
    });
    renderOutsideProject(<ProfilePage/>, api);
    return {api, user: userEvent.setup()};
};

it('ProfilePage: shows the account from the profile and signs out', async () => {
    const {user} = await open();
    expect(await screen.findByText('Anna Admin')).toBeInTheDocument();
    expect(screen.getByText('anna@x.io')).toBeInTheDocument();
    expect(screen.getByText('signed in as anna')).toBeInTheDocument();
    await user.click(screen.getByRole('button', {name: 'Sign out'}));
    expect(logOut).toHaveBeenCalled();
});

it('ProfilePage: lists the personal tokens with last use and expiry', async () => {
    await open();
    const table = await screen.findByRole('table', {name: 'Personal tokens'});
    const rows = within(table).getAllByRole('row');
    expect(rows[1]).toHaveTextContent('claude-codeahp_…a1b2never2024/02/03never');
    expect(rows[2]).toHaveTextContent('ciahp_…zz99');
    expect(rows[2]).toHaveTextContent('expired');
    expect(rows[2]).not.toHaveTextContent('never');
});

it('ProfilePage: says when there are no personal tokens', async () => {
    await open({'GET /me/tokens': () => []});
    expect(await screen.findByText('No personal tokens yet.')).toBeInTheDocument();
});

it('ProfilePage: shows the error of the tokens in the toast', async () => {
    await open({'GET /me/tokens': () => {
        throw apiError('tokens unavailable');
    }});
    expect(await screen.findByRole('alert')).toHaveTextContent('tokens unavailable');
});

it('ProfilePage: warns when the profile does not load and offers no tokens', async () => {
    await open({'GET /me': () => {
        throw apiError('profile unavailable');
    }});
    expect(await screen.findByText('profile unavailable')).toBeInTheDocument();
    expect(screen.queryByText('Personal access tokens')).toBeNull();
});

it('ProfilePage: explains a hub without sign-in and loads nothing', async () => {
    auth.value = {authState: 'disabled', user: null, logOut};
    const {api} = await open();
    expect(screen.getByText(/runs without sign-in/)).toBeInTheDocument();
    expect(screen.getByRole('link', {name: 'Pick a project'})).toHaveAttribute('href', '/');
    expect(screen.queryByText('Account')).toBeNull();
    expect(api.calls).toEqual([]);
});

it('ProfilePage > create and revoke tokens: asks for a name first', async () => {
    const {api, user} = await open();
    await user.click(await screen.findByRole('button', {name: 'Create token'}));
    expect(screen.getByText('A name is needed.')).toBeInTheDocument();
    expect(api.bodies('POST /me/tokens')).toEqual([]);
});

it('ProfilePage > create and revoke tokens: creates a token, shows it once, fills it into the guide and reloads the list', async () => {
    const created = {...token, id: 9, name: 'laptop', token: 'ahp_secret_value'};
    const {api, user} = await open({'POST /me/tokens': () => created});
    await user.type(await screen.findByLabelText('Name'), ' laptop ');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    expect(await screen.findByText(/is ready/)).toBeInTheDocument();
    expect(api.bodies('POST /me/tokens')).toEqual([{name: 'laptop', expiresAt: null}]);
    await waitFor(() => expect(api.bodies('GET /me/tokens')).toHaveLength(2));
    expect(screen.getByLabelText('Name')).toHaveValue('');
    await user.click(screen.getByRole('button', {name: 'Copy the token'}));
    expect(await navigator.clipboard.readText()).toBe('ahp_secret_value');
    await user.click(screen.getByRole('button', {name: 'Directly on this computer'}));
    expect(screen.getByText(/--header "X-Auth-Token: ahp_secret_value"/)).toBeInTheDocument();
    expect(screen.getByText('Your new token is filled in.')).toBeInTheDocument();
    await user.click(screen.getByRole('button', {name: 'Close'}));
    expect(screen.queryByText(/is ready/)).toBeNull();
});

it('ProfilePage > create and revoke tokens: sends the expiry as the end of the picked day', async () => {
    const {api, user} = await open({'POST /me/tokens': () => ({...token, id: 9, token: 't'})});
    await user.type(await screen.findByLabelText('Name'), 'dated');
    await user.type(screen.getByLabelText(/Expires/), '12/31/2099');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    await waitFor(() => expect(api.bodies('POST /me/tokens')).toHaveLength(1));
    const body = api.bodies('POST /me/tokens')[0] as { expiresAt: string };
    expect(new Date(body.expiresAt).getFullYear()).toBe(2099);
});

it('ProfilePage > create and revoke tokens: shows the error and keeps the form when creating fails', async () => {
    const {user} = await open({'POST /me/tokens': () => {
        throw apiError('too many tokens');
    }});
    await user.type(await screen.findByLabelText('Name'), 'x');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    expect(await screen.findByRole('alert')).toHaveTextContent('too many tokens');
    expect(screen.getByLabelText('Name')).toHaveValue('x');
});

it('ProfilePage > create and revoke tokens: revokes a token after a confirmation and reloads the list', async () => {
    const {api, user} = await open({'DELETE /me/tokens/5': () => undefined});
    await user.click(await screen.findByRole('button', {name: 'Revoke claude-code'}));
    const dialog = await screen.findByRole('dialog', {name: 'Revoke claude-code?'});
    await user.click(within(dialog).getByRole('button', {name: 'Revoke token'}));
    await waitFor(() => expect(api.bodies('GET /me/tokens')).toHaveLength(2));
    expect(api.calls.map(c => c.key)).toContain('DELETE /me/tokens/5');
});

it('ProfilePage > create and revoke tokens: drops the shown token when that one is revoked, and keeps it for another', async () => {
    const created = {...token, id: 9, name: 'laptop', token: 'ahp_secret_value'};
    const {user} = await open({'POST /me/tokens': () => created, 'DELETE /me/tokens/5': () => undefined});
    await user.type(await screen.findByLabelText('Name'), 'laptop');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    await screen.findByText(/is ready/);
    await user.click(screen.getByRole('button', {name: 'Revoke claude-code'}));
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', {name: 'Revoke token'}));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(screen.getByText(/is ready/)).toBeInTheDocument();
});

it('ProfilePage > create and revoke tokens: drops the shown token when it is revoked itself', async () => {
    const created = {...token, id: 5, name: 'claude-code', token: 'ahp_secret_value'};
    const {user} = await open({'POST /me/tokens': () => created, 'DELETE /me/tokens/5': () => undefined});
    await user.type(await screen.findByLabelText('Name'), 'claude-code');
    await user.click(screen.getByRole('button', {name: 'Create token'}));
    await screen.findByText(/is ready/);
    await user.click(screen.getAllByRole('button', {name: 'Revoke claude-code'})[0]);
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', {name: 'Revoke token'}));
    await waitFor(() => expect(screen.queryByText(/is ready/)).toBeNull());
});

it('ProfilePage > create and revoke tokens: keeps the dialog and shows the error when revoking fails', async () => {
    const {user} = await open({'DELETE /me/tokens/5': () => {
        throw apiError('cannot revoke');
    }});
    await user.click(await screen.findByRole('button', {name: 'Revoke claude-code'}));
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', {name: 'Revoke token'}));
    expect(await screen.findByRole('alert', {hidden: true})).toHaveTextContent('cannot revoke');
    expect(screen.getByRole('dialog')).toBeInTheDocument();
});

it('ProfilePage > Claude Code guide: guides a dev container with the keychain and the proxy of the first project', async () => {
    const {user} = await open();
    await screen.findByText('Connect Claude Code');
    expect(screen.getByText(`security add-generic-password -U -s automationhub -a ${window.location.host} -w`)).toBeInTheDocument();
    expect(screen.getByText(`automationhub mcp-proxy --upstream ${window.location.origin} --project p1`)).toBeInTheDocument();
    await user.click(screen.getAllByRole('button', {name: 'Copy the command'})[1]);
    expect(await navigator.clipboard.readText()).toContain('--project p1');
});

it('ProfilePage > Claude Code guide: uses the picked project and a placeholder before a token exists', async () => {
    const {user} = await open();
    await screen.findByText('Connect Claude Code');
    await user.click(screen.getByRole('combobox', {name: /Project/}));
    await user.click(screen.getByRole('option', {name: 'Second'}));
    await user.click(screen.getByRole('button', {name: 'Directly on this computer'}));
    expect(screen.getByText(/api\/p2\/mcp/)).toBeInTheDocument();
    expect(screen.getByText(/<your personal token>/, {selector: 'pre'})).toBeInTheDocument();
});

it('ProfilePage > Claude Code guide: works without any project', async () => {
    const {user} = await open({'GET /projects': () => []});
    await user.click(await screen.findByRole('button', {name: 'Directly on this computer'}));
    expect(screen.queryByRole('combobox', {name: /Project/})).toBeNull();
    expect(screen.getByText(/api\/%3Cproject%3E\/mcp/)).toBeInTheDocument();
});
