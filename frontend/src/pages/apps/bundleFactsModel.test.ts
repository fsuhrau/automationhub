import {expect, it} from 'vitest';
import {FactSpec} from '../../types/bundle.fact';
import {androidFacts, appleFacts, fileFacts} from './bundleFactsModel';
import {androidInfo, appleInfo, bundleInfo, detail} from './appsFixtures';

const find = (facts: FactSpec[], label: string): FactSpec | undefined => facts.find(f => f.label === label);

it('lists what the manifest says, naming the api levels and densities', () => {
    const facts = androidFacts(androidInfo({maxSdk: '33', screens: ['normal', 'large']}));
    expect(find(facts, 'Package')).toMatchObject({text: 'com.demo.quest', copy: true, monospace: true});
    expect(find(facts, 'Min SDK')?.text).toBe('24 (Android 7.0)');
    expect(find(facts, 'Max SDK')?.text).toBe('33 (Android 13)');
    expect(find(facts, 'Densities')?.text).toBe('any, 480');
    expect(find(facts, 'Screens')?.text).toBe('normal, large');
    expect(find(facts, 'Locales (2)')).toMatchObject({chips: ['en', 'de'], max: 24});
    expect(find(facts, 'Flags')?.chips).toEqual(['debuggable']);
});

it('leaves out what the manifest does not say', () => {
    const facts = androidFacts(androidInfo({maxSdk: undefined, compileSdk: undefined, locales: undefined, screens: undefined, densities: undefined}));
    expect(find(facts, 'Max SDK')?.text).toBeUndefined();
    expect(find(facts, 'Compile SDK')?.text).toBeUndefined();
    expect(find(facts, 'Locales (0)')?.chips).toBeUndefined();
    expect(find(facts, 'Densities')?.text).toBeUndefined();
});

it('flags a test only or android tv build', () => {
    const facts = androidFacts(androidInfo({debuggable: false, testOnly: true, leanback: true}));
    expect(find(facts, 'Flags')?.chips).toEqual(['test only', 'Android TV']);
});

it('lists what the Info.plist says', () => {
    const info = bundleInfo({identifier: 'com.demo.ios', version: '2.0', build: '7'});
    const facts = appleFacts(appleInfo(), info);
    expect(find(facts, 'Bundle identifier')).toMatchObject({text: 'com.demo.ios', copy: true});
    expect(find(facts, 'Xcode')?.text).toBe('15.0 (15A)');
    expect(find(facts, 'Orientations')?.text).toBe('Landscape left');
    expect(find(facts, 'File sharing')?.text).toBe('on');
    expect(find(facts, 'Frameworks (1)')).toMatchObject({chips: ['UIKit'], max: 16});
    expect(find(facts, 'Extensions (0)')).toBeDefined();
});

it('leaves out the xcode build and file sharing when absent', () => {
    const facts = appleFacts(appleInfo({xcodeBuild: undefined, fileSharing: false, xcode: undefined, extensions: undefined}), bundleInfo());
    expect(find(facts, 'Xcode')?.text).toBeUndefined();
    expect(find(facts, 'File sharing')?.text).toBeUndefined();
    expect(find(facts, 'Extensions (0)')).toBeDefined();
    expect(find(appleFacts(appleInfo({xcodeBuild: undefined}), bundleInfo()), 'Xcode')?.text).toBe('15.0');
});

it('describes the file of a bundle', () => {
    const facts = fileFacts(detail(1), Date.parse('2001-02-04T04:05:06Z'));
    expect(find(facts, 'File name')?.text).toBe('quest-1_hash1.apk');
    expect(find(facts, 'Size')?.text).toBe('4.77MB');
    expect(find(facts, 'Installed size')?.text).toBe('4.77MB');
    expect(find(facts, 'Uploaded')?.text).toMatch(/2001\/02\/03 .* \(1 day ago\)/);
    expect(find(facts, 'Details read')?.text).toBeDefined();
    expect(find(facts, 'Engine')?.text).toBe('Unity · 2022.3 · IL2CPP');
    expect(find(facts, 'Launch activity')?.text).toBeUndefined();
});

it('falls back to what the master stored when the details were not read', () => {
    const old = detail(1, {info: undefined, size: 0, android: {LaunchActivity: 'Main'}, executable: {Executable: 'game'}});
    const facts = fileFacts(old, 0);
    expect(find(facts, 'Launch activity')?.text).toBe('Main');
    expect(find(facts, 'Executable')?.text).toBe('game');
    expect(find(facts, 'Size')?.text).toBeUndefined();
    expect(find(facts, 'Installed size')?.text).toBeUndefined();
    expect(find(facts, 'Details read')?.text).toBeUndefined();
});
