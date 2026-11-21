import {expect, it} from 'vitest';
import {screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IAppData} from '../../types/app';
import IDeviceData from '../../types/device';
import {PlatformType} from '../../types/platform.type.enum';
import {androidApp, appleInfo, bundleInfo, detail, device, iosDetail, profile} from './appsFixtures';
import {fakeApi, FakeApi, renderBundles} from './appsHarness';

const iosApp = {...androidApp, platform: PlatformType.iOS} as IAppData;
const DAY = 86400000;

const phone = (id: number, udid: string, over = {}): IDeviceData =>
    device(id, {platformType: PlatformType.iOS, osVersion: '17.0', deviceIdentifier: udid, ...over});

const master = (shown = iosDetail(7)): FakeApi => fakeApi({
    'GET /demo/app/3/bundle/7': () => shown,
    'GET /demo/devices': () => [phone(1, 'udid-1', {alias: 'My iPhone'}), phone(2, 'other', {osVersion: '14.0'})],
});

const open = async (api: FakeApi): Promise<void> => {
    renderBundles(api, 'bundle/7', iosApp);
    await screen.findByRole('heading', {level: 1});
};

it('shows the Info.plist, permissions and profile of an ios bundle', async () => {
    await open(master());
    for (const title of ['iOS', 'Permissions (1)', 'Provisioning profile']) {
        expect(screen.getByRole('heading', {name: title})).toBeInTheDocument();
    }
    expect(screen.getByText('Camera')).toBeInTheDocument();
    expect(screen.getByText('Scan codes')).toBeInTheDocument();
    expect(screen.getByText('Landscape left')).toBeInTheDocument();
    expect(screen.getByText('15.0 (15A)')).toBeInTheDocument();
    expect(screen.getAllByText('Development').length).toBeGreaterThan(0);
});

it('names a mac bundle after its platform', async () => {
    await open(master(detail(7, {info: bundleInfo({android: undefined, apple: appleInfo({platform: 'macosx'})})})));
    expect(screen.getByRole('heading', {name: 'macOS'})).toBeInTheDocument();
});

it('says no permission is asked when the app has no usage descriptions', async () => {
    await open(master(detail(7, {info: bundleInfo({android: undefined, apple: appleInfo({usageDescriptions: []})})})));
    expect(screen.getByText('The app asks for no protected access.')).toBeInTheDocument();
});

it('shows the devices of the profile that are in the pool and all of them on request', async () => {
    await open(master());
    expect(await screen.findAllByText('My iPhone')).toHaveLength(2);
    expect(screen.getByText('2 devices, 1 of them in the device pool')).toBeInTheDocument();
    expect(screen.queryByText('UDID-2')).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', {name: 'Show all 2 devices'}));
    expect(screen.getByText('UDID-2')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', {name: 'Show only devices in the pool'}));
    expect(screen.queryByText('UDID-2')).not.toBeInTheDocument();
});

it('explains why an iphone can not install the bundle', async () => {
    await open(master());
    expect(await screen.findByText('needs iOS 15.0, device has 14.0; not in the provisioning profile')).toBeInTheDocument();
});

it('shows the entitlements of the profile on request', async () => {
    await open(master());
    await userEvent.click(screen.getByRole('button', {name: 'Show 2 entitlements'}));
    expect(screen.getByText('aps-environment')).toBeInTheDocument();
    expect(screen.getByText('x, y')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', {name: 'Hide entitlements'}));
    expect(screen.queryByText('aps-environment')).not.toBeInTheDocument();
});

it('shows a profile for every device without a list of them', async () => {
    const all = {...profile, allDevices: true, devices: undefined, entitlements: undefined};
    await open(master(detail(7, {info: bundleInfo({android: undefined, apple: appleInfo({profile: all})})})));
    expect(screen.getByText('every device (enterprise)')).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: /entitlements/})).not.toBeInTheDocument();
});

it('shows a profile without devices', async () => {
    const none = {...profile, devices: undefined, teamId: undefined};
    await open(master(detail(7, {info: bundleInfo({android: undefined, apple: appleInfo({profile: none})})})));
    expect(screen.getByText('none (installs from the App Store or TestFlight)')).toBeInTheDocument();
});

it('warns when the profile has expired or is about to', async () => {
    const expired = {...profile, expiresAt: new Date(Date.now() - DAY).toISOString()};
    await open(master(detail(7, {info: bundleInfo({android: undefined, apple: appleInfo({profile: expired})})})));
    expect(screen.getByText('expired')).toBeInTheDocument();
});

it('counts the days left of a profile that is about to expire', async () => {
    const soon = {...profile, expiresAt: new Date(Date.now() + 10 * DAY).toISOString()};
    await open(master(detail(7, {info: bundleInfo({android: undefined, apple: appleInfo({profile: soon})})})));
    expect(screen.getByText(/^(9|10) days left$/)).toBeInTheDocument();
});

it('shows a certificate that signs with another one as its issuer', async () => {
    const signed = {certificates: [{
        subject: 'CN=Dev, O=A', issuer: 'CN=Apple CA, O=Apple', serial: '2', notBefore: '2001-01-01T00:00:00Z', notAfter: '2999-01-01T00:00:00Z',
        sha1: 'c'.repeat(40), sha256: 'd'.repeat(64), debug: false,
    }]};
    await open(master(detail(7, {info: bundleInfo({android: undefined, apple: appleInfo(), signing: signed})})));
    expect(screen.getByText('Certificates the provisioning profile allows to sign the app')).toBeInTheDocument();
    expect(screen.getByText('Apple CA')).toBeInTheDocument();
    expect(screen.queryByText('debug key')).not.toBeInTheDocument();
});
