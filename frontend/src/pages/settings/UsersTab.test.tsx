import {expect, it, vi} from 'vitest';
import {screen, within} from '@testing-library/react';
import SettingsPage from './SettingsPage';
import {FakeApi, failure, fakeApi} from '../../testing/fakeApi';
import {renderInProject} from './settingsHarness';

const auth = vi.hoisted(() => ({user: null as { name: string } | null}));
vi.mock('../../hooks/AuthProvider', () => ({useAuth: () => ({user: auth.user})}));

const users = [
    {id: 1, name: 'Ann', email: 'ann@x.io', role: 'admin'},
    {id: 2, name: 'Bob', email: 'bob@x.io', role: ''},
];

const open = (handler: () => unknown): FakeApi => {
    const api = fakeApi({'GET /p1/settings/users': handler});
    renderInProject(<SettingsPage/>, {api, url: '/project/p1/settings?tab=people'});
    return api;
};

it('Settings > People: lists the accounts and marks the signed-in person', async () => {
    auth.user = {name: 'Bob'};
    open(() => users);
    const table = await screen.findByRole('table', {name: 'People'});
    const rows = within(table).getAllByRole('row');
    expect(rows[1]).toHaveTextContent('AAnnann@x.ioadmin');
    expect(rows[2]).toHaveTextContent('BBobyoubob@x.io–');
    expect(within(rows[1]).queryByText('you')).toBeNull();
});

it('Settings > People: marks nobody when nobody is signed in', async () => {
    auth.user = null;
    open(() => users);
    await screen.findByRole('table', {name: 'People'});
    expect(screen.queryByText('you')).toBeNull();
});

it('Settings > People: says when the hub has no accounts', async () => {
    open(() => []);
    expect(await screen.findByText(/this hub runs without sign-in/)).toBeInTheDocument();
});

it('Settings > People: shows the error when the accounts do not load', async () => {
    open(() => {
        throw failure('not allowed');
    });
    expect(await screen.findByRole('alert')).toHaveTextContent('not allowed');
});
