import {expect, it} from 'vitest';
import {screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IAppBinaryData, IAppBundleDetail} from '../../types/app';
import {androidInfo, bundle, bundleInfo, detail} from './appsFixtures';
import {failure, fakeApi, FakeApi} from '../../testing/fakeApi';
import {renderBundles} from './appsHarness';

const older = detail(1, {info: bundleInfo({
    version: '1.0', android: androidInfo({permissions: [{name: 'android.permission.CAMERA'}], versionName: '1.0'}),
    contents: {
        files: 1000, size: 4000000, compressed: 2000000,
        groups: [{path: 'lib', files: 3, size: 3000000, compressed: 2000000}, {path: 'old', files: 1, size: 10, compressed: 10}],
        largest: [{path: 'lib/libunity.so', size: 2000000, compressed: 1000000}, {path: 'gone.bin', size: 5, compressed: 5}],
    },
})});

const newer = detail(2, {tags: 'rc'});

const master = (details: IAppBundleDetail[], list?: IAppBinaryData[]): FakeApi => fakeApi({
    'GET /demo/app/3/bundles': () => list ?? details,
    ...Object.fromEntries(details.map(d => [`GET /demo/app/3/bundle/${d.id}`, () => d])),
});

const open = async (api: FakeApi, query = ''): Promise<void> => {
    renderBundles(api, `bundles/compare${query}`);
    await screen.findByText(/attributes? changed|The same bundle|No attribute changed/);
};

it('compares the newest bundle with the one before it and shows what changed', async () => {
    await open(master([older, newer]));
    expect(screen.getByText(/\d+ attributes changed/)).toBeInTheDocument();
    const app = screen.getByRole('table', {name: 'App'});
    expect(within(app).getByText('Version')).toBeInTheDocument();
    expect(within(app).getByText('1.0')).toBeInTheDocument();
    expect(within(app).getByText('1.2.0')).toBeInTheDocument();
    expect(within(app).queryByText('Name')).not.toBeInTheDocument();
});

it('shows the added permission and the size that grew', async () => {
    await open(master([older, newer]));
    const android = screen.getByRole('table', {name: 'Android'});
    expect(within(android).getByText('+ INTERNET (up to API 30)')).toBeInTheDocument();
    const size = screen.getByRole('table', {name: 'Size'});
    expect(within(size).getAllByText(/\+976\.56KB/).length).toBeGreaterThan(0);
    expect(within(size).getByText(/^\+200 /)).toBeInTheDocument();
});

it('shows where the size moved by part and by file', async () => {
    await open(master([older, newer]));
    const parts = screen.getByRole('table', {name: 'Size by part'});
    expect(within(parts).getByText('lib')).toBeInTheDocument();
    expect(within(parts).getByText('assets')).toBeInTheDocument();
    expect(within(parts).getAllByText('not there').length).toBeGreaterThan(0);
    const files = screen.getByRole('table', {name: 'Largest files'});
    expect(within(files).getAllByText('unknown')).toHaveLength(1);
    expect(within(files).getAllByText('not among largest').length).toBeGreaterThan(0);
});

it('shows the heads of both bundles', async () => {
    await open(master([older, newer]));
    expect(screen.getByRole('link', {name: '1.0 (12) · #1'})).toHaveAttribute('href', '/project/demo/app:3/bundle/1');
    expect(screen.getByRole('link', {name: '1.2.0 (12) · #2'})).toHaveAttribute('href', '/project/demo/app:3/bundle/2');
    expect(screen.getAllByText(/· 2 runs/)).toHaveLength(2);
});

it('shows every attribute when only changes is switched off', async () => {
    await open(master([older, newer]));
    await userEvent.click(screen.getByRole('switch', {name: 'Only changes'}));
    const app = screen.getByRole('table', {name: 'App'});
    expect(within(app).getByText('Name')).toBeInTheDocument();
    expect(within(app).getAllByText('same').length).toBeGreaterThan(0);
    expect(within(screen.getByRole('table', {name: 'Android'})).getAllByText(/the same in both|unchanged/).length).toBeGreaterThan(0);
});

it('compares the bundles the url picks and says when both sides are the same', async () => {
    await open(master([older, newer]), '?a=2&b=2');
    expect(screen.getByText('The same bundle on both sides.')).toBeInTheDocument();
});

it('points out that the from side is the newer bundle', async () => {
    await open(master([older, newer]), '?a=2&b=1');
    expect(screen.getByText('From is the newer bundle')).toBeInTheDocument();
});

it('says when no attribute changed between two identical bundles', async () => {
    await open(master([detail(1), detail(2, {size: 5000000})]), '?a=1&b=2');
    expect(screen.getByText('No attribute changed.')).toBeInTheDocument();
    expect(screen.getAllByText('No size changed.')).toHaveLength(2);
});

it('says when the from side is picked and the to side is the newest', async () => {
    await open(master([older, newer]), '?a=1');
    expect(screen.getByRole('link', {name: /#2/})).toBeInTheDocument();
});

it('swaps the two sides', async () => {
    await open(master([older, newer]));
    await userEvent.click(screen.getByRole('button', {name: 'Swap the bundles'}));
    expect(await screen.findByText('From is the newer bundle')).toBeInTheDocument();
});

it('picks another bundle for a side', async () => {
    const third = detail(3, {info: bundleInfo({version: '3.0'})});
    await open(master([older, newer, third]));
    await userEvent.click(screen.getByRole('combobox', {name: 'From'}));
    await userEvent.click(within(await screen.findByRole('listbox')).getByRole('option', {name: /#1/}));
    expect(await screen.findByRole('link', {name: '1.0 (12) · #1'})).toBeInTheDocument();
    expect(screen.getByRole('link', {name: '3.0 (12) · #3'})).toBeInTheDocument();
});

it('marks tags and archived bundles in the pickers', async () => {
    const gone = bundle(1, {archivedAt: '2020-01-01T00:00:00Z'});
    await open(master([detail(1, {archivedAt: gone.archivedAt}), newer], [bundle(2, {tags: 'rc'}), gone]));
    await userEvent.click(screen.getByRole('combobox', {name: 'From'}));
    const options = within(await screen.findByRole('listbox'));
    expect(options.getByText('rc')).toBeInTheDocument();
    expect(options.getByLabelText('archived')).toBeInTheDocument();
});

it('shows an archived bundle and one without details in the heads', async () => {
    await open(master([detail(1, {info: undefined, archivedAt: '2020-01-01T00:00:00Z'}), newer]));
    expect(screen.getByText('archived')).toBeInTheDocument();
    expect(screen.getByText(/only the basics can be compared/)).toBeInTheDocument();
});

it('needs two bundles to compare', async () => {
    renderBundles(master([newer]), 'bundles/compare');
    expect(await screen.findByText('The app needs two bundles to compare.')).toBeInTheDocument();
});

it('shows why the bundles did not load', async () => {
    renderBundles(fakeApi({'GET /demo/app/3/bundles': () => {
        throw failure('master is down');
    }}), 'bundles/compare');
    expect(await screen.findByText('master is down')).toBeInTheDocument();
});

it('shows why a bundle did not load', async () => {
    const api = fakeApi({
        'GET /demo/app/3/bundles': () => [bundle(1), bundle(2)],
        'GET /demo/app/3/bundle/1': () => detail(1),
        'GET /demo/app/3/bundle/2': () => {
            throw failure('bundle 2 is broken');
        },
    });
    renderBundles(api, 'bundles/compare');
    expect(await screen.findByText('bundle 2 is broken')).toBeInTheDocument();
});

it('links back to the list of bundles', async () => {
    await open(master([older, newer]));
    expect(screen.getByRole('link', {name: 'All bundles'})).toHaveAttribute('href', '/project/demo/app:3/bundles');
});
