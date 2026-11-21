// The bundle details page without React: naming what the master read from a
// bundle file, and whether the project's devices can install it.
import IDeviceData from '../../types/device';
import {PlatformType} from '../../types/platform.type.enum';
import {BundleRun, IAppBundleDetail} from '../../types/app';
import {AndroidInfo, AppleInfo, BundleInfo, Engine, ProvisioningProfile} from '../../types/bundle.info';
import {bundleVersion} from './bundlesModel';

// Android version of an API level
const ANDROID_VERSIONS: Record<number, string> = {
    16: '4.1', 17: '4.2', 18: '4.3', 19: '4.4', 20: '4.4W', 21: '5.0', 22: '5.1', 23: '6', 24: '7.0', 25: '7.1',
    26: '8.0', 27: '8.1', 28: '9', 29: '10', 30: '11', 31: '12', 32: '12L', 33: '13', 34: '14', 35: '15', 36: '16', 37: '17',
};

// apiLevel names an API level with its Android version: "24 (Android 7.0)"
export const apiLevel = (level: string | undefined): string => {
    if (!level) return '–';
    const name = ANDROID_VERSIONS[+level];
    return name ? `${level} (Android ${name})` : level;
};

// androidVersionOf is the lowest Android version of an API level, for
// comparing with the version a device reports
export const androidVersionOf = (level: string | undefined): string | undefined =>
    level ? ANDROID_VERSIONS[+level]?.replace(/[^\d.]/g, '') : undefined;

// compareVersions compares dotted versions numerically: 10.0 > 9.3
export const compareVersions = (a: string, b: string): number => {
    const pa = a.split(/[.\-_ ]/).map(p => parseInt(p, 10) || 0);
    const pb = b.split(/[.\-_ ]/).map(p => parseInt(p, 10) || 0);
    for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
        const d = (pa[i] ?? 0) - (pb[i] ?? 0);
        if (d !== 0) return d;
    }
    return 0;
};

// shortPermission drops the android.permission. prefix
export const shortPermission = (name: string): string => name.replace(/^android\.permission\./, '');

// permissions a user has to grant at runtime, worth pointing out
const DANGEROUS = new Set([
    'ACCESS_FINE_LOCATION', 'ACCESS_COARSE_LOCATION', 'ACCESS_BACKGROUND_LOCATION', 'CAMERA', 'RECORD_AUDIO',
    'READ_CONTACTS', 'WRITE_CONTACTS', 'GET_ACCOUNTS', 'READ_CALENDAR', 'WRITE_CALENDAR', 'READ_PHONE_STATE',
    'READ_PHONE_NUMBERS', 'CALL_PHONE', 'READ_CALL_LOG', 'WRITE_CALL_LOG', 'SEND_SMS', 'RECEIVE_SMS', 'READ_SMS',
    'READ_EXTERNAL_STORAGE', 'WRITE_EXTERNAL_STORAGE', 'READ_MEDIA_IMAGES', 'READ_MEDIA_VIDEO', 'READ_MEDIA_AUDIO',
    'BODY_SENSORS', 'ACTIVITY_RECOGNITION', 'POST_NOTIFICATIONS', 'BLUETOOTH_CONNECT', 'BLUETOOTH_SCAN',
    'NEARBY_WIFI_DEVICES', 'UWB_RANGING',
]);
export const isDangerous = (name: string): boolean => name.startsWith('android.permission.') && DANGEROUS.has(shortPermission(name));

// usageName turns NSCameraUsageDescription into "Camera"
export const usageName = (key: string): string =>
    key.replace(/^NS/, '').replace(/UsageDescription$/, '').replace(/([a-z])([A-Z])/g, '$1 $2');

// orientationName turns UIInterfaceOrientationLandscapeLeft into "Landscape left"
export const orientationName = (o: string): string => {
    const s = o.replace(/^UIInterfaceOrientation/, '').replace(/([a-z])([A-Z])/g, '$1 $2').toLowerCase();
    return s.charAt(0).toUpperCase() + s.slice(1);
};

export const PROFILE_KINDS: Record<string, string> = {
    'development': 'Development',
    'ad-hoc': 'Ad Hoc',
    'enterprise': 'Enterprise (In-House)',
    'app-store': 'App Store',
};

export type Expiry = 'expired' | 'soon' | 'valid';

// expiry says whether a date has passed or is within 30 days
export const expiry = (date: string | Date, now: number): Expiry => {
    const t = new Date(date).valueOf();
    if (t < now) return 'expired';
    if (t - now < 30 * 86400 * 1000) return 'soon';
    return 'valid';
};

// commonName is the CN of an X.509 subject, else the whole subject
export const commonName = (subject: string): string => {
    const m = /(?:^|,)CN=([^,]+)/.exec(subject);
    return m ? m[1] : subject;
};

export type Fit = { ok: boolean, reasons: string[] };

const tooOld = (version: string, min: string | undefined): boolean => !!min && !!version && compareVersions(version, min) < 0;

const androidReasons = (android: AndroidInfo, version: string): string[] => {
    const min = androidVersionOf(android.minSdk);
    return tooOld(version, min) ? [`needs Android ${min}, device has ${version}`] : [];
};

const buildTargetReasons = (apple: AppleInfo, simulator: boolean): string[] => {
    const platform = apple.platform ?? '';
    if (simulator && platform === 'iphoneos') return ['built for devices, not the simulator'];
    if (!simulator && platform === 'iphonesimulator') return ['built for the simulator'];
    return [];
};

const profileReasons = (profile: ProvisioningProfile | undefined, device: IDeviceData): string[] => {
    if (!profile) return [];
    if (profile.kind === 'app-store') return ['App Store builds install from TestFlight or the store only'];
    const udids = (profile.devices ?? []).map(d => d.toLowerCase());
    const known = profile.allDevices || udids.includes((device.deviceIdentifier ?? '').toLowerCase());
    return known ? [] : ['not in the provisioning profile'];
};

const appleReasons = (apple: AppleInfo, device: IDeviceData): string[] => {
    const simulator = device.platformType === PlatformType.IOSSimulator;
    const version = device.osVersion ?? '';
    const reasons = buildTargetReasons(apple, simulator);
    if (tooOld(version, apple.minOs)) {
        reasons.push(`needs ${device.platformType === PlatformType.Mac ? 'macOS' : 'iOS'} ${apple.minOs}, device has ${version}`);
    }
    // only a real iPhone or iPad needs a profile that knows it
    if (!simulator && device.platformType === PlatformType.iOS) reasons.push(...profileReasons(apple.profile, device));
    return reasons;
};

// deviceFit says whether a project device can install and run the bundle,
// with every reason it can't
export const deviceFit = (info: BundleInfo, device: IDeviceData): Fit => {
    const reasons = [
        ...(info.android ? androidReasons(info.android, device.osVersion ?? '') : []),
        ...(info.apple ? appleReasons(info.apple, device) : []),
    ];
    return {ok: reasons.length === 0, reasons};
};

// platformsOf are the device platforms a bundle installs on
export const platformsOf = (appPlatform: PlatformType | undefined): PlatformType[] => {
    switch (appPlatform) {
        case PlatformType.iOS:
        case PlatformType.IOSSimulator:
            return [PlatformType.iOS, PlatformType.IOSSimulator];
        case undefined:
            return [];
        default:
            return [appPlatform];
    }
};

// fileName is the name a bundle was uploaded with, without the hash the
// master adds to it
export const fileName = (appPath: string, hash: string): string => {
    if (!appPath) return '–';
    return hash ? appPath.replace(`_${hash}`, '') : appPath;
};

// entitlementValue shows an entitlement in one line
export const entitlementValue = (v: unknown): string => {
    if (Array.isArray(v)) return v.map(entitlementValue).join(', ');
    if (v !== null && typeof v === 'object') return JSON.stringify(v);
    return String(v);
};

// engineLabel is the game engine of a bundle in one line: "Unity · 2022.3 · IL2CPP"
export const engineLabel = (engine: Engine | undefined): string | undefined =>
    engine && [engine.name, engine.version, engine.backend].filter(Boolean).join(' · ');

// bundleTitle is what to call a bundle on its page
export const bundleTitle = (b: IAppBundleDetail): string => b.info?.displayName || b.name || b.identifier || `Bundle #${b.id}`;

// buildOf is the build number of a bundle when it says more than the version
export const buildOf = (b: IAppBundleDetail): string | undefined => {
    const build = b.info?.build;
    return build && build !== bundleVersion(b) ? build : undefined;
};

// runSummary counts the results of a test run that used a bundle, as text
export const runSummary = (r: BundleRun): string => {
    if (r.passed + r.failed + r.unstable === 0) return '';
    return `${r.passed} passed${r.failed ? `, ${r.failed} failed` : ''}${r.unstable ? `, ${r.unstable} unstable` : ''}`;
};

export interface Badge {
    label: string,
    // warnings stand out, the rest is outlined
    warn?: boolean,
    outlined?: boolean,
}

type BadgeRule = (b: IAppBundleDetail) => Badge | undefined;

const badgeRules: BadgeRule[] = [
    b => b.info?.format ? {label: `.${b.info.format}`, outlined: true} : undefined,
    b => b.info?.engine ? {label: engineLabel(b.info.engine) ?? '', outlined: true} : undefined,
    b => b.info?.android?.debuggable ? {label: 'debuggable', warn: true} : undefined,
    b => b.info?.signing?.certificates?.some(c => c.debug) ? {label: 'debug signed', warn: true} : undefined,
    b => {
        const kind = b.info?.apple?.profile?.kind;
        return kind ? {label: PROFILE_KINDS[kind] ?? kind, outlined: true} : undefined;
    },
];

// badgesOf are the chips under the title of a bundle: what it is, what is
// risky about it, and its tags
export const badgesOf = (b: IAppBundleDetail, tags: string[]): Badge[] => [
    ...badgeRules.map(rule => rule(b)).filter((x): x is Badge => x !== undefined),
    ...tags.map(label => ({label})),
];

// shareOf is the part of the whole a size makes up, 0..1
export const shareOf = (size: number, total: number): number => total > 0 ? size / total : 0;

// barWidth is the width in percent of the bar of a size next to the longest; never invisible
export const barWidth = (size: number, max: number): number => Math.max(0.5, size / max * 100);
