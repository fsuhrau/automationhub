import {expect, it} from 'vitest';
import {BundleInfo} from '../../types/bundle.info';
import IDeviceData from '../../types/device';
import {PlatformType} from '../../types/platform.type.enum';
import {androidInfo, appleInfo, bundleInfo, detail, device, profile} from './appsFixtures';
import {
    androidVersionOf, apiLevel, badgesOf, barWidth, buildOf, bundleTitle, commonName, compareVersions, deviceFit, engineLabel, entitlementValue, expiry,
    fileName, isDangerous, orientationName, platformsOf, runSummary, shareOf, shortPermission, usageName,
} from './bundleDetailModel';

it('names an api level with its android version', () => {
    expect(apiLevel('24')).toBe('24 (Android 7.0)');
    expect(apiLevel('99')).toBe('99');
    expect(apiLevel(undefined)).toBe('–');
    expect(androidVersionOf('33')).toBe('13');
    expect(androidVersionOf('32')).toBe('12');
    expect(androidVersionOf(undefined)).toBeUndefined();
});

it('compares dotted versions numerically', () => {
    expect(compareVersions('10.0', '9.3')).toBeGreaterThan(0);
    expect(compareVersions('9', '9.0.0')).toBe(0);
    expect(compareVersions('1.2', '1.10')).toBeLessThan(0);
});

it('shortens permissions and spots the ones asked at runtime', () => {
    expect(shortPermission('android.permission.CAMERA')).toBe('CAMERA');
    expect(isDangerous('android.permission.CAMERA')).toBe(true);
    expect(isDangerous('android.permission.INTERNET')).toBe(false);
    expect(isDangerous('com.other.CAMERA')).toBe(false);
});

it('turns plist keys into words', () => {
    expect(usageName('NSPhotoLibraryUsageDescription')).toBe('Photo Library');
    expect(orientationName('UIInterfaceOrientationLandscapeLeft')).toBe('Landscape left');
});

it('says whether a date has passed or is near', () => {
    const now = Date.UTC(2024, 0, 1);
    expect(expiry('2023-12-31', now)).toBe('expired');
    expect(expiry('2024-01-15', now)).toBe('soon');
    expect(expiry('2024-06-01', now)).toBe('valid');
});

it('takes the common name of a subject', () => {
    expect(commonName('C=US,CN=My App,O=Org')).toBe('My App');
    expect(commonName('O=Org')).toBe('O=Org');
});

it('lists the platforms a bundle installs on', () => {
    expect(platformsOf(PlatformType.iOS)).toEqual([PlatformType.iOS, PlatformType.IOSSimulator]);
    expect(platformsOf(PlatformType.Android)).toEqual([PlatformType.Android]);
    expect(platformsOf(undefined)).toEqual([]);
});

it('shows the name a file was uploaded with and a line for what is not known', () => {
    expect(fileName('game_abc.apk', 'abc')).toBe('game.apk');
    expect(fileName('game.apk', '')).toBe('game.apk');
    expect(fileName('', 'abc')).toBe('–');
});

it('shows entitlements in one line', () => {
    expect(entitlementValue(['a', {b: 1}])).toBe('a, {"b":1}');
    expect(entitlementValue(true)).toBe('true');
    expect(entitlementValue(null)).toBe('null');
});

it('sizes the bars of the size view', () => {
    expect(shareOf(1, 4)).toBe(0.25);
    expect(shareOf(1, 0)).toBe(0);
    expect(barWidth(50, 100)).toBe(50);
    expect(barWidth(0, 100)).toBe(0.5);
});

it('says how an engine and a test run are shown', () => {
    expect(engineLabel({name: 'Unity', version: '6', backend: 'Mono'})).toBe('Unity · 6 · Mono');
    expect(engineLabel(undefined)).toBeUndefined();
    const run = {id: 1, testId: 1, testName: '', createdAt: '', status: '', passed: 2, failed: 0, unstable: 0, open: 0};
    expect(runSummary(run)).toBe('2 passed');
    expect(runSummary({...run, failed: 1, unstable: 3})).toBe('2 passed, 1 failed, 3 unstable');
    expect(runSummary({...run, passed: 0})).toBe('');
});

it('titles a bundle by display name, name, identifier or id', () => {
    expect(bundleTitle(detail(1))).toBe('Quest');
    expect(bundleTitle(detail(1, {info: undefined}))).toBe('quest-1.apk');
    expect(bundleTitle(detail(1, {info: undefined, name: ''}))).toBe('com.demo.quest');
    expect(bundleTitle(detail(1, {info: undefined, name: '', identifier: ''}))).toBe('Bundle #1');
});

it('shows the build only when it differs from the version', () => {
    expect(buildOf(detail(1))).toBe('12');
    expect(buildOf(detail(1, {info: bundleInfo({build: '1.2.0'})}))).toBeUndefined();
    expect(buildOf(detail(1, {info: undefined}))).toBeUndefined();
});

it('badges what a bundle is and what is risky about it, then its tags', () => {
    const labels = badgesOf(detail(1), ['rc']).map(b => b.label);
    expect(labels).toEqual(['.apk', 'Unity · 2022.3 · IL2CPP', 'debuggable', 'debug signed', 'rc']);
    expect(badgesOf(detail(1, {info: undefined}), [])).toEqual([]);
    const ios = detail(1, {info: bundleInfo({android: undefined, signing: undefined, apple: appleInfo({profile: {...profile, kind: 'ad-hoc'}})})});
    expect(badgesOf(ios, []).map(b => b.label)).toContain('Ad Hoc');
});

const info = (over = {}): BundleInfo => bundleInfo({android: undefined, apple: appleInfo(over)});
const phone = (over = {}): IDeviceData => device(1, {platformType: PlatformType.iOS, osVersion: '17.0', deviceIdentifier: 'udid-1', ...over});

it('fits an android bundle to the version of the device', () => {
    const android = bundleInfo({android: androidInfo({minSdk: '30'})});
    expect(deviceFit(android, device(1, {osVersion: '14'}))).toEqual({ok: true, reasons: []});
    expect(deviceFit(android, device(1, {osVersion: '10'})).reasons).toEqual(['needs Android 11, device has 10']);
    expect(deviceFit(android, device(1, {osVersion: ''})).ok).toBe(true);
});

it('keeps a simulator build off devices and a device build off the simulator', () => {
    expect(deviceFit(info({platform: 'iphonesimulator'}), phone()).reasons).toEqual(['built for the simulator']);
    const simulator = phone({platformType: PlatformType.IOSSimulator});
    expect(deviceFit(info({platform: 'iphoneos'}), simulator).reasons).toEqual(['built for devices, not the simulator']);
    expect(deviceFit(info({platform: 'iphonesimulator'}), simulator).ok).toBe(true);
});

it('checks the minimum os of an apple bundle', () => {
    expect(deviceFit(info({minOs: '18.0'}), phone()).reasons).toEqual(['needs iOS 18.0, device has 17.0']);
    const mac = phone({platformType: PlatformType.Mac});
    expect(deviceFit(info({minOs: '18.0', platform: 'macosx'}), mac).reasons).toEqual(['needs macOS 18.0, device has 17.0']);
});

it('checks the provisioning profile against the device', () => {
    const known = {...profile, devices: ['UDID-1']};
    expect(deviceFit(info({profile: known}), phone()).ok).toBe(true);
    expect(deviceFit(info({profile: {...known, allDevices: true}}), phone({deviceIdentifier: 'x'})).ok).toBe(true);
    expect(deviceFit(info({profile: known}), phone({deviceIdentifier: 'x'})).reasons).toEqual(['not in the provisioning profile']);
    expect(deviceFit(info({profile: {...known, devices: undefined}}), phone({deviceIdentifier: undefined})).ok).toBe(false);
    expect(deviceFit(info({profile: {...known, kind: 'app-store'}}), phone()).reasons)
        .toEqual(['App Store builds install from TestFlight or the store only']);
    expect(deviceFit(info({profile: undefined}), phone()).ok).toBe(true);
});
