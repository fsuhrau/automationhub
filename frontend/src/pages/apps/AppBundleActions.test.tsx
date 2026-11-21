import {expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IAppBundleDetail} from '../../types/app';
import {bundle, detail} from './appsFixtures';
import {failure, fakeApi, FakeApi, Handler, renderBundles} from './appsHarness';

const BUNDLE = '/demo/app/3/bundle/7';

// a master that keeps the bundle, so a bundle loaded after a change shows it
const master = (store: { shown: IAppBundleDetail }, extra: Record<string, Handler> = {}): FakeApi => fakeApi({
    [`GET ${BUNDLE}`]: () => store.shown,
    'GET /demo/devices': () => [],
    'GET /demo/app/3/bundles': () => [bundle(7)],
    ...extra,
});

const open = async (api: FakeApi): Promise<void> => {
    renderBundles(api, 'bundle/7');
    await screen.findByRole('heading', {level: 1});
};

const confirmWith = async (name: string): Promise<void> => {
    await userEvent.click(within(await screen.findByRole('dialog')).getByRole('button', {name}));
};

it('reads the file again and shows the new details', async () => {
    const store = {shown: detail(7, {info: undefined})};
    await open(master(store, {[`POST ${BUNDLE}/read`]: () => {
        store.shown = detail(7);
    }}));
    await userEvent.click(screen.getByRole('button', {name: 'Read again'}));
    expect(await screen.findByRole('heading', {name: 'Android'})).toBeInTheDocument();
});

it('shows that the file is being read', async () => {
    let finish!: () => void;
    const store = {shown: detail(7)};
    await open(master(store, {[`POST ${BUNDLE}/read`]: () => new Promise<void>(resolve => {
        finish = resolve;
    })}));
    await userEvent.click(screen.getByRole('button', {name: 'Read again'}));
    expect(await screen.findByRole('button', {name: 'Reading…'})).toBeDisabled();
    finish();
    expect(await screen.findByRole('button', {name: 'Read again'})).toBeEnabled();
});

it('shows why reading the file failed', async () => {
    await open(master({shown: detail(7)}, {[`POST ${BUNDLE}/read`]: () => {
        throw failure('file is gone');
    }}));
    await userEvent.click(screen.getByRole('button', {name: 'Read again'}));
    expect(await screen.findByText('file is gone')).toBeInTheDocument();
});

it('asks before archiving and shows the bundle as archived', async () => {
    const store = {shown: detail(7)};
    await open(master(store, {[`POST ${BUNDLE}/archive`]: () => {
        store.shown = detail(7, {archivedAt: '2020-01-01T00:00:00Z'});
    }}));
    await userEvent.click(screen.getByRole('button', {name: 'Archive bundle'}));
    expect(screen.getByText(/Archive version 1.2.0/)).toBeInTheDocument();
    await confirmWith('Archive bundle');
    expect(await screen.findByText(/The file is gone; the details and test runs stay/)).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: 'Archive bundle'})).not.toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Read again'})).toBeDisabled();
});

it('keeps the bundle when the archive question is cancelled', async () => {
    await open(master({shown: detail(7)}));
    await userEvent.click(screen.getByRole('button', {name: 'Archive bundle'}));
    await confirmWith('Cancel');
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
});

it('deletes the bundle after the question and goes back to the list', async () => {
    const api = master({shown: detail(7)}, {[`DELETE ${BUNDLE}`]: () => undefined});
    await open(api);
    await userEvent.click(screen.getByRole('button', {name: 'Delete bundle'}));
    await confirmWith('Delete bundle');
    expect(await screen.findByRole('heading', {name: 'Bundles'})).toBeInTheDocument();
    expect(api.calls.some(c => c.method === 'DELETE')).toBe(true);
});

it('stays on the bundle when deleting fails', async () => {
    await open(master({shown: detail(7)}, {[`DELETE ${BUNDLE}`]: () => {
        throw failure('still in use');
    }}));
    await userEvent.click(screen.getByRole('button', {name: 'Delete bundle'}));
    await confirmWith('Delete bundle');
    expect(await screen.findByText('still in use')).toBeInTheDocument();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
});

it('tells about a bundle whose details were never read and offers to read them', async () => {
    const store = {shown: detail(7, {info: undefined})};
    await open(master(store, {[`POST ${BUNDLE}/read`]: () => undefined}));
    expect(screen.getByText(/uploaded before the master read all details from the file. Read it again/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', {name: 'Read now'}));
    await waitFor(() => expect(screen.getByRole('button', {name: 'Read now'})).toBeEnabled());
});

it('tells about an archived bundle whose details were never read', async () => {
    await open(master({shown: detail(7, {info: undefined, archivedAt: '2020-01-01T00:00:00Z'})}));
    expect(screen.getByText(/and its file is archived/)).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: 'Read now'})).not.toBeInTheDocument();
});

it('shows why the bundle did not load and tries again', async () => {
    let fail = true;
    const api = master({shown: detail(7)}, {[`GET ${BUNDLE}`]: () => {
        if (fail) throw failure('master is down');
        return detail(7);
    }});
    renderBundles(api, 'bundle/7');
    expect(await screen.findByText('master is down')).toBeInTheDocument();
    expect(screen.getByRole('link', {name: 'All bundles'})).toHaveAttribute('href', '/project/demo/app:3/bundles');
    fail = false;
    await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
    expect(await screen.findByRole('heading', {level: 1})).toBeInTheDocument();
});
