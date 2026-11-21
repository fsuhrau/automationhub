import {expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IAppBinaryData} from '../../types/app';
import {bundle} from './appsFixtures';
import {failure, fakeApi, FakeApi, renderBundles} from './appsHarness';

// a master with a list of bundles that changes like the real one
const master = (bundles: IAppBinaryData[], extra: Parameters<typeof fakeApi>[0] = {}): FakeApi =>
    fakeApi({'GET /demo/app/3/bundles': () => bundles, ...extra});

const openMenu = async (id: number): Promise<void> => {
    await userEvent.click(await screen.findByRole('button', {name: `More for #${id}`}));
};

it('lists the bundles newest first with their details', async () => {
    renderBundles(master([bundle(1, {tags: 'rc'}), bundle(2, {archivedAt: '2002-01-01T00:00:00Z', runCount: undefined, size: 0})]), 'bundles');
    const rows = await screen.findAllByRole('row');
    expect(rows).toHaveLength(3);
    expect(within(rows[1]).getByText('1.2.0')).toBeInTheDocument();
    expect(within(rows[1]).getByText('newest')).toBeInTheDocument();
    expect(within(rows[1]).getByText('archived')).toBeInTheDocument();
    expect(within(rows[2]).getByText('rc')).toBeInTheDocument();
    expect(screen.getByText('2 bundles')).toBeInTheDocument();
    expect(screen.getByRole('link', {name: 'Compare'})).toHaveAttribute('href', '/project/demo/app:3/bundles/compare');
});

it('links a bundle to its details', async () => {
    renderBundles(master([bundle(5)]), 'bundles');
    expect(await screen.findByRole('link', {name: '1.5.0'})).toHaveAttribute('href', '/project/demo/app:3/bundle/5');
    expect(screen.queryByRole('link', {name: 'Compare'})).not.toBeInTheDocument();
    expect(screen.getByText('1 bundle')).toBeInTheDocument();
});

it('shows the engine and build of a bundle read from its file', async () => {
    const info = {format: 'apk', build: '99', engine: {name: 'Unity'}, readAt: ''};
    renderBundles(master([bundle(1, {info, version: '1.0'})]), 'bundles');
    expect(await screen.findByText('build 99')).toBeInTheDocument();
    expect(screen.getByText(/· Unity/)).toBeInTheDocument();
});

it('invites to upload when there are no bundles', async () => {
    renderBundles(master([]), 'bundles');
    expect(await screen.findByText('No bundles yet')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Choose files'})).toBeInTheDocument();
});

it('shows why the list failed and loads it again on retry', async () => {
    let fail = true;
    const api = fakeApi({'GET /demo/app/3/bundles': () => {
        if (fail) throw failure('master is down');
        return [bundle(1)];
    }});
    renderBundles(api, 'bundles');
    expect(await screen.findByText('master is down')).toBeInTheDocument();
    fail = false;
    await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
    expect(await screen.findByText('1.1.0')).toBeInTheDocument();
});

it('finds bundles by text and by clicking a tag', async () => {
    renderBundles(master([bundle(1, {tags: 'nightly'}), bundle(2)]), 'bundles');
    await userEvent.type(await screen.findByRole('textbox', {name: 'Find a bundle'}), 'zzz');
    expect(screen.getByText('No bundle matches.')).toBeInTheDocument();
    await userEvent.clear(screen.getByRole('textbox', {name: 'Find a bundle'}));
    await userEvent.click(screen.getByText('nightly'));
    expect(screen.getByRole('textbox', {name: 'Find a bundle'})).toHaveValue('nightly');
    expect(screen.getByText('1 of 2')).toBeInTheDocument();
});

it('copies the hash of a bundle from its menu', async () => {
    renderBundles(master([bundle(1)]), 'bundles');
    const user = userEvent.setup();
    await user.click(await screen.findByRole('button', {name: 'More for #1'}));
    await user.click(screen.getByRole('menuitem', {name: /Copy hash/}));
    expect(await navigator.clipboard.readText()).toBe(bundle(1).hash);
    await waitFor(() => expect(screen.queryByRole('menu')).not.toBeInTheDocument());
});

it('links the menu to the details and to the comparison with the previous bundle', async () => {
    renderBundles(master([bundle(1), bundle(2)]), 'bundles');
    await openMenu(2);
    expect(screen.getByRole('menuitem', {name: 'Details'})).toHaveAttribute('href', '/project/demo/app:3/bundle/2');
    expect(screen.getByRole('menuitem', {name: /Compare with previous/})).toHaveAttribute('href', '/project/demo/app:3/bundles/compare?b=2');
});

it('can not compare the oldest bundle with a previous one', async () => {
    renderBundles(master([bundle(1), bundle(2)]), 'bundles');
    await openMenu(1);
    expect(screen.getByRole('menuitem', {name: /Compare with previous/})).toHaveAttribute('aria-disabled', 'true');
});
