// The bundle details page without React: naming what the master read from a
// bundle file, and whether the project's devices can install it.
import IDeviceData from '../../types/device';
import {PlatformType} from '../../types/platform.type.enum';
import {BundleInfo} from '../../types/bundle.info';

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

// deviceFit says whether a project device can install and run the bundle,
// with every reason it can't
export const deviceFit = (info: BundleInfo, device: IDeviceData): Fit => {
    const reasons: string[] = [];
    const version = device.osVersion ?? '';
    if (info.android) {
        const min = androidVersionOf(info.android.minSdk);
        if (min && version && compareVersions(version, min) < 0) {
            reasons.push(`needs Android ${min}, device has ${version}`);
        }
    }
    if (info.apple) {
        const simulator = device.platformType === PlatformType.IOSSimulator;
        const platform = info.apple.platform ?? '';
        if (simulator && platform === 'iphoneos') reasons.push('built for devices, not the simulator');
        if (!simulator && platform === 'iphonesimulator') reasons.push('built for the simulator');
        const min = info.apple.minOs;
        if (min && version && compareVersions(version, min) < 0) {
            reasons.push(`needs ${device.platformType === PlatformType.Mac ? 'macOS' : 'iOS'} ${min}, device has ${version}`);
        }
        const profile = info.apple.profile;
        if (!simulator && device.platformType === PlatformType.iOS && profile && !profile.allDevices && profile.kind !== 'app-store') {
            const udids = (profile.devices ?? []).map(d => d.toLowerCase());
            if (!udids.includes((device.deviceIdentifier ?? '').toLowerCase())) {
                reasons.push('not in the provisioning profile');
            }
        }
        if (!simulator && device.platformType === PlatformType.iOS && profile?.kind === 'app-store') {
            reasons.push('App Store builds install from TestFlight or the store only');
        }
    }
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
