import {expect, it} from 'vitest';
import {IAppBundleDetail} from '../../types/app';
import {androidInfo, appleInfo, bundleInfo, detail, profile} from './appsFixtures';
import {compareBundles, hideEmpty, onlyChanges, previousOf} from './bundleCompareModel';

const titles = (info: ReturnType<typeof compareBundles>): string[] => info.sections.map(s => s.title);

it('compares two android bundles in the sections they share', () => {
    const result = compareBundles(detail(1), detail(2, {version: '2.0', info: bundleInfo({version: '2.0'})}));
    expect(titles(result)).toEqual(['App', 'Size', 'Android', 'Signing', 'Warnings']);
    expect(result.changes).toBe(1);
    expect(result.parts).toHaveLength(2);
    expect(result.files).toHaveLength(1);
});

it('adds the apple and profile sections for ios bundles', () => {
    const ios = (id: number, kind: 'development' | 'ad-hoc'): IAppBundleDetail => detail(id, {info: bundleInfo({
        android: undefined, signing: undefined, apple: appleInfo({profile: {...profile, kind}}),
    })});
    const result = compareBundles(ios(1, 'development'), ios(2, 'ad-hoc'));
    expect(titles(result)).toEqual(['App', 'Size', 'iOS', 'Provisioning profile', 'Warnings']);
    expect(result.changes).toBeGreaterThan(0);
});

it('has no profile section when neither bundle has a profile', () => {
    const ios = detail(1, {info: bundleInfo({android: undefined, signing: undefined, apple: appleInfo()})});
    expect(titles(compareBundles(ios, ios))).toEqual(['App', 'Size', 'iOS', 'Warnings']);
});

it('compares bundles without details by what the master stored', () => {
    const result = compareBundles(detail(1, {info: undefined, version: '1'}), detail(2, {info: undefined, version: '2', tags: 'a,b'}));
    expect(titles(result)).toEqual(['App', 'Size', 'Warnings']);
    expect(result.sections[0].rows.find(r => r.label === 'Version')).toMatchObject({a: '1', b: '2', change: 'changed'});
    expect(result.sections[0].rows.find(r => r.label === 'Tags')).toMatchObject({added: ['a', 'b']});
    expect(result.parts).toEqual([]);
});

it('compares the signing of one bundle with none', () => {
    const result = compareBundles(detail(1), detail(2, {info: bundleInfo({signing: undefined, android: androidInfo()})}));
    const signing = result.sections.find(s => s.title === 'Signing');
    expect(signing?.rows.find(r => r.label === 'First certificate expiry')).toMatchObject({a: '2999-02-03', b: undefined});
});

it('keeps only the rows that changed, and only the sections with such rows', () => {
    const result = compareBundles(detail(1), detail(2, {info: bundleInfo({version: '9'})}));
    const changed = onlyChanges(result.sections);
    expect(changed.map(s => s.title)).toEqual(['App']);
    expect(changed[0].rows.map(r => r.label)).toEqual(['Version']);
});

it('drops the rows that neither bundle has', () => {
    const result = compareBundles(detail(1, {info: undefined}), detail(2, {info: undefined}));
    const shown = hideEmpty(result.sections);
    expect(shown.map(s => s.title)).toEqual(['App', 'Size']);
    expect(shown.flatMap(s => s.rows).some(r => r.kind === 'value' && r.a === undefined && r.b === undefined)).toBe(false);
    expect(hideEmpty(compareBundles(detail(1), detail(2)).sections).map(s => s.title)).toContain('Warnings');
});

it('finds the bundle uploaded before another one', () => {
    const list = [detail(5), detail(2), detail(3)];
    expect(previousOf(list, 5)?.id).toBe(3);
    expect(previousOf(list, 2)).toBeUndefined();
});
