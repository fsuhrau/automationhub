import {expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IAppBinaryData} from '../../types/app';
import {bundle} from './appsFixtures';
import {failure, fakeApi, FakeApi} from '../../testing/fakeApi';
import {renderBundles} from './appsHarness';

// a master that keeps the bundles, so a list loaded after a change shows it
const master = (store: { list: IAppBinaryData[] }, extra: Parameters<typeof fakeApi>[0] = {}): FakeApi =>
    fakeApi({'GET /demo/app/3/bundles': () => store.list, ...extra});

const choose = async (id: number, item: RegExp | string): Promise<void> => {
    await userEvent.click(await screen.findByRole('button', {name: `More for #${id}`}));
    await userEvent.click(screen.getByRole('menuitem', {name: item}));
};

const confirmWith = async (name: string): Promise<void> => {
    await userEvent.click(within(await screen.findByRole('dialog')).getByRole('button', {name}));
};

it('deletes a bundle after the question and reloads the list', async () => {
    const store = {list: [bundle(1, {runCount: 1}), bundle(2)]};
    const api = master(store, {'DELETE /demo/app/3/bundle/2': () => {
        store.list = [bundle(1)];
    }});
    renderBundles(api, 'bundles');
    await choose(2, 'Delete bundle');
    expect(screen.getByText(/Delete version 1.2.0/)).toBeInTheDocument();
    await confirmWith('Delete bundle');
    await waitFor(() => expect(screen.queryByText('1.2.0')).not.toBeInTheDocument());
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
});

it('keeps a bundle when the question is cancelled', async () => {
    renderBundles(master({list: [bundle(1)]}), 'bundles');
    await choose(1, 'Delete bundle');
    await confirmWith('Cancel');
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    expect(screen.getByText('1.1.0')).toBeInTheDocument();
});

it('shows why a bundle could not be deleted and keeps the question open', async () => {
    const api = master({list: [bundle(1)]}, {'DELETE /demo/app/3/bundle/1': () => {
        throw failure('bundle is in use');
    }});
    renderBundles(api, 'bundles');
    await choose(1, 'Delete bundle');
    await confirmWith('Delete bundle');
    expect(await screen.findByText('bundle is in use')).toBeInTheDocument();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
});

it('archives a bundle and shows it as archived', async () => {
    const store = {list: [bundle(1)]};
    const api = master(store, {'POST /demo/app/3/bundle/1/archive': () => {
        store.list = [bundle(1, {archivedAt: '2020-01-01T00:00:00Z'})];
    }});
    renderBundles(api, 'bundles');
    await choose(1, /Archive bundle/);
    await confirmWith('Archive bundle');
    expect(await screen.findByText('archived')).toBeInTheDocument();
    expect(api.calls.some(c => c.url.endsWith('/archive'))).toBe(true);
});

it('does not offer to archive a bundle twice', async () => {
    renderBundles(master({list: [bundle(1, {archivedAt: '2020-01-01T00:00:00Z'})]}), 'bundles');
    await userEvent.click(await screen.findByRole('button', {name: 'More for #1'}));
    expect(screen.getByRole('menuitem', {name: 'Archived'})).toHaveAttribute('aria-disabled', 'true');
});

it('saves the tags of a bundle as a clean comma separated list', async () => {
    const store = {list: [bundle(1)]};
    const api = master(store, {'PUT /demo/app/3/bundle/1': () => {
        store.list = [bundle(1, {tags: 'rc,beta'})];
    }});
    renderBundles(api, 'bundles');
    await userEvent.click(await screen.findByRole('button', {name: 'Add tags to #1'}));
    await userEvent.type(screen.getByRole('textbox', {name: 'Tags'}), ' rc, beta , rc');
    await userEvent.click(screen.getByRole('button', {name: 'Save'}));
    expect(await screen.findByText('beta')).toBeInTheDocument();
    expect(api.calls.find(c => c.method === 'PUT')?.data).toEqual({tags: 'rc,beta'});
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
});

it('edits the tags a bundle has', async () => {
    renderBundles(master({list: [bundle(1, {tags: 'rc'})]}), 'bundles');
    await userEvent.click(await screen.findByRole('button', {name: 'Edit tags of #1'}));
    expect(screen.getByRole('textbox', {name: 'Tags'})).toHaveValue('rc');
    await userEvent.click(screen.getByRole('button', {name: 'Cancel'}));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
});
