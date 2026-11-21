import {expect, it} from 'vitest';
import {screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IAppBundleDetail} from '../../types/app';
import {androidInfo, appleInfo, bundleInfo, detail, profile} from './appsFixtures';
import {fakeApi} from '../../testing/fakeApi';
import {renderBundles} from './appsHarness';

type Bundle = IAppBundleDetail;

const many = (prefix: string, n: number): string[] => Array.from({length: n}, (_, i) => `${prefix}${i}`);

const compare = async (a: Bundle, b: Bundle): Promise<void> => {
    renderBundles(fakeApi({
        'GET /demo/app/3/bundles': () => [a, b],
        [`GET /demo/app/3/bundle/${a.id}`]: () => a,
        [`GET /demo/app/3/bundle/${b.id}`]: () => b,
    }), 'bundles/compare');
    await screen.findByText(/attributes? changed|No attribute/);
};

it('shows a long list of changes in part and all of it on request', async () => {
    const locales = (list: string[]): Bundle => detail(list.length, {info: bundleInfo({android: androidInfo({locales: list})})});
    await compare(locales(many('a', 1)), locales(many('b', 35)));
    const android = screen.getByRole('table', {name: 'Android'});
    expect(within(android).getAllByText(/^\+ b/)).toHaveLength(30);
    await userEvent.click(within(android).getByRole('button', {name: '+5 more'}));
    expect(within(android).getAllByText(/^\+ b/)).toHaveLength(35);
});

it('shows a long list of size changes in part and all of it on request', async () => {
    const groups = (n: number): Bundle => detail(n, {info: bundleInfo({contents: {
        files: 1, size: 1, compressed: 1, groups: many('g', n).map((path, i) => ({path, files: 1, size: 100 + i, compressed: 1})),
    }})});
    await compare(groups(1), groups(20));
    const parts = screen.getByRole('table', {name: 'Size by part'});
    expect(within(parts).getAllByRole('row')).toHaveLength(16);
    await userEvent.click(screen.getByRole('button', {name: /Show all/}));
    expect(within(parts).getAllByRole('row').length).toBeGreaterThan(16);
});

it('compares provisioning profiles and signing of ios bundles', async () => {
    const ios = (id: number, over = {}): Bundle => detail(id, {platform: 'iOS', info: bundleInfo({
        android: undefined, apple: appleInfo({profile: {...profile, ...over}}), engine: undefined,
        signing: {schemes: [], certificates: [{subject: 'CN=Dev', issuer: 'CN=CA', serial: '1', notBefore: '2001-01-01T00:00:00Z',
            notAfter: '2999-01-01T00:00:00Z', sha1: 'e'.repeat(40), sha256: 'f'.repeat(64), debug: id === 1}]},
    })});
    await compare(ios(1), ios(2, {name: 'New profile', allDevices: true, kind: 'ad-hoc'}));
    const prof = screen.getByRole('table', {name: 'Provisioning profile'});
    expect(within(prof).getByText('New profile')).toBeInTheDocument();
    expect(within(prof).getByText('every device')).toBeInTheDocument();
    expect(within(prof).getByText('Ad Hoc')).toBeInTheDocument();
    expect(within(screen.getByRole('table', {name: 'Signing'})).getAllByText(/Dev \(eeeeeeee…\)/).length).toBeGreaterThan(0);
});

it('names the section of a mac bundle', async () => {
    const mac = (id: number, minOs: string): Bundle => detail(id, {info: bundleInfo({
        android: undefined, apple: appleInfo({platform: 'macosx', minOs}), engine: undefined, signing: undefined,
    })});
    await compare(mac(1, '11'), mac(2, '12'));
    expect(screen.getByRole('table', {name: 'macOS'})).toBeInTheDocument();
});

it('compares a bundle that has details with one that has none', async () => {
    await compare(detail(1, {info: undefined}), detail(2));
    expect(screen.getByRole('table', {name: 'Android'})).toBeInTheDocument();
});
