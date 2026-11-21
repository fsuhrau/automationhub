// Bundles, apps and devices the tests of the bundle pages work with.
import {IAppBinaryData, IAppBundleDetail, IAppData} from '../../types/app';
import {AndroidInfo, AppleInfo, BundleInfo} from '../../types/bundle.info';
import IDeviceData from '../../types/device';
import {PlatformType} from '../../types/platform.type.enum';

export const PROJECT = 'demo';
export const APP_ID = 3;

export const androidApp = {id: APP_ID, name: 'Pixel Quest', identifier: 'com.demo.quest', platform: PlatformType.Android} as IAppData;

const FAR_PAST = '2001-02-03T04:05:06Z';
const FAR_FUTURE = '2999-02-03T04:05:06Z';

export const androidInfo = (over: Partial<AndroidInfo> = {}): AndroidInfo => ({
    packageName: 'com.demo.quest', versionName: '1.2.0', versionCode: '12', minSdk: '24', targetSdk: '34', compileSdk: '34',
    launchActivity: 'com.demo.Main', abis: ['arm64-v8a'], locales: ['en', 'de'], densities: ['65534', '480'], screens: ['normal'],
    permissions: [{name: 'android.permission.CAMERA'}, {name: 'android.permission.INTERNET', maxSdk: '30'}],
    features: [{name: 'android.hardware.camera', required: true}, {name: 'android.hardware.nfc', required: false}],
    debuggable: true, testOnly: false, leanback: false, ...over,
});

export const appleInfo = (over: Partial<AppleInfo> = {}): AppleInfo => ({
    platform: 'iphoneos', minOs: '15.0', bundleName: 'Quest', executable: 'Quest', xcode: '15.0', xcodeBuild: '15A',
    deviceFamily: ['iPhone'], orientations: ['UIInterfaceOrientationLandscapeLeft'], requiredCapabilities: ['arm64'],
    usageDescriptions: [{key: 'NSCameraUsageDescription', text: 'Scan codes'}], extensions: [], frameworks: ['UIKit'],
    fileSharing: true, ...over,
});

export const profile = {
    name: 'Dev profile', uuid: 'uuid-1', teamName: 'Team', teamId: 'T1', kind: 'development' as const, createdAt: FAR_PAST,
    expiresAt: FAR_FUTURE, allDevices: false, devices: ['UDID-1', 'UDID-2'], entitlements: {'aps-environment': 'development', 'b.key': ['x', 'y']},
};

export const certificate = {
    subject: 'CN=Debug, O=Demo', issuer: 'CN=Debug, O=Demo', serial: '1', notBefore: FAR_PAST, notAfter: FAR_FUTURE,
    sha1: 'a'.repeat(40), sha256: 'b'.repeat(64), debug: true,
};

export const bundleInfo = (over: Partial<BundleInfo> = {}): BundleInfo => ({
    format: 'apk', displayName: 'Quest', identifier: 'com.demo.quest', version: '1.2.0', build: '12', icon: 'icon.png',
    engine: {name: 'Unity', version: '2022.3', backend: 'IL2CPP'}, android: androidInfo(),
    signing: {schemes: ['v2', 'v3'], certificates: [certificate]},
    contents: {
        files: 1200, size: 5000000, compressed: 3000000,
        groups: [{path: 'lib', files: 3, size: 4000000, compressed: 2500000}, {path: 'assets', files: 1, size: 1000000, compressed: 500000}],
        largest: [{path: 'lib/libunity.so', size: 3000000, compressed: 1800000}],
    },
    warnings: ['Built without a keystore'], readAt: FAR_PAST, ...over,
});

export const bundle = (id: number, over: Partial<IAppBinaryData> = {}): IAppBinaryData => ({
    id, name: `quest-${id}.apk`, platform: 'Android', version: `1.${id}.0`, appPath: `quest-${id}_hash${id}.apk`, identifier: 'com.demo.quest',
    launchActivity: '', additional: '', hash: `hash${id}`.padEnd(40, '0'), size: 5000000, tags: '', runCount: 2,
    createdAt: new Date(FAR_PAST), updatedAt: new Date(FAR_PAST), deletedAt: new Date(FAR_PAST), ...over,
});

export const detail = (id: number, over: Partial<IAppBundleDetail> = {}): IAppBundleDetail => ({
    ...bundle(id), runCount: 2, info: bundleInfo(),
    runs: [
        {id: 40, testId: 7, testName: 'Smoke', createdAt: FAR_PAST, status: 'passed', passed: 3, failed: 1, unstable: 1, open: 0},
        {id: 41, testId: 8, testName: '', createdAt: FAR_PAST, status: 'weird', passed: 0, failed: 0, unstable: 0, open: 1},
    ],
    ...over,
});

export const iosDetail = (id: number): IAppBundleDetail => detail(id, {
    platform: 'iOS',
    info: bundleInfo({format: 'ipa', android: undefined, signing: undefined, apple: appleInfo({profile}), engine: undefined}),
});

export const device = (id: number, over: Partial<IDeviceData> = {}): IDeviceData => ({
    id, name: `Pixel ${id}`, alias: '', os: 'Android', osVersion: '14', platformType: PlatformType.Android, deviceIdentifier: `serial-${id}`,
    ...over,
} as IDeviceData);
