import {expect, it} from 'vitest';
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IAppBundleDetail} from '../../types/app';
import {androidInfo, bundleInfo, detail, device} from './appsFixtures';
import {failure, fakeApi, FakeApi, Handler, renderBundles} from './appsHarness';

const DETAIL = 'GET /demo/app/3/bundle/7';

const master = (shown: () => IAppBundleDetail, extra: Record<string, Handler> = {}): FakeApi => fakeApi({
    [DETAIL]: shown,
    'GET /demo/devices': () => [device(1), device(2, {osVersion: '6', alias: 'Old phone'}), device(3, {platformType: 1 - 1})],
    ...extra,
});

const open = async (api: FakeApi): Promise<void> => {
    renderBundles(api, 'bundle/7');
    await screen.findByRole('heading', {level: 1});
};

it('shows what the master read from the bundle in its sections', async () => {
    await open(master(() => detail(7, {tags: 'rc'})));
    expect(screen.getByRole('heading', {level: 1})).toHaveTextContent('Quest');
    expect(screen.getByText('Version 1.2.0 (build 12) · #7')).toBeInTheDocument();
    for (const title of ['File', 'Devices', 'Android', 'Permissions (2)', 'Signing', 'Contents', 'Test runs (2)']) {
        expect(screen.getByRole('heading', {name: title})).toBeInTheDocument();
    }
    expect(screen.getAllByText('debuggable')).toHaveLength(2);
    for (const badge of ['.apk', 'debug signed', 'rc']) {
        expect(screen.getByText(badge)).toBeInTheDocument();
    }
    expect(screen.getAllByText('Unity · 2022.3 · IL2CPP')).toHaveLength(2);
    expect(screen.getByText('Built without a keystore')).toBeInTheDocument();
});

it('lists the manifest of an android bundle', async () => {
    await open(master(() => detail(7)));
    expect(screen.getByText('24 (Android 7.0)')).toBeInTheDocument();
    expect(screen.getByText('any, 480')).toBeInTheDocument();
    expect(screen.getByText('Locales (2)')).toBeInTheDocument();
    expect(screen.getByText('arm64-v8a')).toBeInTheDocument();
    expect(screen.getByText('CAMERA')).toBeInTheDocument();
    expect(screen.getByText('1 need the player\'s consent at runtime')).toBeInTheDocument();
    expect(screen.getByText('up to API 30')).toBeInTheDocument();
    expect(screen.getByText('camera')).toBeInTheDocument();
    expect(screen.getByText('optional')).toBeInTheDocument();
});

it('shows the size by part and the largest files', async () => {
    await open(master(() => detail(7)));
    expect(within(screen.getByRole('list', {name: 'Size by part'})).getAllByRole('listitem')).toHaveLength(2);
    expect(within(screen.getByRole('table', {name: 'Largest files'})).getByText('lib/libunity.so')).toBeInTheDocument();
    expect(screen.getByText('1,200 files · 4.77MB installed · 2.86MB compressed')).toBeInTheDocument();
});

it('says which devices of the pool can install the bundle and why not', async () => {
    await open(master(() => detail(7)));
    expect(await screen.findAllByLabelText('can install')).toHaveLength(1);
    expect(screen.getAllByLabelText('can\'t install')).toHaveLength(1);
    expect(screen.getByText('needs Android 7.0, device has 6')).toBeInTheDocument();
    expect(screen.getByRole('link', {name: 'Old phone'})).toHaveAttribute('href', '/project/demo/device/2');
});

it('says when the pool has no device or the devices do not load', async () => {
    const api = master(() => detail(7), {'GET /demo/devices': () => []});
    await open(api);
    expect(await screen.findByText('The pool has no device for this platform.')).toBeInTheDocument();
});

it('shows no devices when they fail to load', async () => {
    await open(master(() => detail(7), {'GET /demo/devices': () => {
        throw failure('no devices');
    }}));
    expect(await screen.findByText('The pool has no device for this platform.')).toBeInTheDocument();
});

it('copies the ids of a bundle with the copy buttons', async () => {
    const user = userEvent.setup();
    await open(master(() => detail(7)));
    await user.click(screen.getByRole('button', {name: 'Copy Package'}));
    expect(await navigator.clipboard.readText()).toBe('com.demo.quest');
    await user.click(screen.getAllByRole('button', {name: 'Copy SHA-1'})[0]);
    expect(await navigator.clipboard.readText()).toBe(detail(7).hash);
});

it('lists the test runs that used the bundle with their results', async () => {
    await open(master(() => detail(7)));
    const table = screen.getByRole('table', {name: 'Test runs with this bundle'});
    expect(within(table).getByRole('link', {name: '#40'})).toHaveAttribute('href', '/project/demo/app:3/test/7/run/40');
    expect(within(table).getByText('Passed')).toBeInTheDocument();
    expect(within(table).getByText('3 passed, 1 failed, 1 unstable')).toBeInTheDocument();
    expect(within(table).getByText('Test 8')).toBeInTheDocument();
    expect(within(table).getByText('weird')).toBeInTheDocument();
});

it('says when the list of runs is cut and when no test ran', async () => {
    await open(master(() => detail(7, {runCount: 9})));
    expect(screen.getByText('The latest 2')).toBeInTheDocument();
});

it('says that no test ran with the bundle', async () => {
    await open(master(() => detail(7, {runs: [], runCount: 0})));
    expect(screen.getByText('No test ran with this bundle yet.')).toBeInTheDocument();
});

it('links to the list and to the comparison', async () => {
    await open(master(() => detail(7)));
    expect(screen.getByRole('link', {name: 'All bundles'})).toHaveAttribute('href', '/project/demo/app:3/bundles');
    expect(screen.getByRole('link', {name: 'Compare'})).toHaveAttribute('href', '/project/demo/app:3/bundles/compare?b=7');
    expect(screen.getByRole('link', {name: 'Download'})).toHaveAttribute('href', '/upload/quest-7_hash7.apk');
});

it('shows the android details of a bundle without the details of an old master', async () => {
    const old = detail(7, {info: undefined, android: {LaunchActivity: 'com.old.Main'}});
    await open(master(() => old));
    expect(screen.getByText('com.old.Main')).toBeInTheDocument();
    expect(screen.queryByRole('heading', {name: 'Devices'})).not.toBeInTheDocument();
});

it('shows the flags of the manifest', async () => {
    await open(master(() => detail(7, {info: bundleInfo({android: androidInfo({testOnly: true, leanback: true, maxSdk: '33'})})})));
    expect(screen.getByText('test only')).toBeInTheDocument();
    expect(screen.getByText('Android TV')).toBeInTheDocument();
    expect(screen.getByText('33 (Android 13)')).toBeInTheDocument();
});

it('shows no signature for an unsigned bundle', async () => {
    await open(master(() => detail(7, {info: bundleInfo({signing: undefined})})));
    expect(screen.getByText('The bundle is not signed; devices won\'t install it.')).toBeInTheDocument();
    expect(screen.getByText('No signature found')).toBeInTheDocument();
});

it('says that no permission is requested', async () => {
    await open(master(() => detail(7, {info: bundleInfo({android: androidInfo({permissions: [], features: []})})})));
    expect(screen.getByText('None requested.')).toBeInTheDocument();
});

it('waits for the answer of the master', async () => {
    const api = master(() => detail(7));
    renderBundles(api, 'bundle/x');
    expect(await screen.findByRole('progressbar')).toBeInTheDocument();
    expect(api.calls.some(c => c.url.includes('/bundle/'))).toBe(false);
});

it('falls back to the platform icon when the icon does not load', async () => {
    await open(master(() => detail(7)));
    const icon = screen.getByAltText('App icon');
    expect(icon).toHaveAttribute('src', '/api/demo/app/3/bundle/7/icon');
    icon.dispatchEvent(new Event('error'));
    await waitFor(() => expect(screen.queryByAltText('App icon')).not.toBeInTheDocument());
});

it('tells what a part of the size is on hover', async () => {
    await open(master(() => detail(7)));
    await userEvent.hover(within(screen.getByRole('list', {name: 'Size by part'})).getAllByRole('listitem')[0]);
    expect(await screen.findByText(/3 files · 80\.0 %/)).toBeInTheDocument();
});
