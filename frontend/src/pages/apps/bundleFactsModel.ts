// The facts the details page lists about what the master read from a bundle
// file: the file, the Android manifest and the iOS Info.plist.
import {formatDateTime} from '../../components/DateTime';
import {IAppBundleDetail} from '../../types/app';
import {AndroidInfo, AppleInfo, BundleInfo} from '../../types/bundle.info';
import {FactSpec} from '../../types/bundle.fact';
import {byteFormat} from '../tests/value_formatter';
import {apiLevel, engineLabel, fileName, orientationName} from './bundleDetailModel';
import {ago} from './bundlesModel';

const DENSITY_NAMES: Record<string, string> = {'65534': 'any', '65535': 'none'};

const optionalLevel = (level: string | undefined): string | undefined => level ? apiLevel(level) : undefined;

const androidFlags = (a: AndroidInfo): string[] =>
    [a.debuggable && 'debuggable', a.testOnly && 'test only', a.leanback && 'Android TV'].filter((f): f is string => !!f);

// counted is a label with the size of its list: "Locales (3)"
const counted = (label: string, items: string[] | undefined): string => `${label} (${items?.length ?? 0})`;

export const androidFacts = (a: AndroidInfo): FactSpec[] => [
    {label: 'Package', text: a.packageName, copy: true, monospace: true},
    {label: 'Version name', text: a.versionName},
    {label: 'Version code', text: a.versionCode},
    {label: 'Min SDK', text: apiLevel(a.minSdk)},
    {label: 'Target SDK', text: apiLevel(a.targetSdk)},
    {label: 'Max SDK', text: optionalLevel(a.maxSdk)},
    {label: 'Compile SDK', text: optionalLevel(a.compileSdk)},
    {label: 'Launch activity', text: a.launchActivity, copy: true, monospace: true},
    {label: 'ABIs', chips: a.abis},
    {label: 'OpenGL ES', text: a.gles},
    {label: 'Screens', text: a.screens?.join(', ')},
    {label: 'Densities', text: a.densities?.map(d => DENSITY_NAMES[d] ?? d).join(', ')},
    {label: counted('Locales', a.locales), chips: a.locales, max: 24},
    {label: 'Flags', chips: androidFlags(a)},
];

const appleVersionFacts = (apple: AppleInfo, info: BundleInfo): FactSpec[] => [
    {label: 'Bundle identifier', text: info.identifier, copy: true, monospace: true},
    {label: 'Version', text: info.version},
    {label: 'Build', text: info.build},
    {label: 'Bundle name', text: apple.bundleName},
    {label: 'Executable', text: apple.executable, monospace: true},
    {label: 'Platform', text: apple.platform},
    {label: 'Minimum OS', text: apple.minOs},
    {label: 'SDK', text: apple.sdk},
    {label: 'Xcode', text: apple.xcode && `${apple.xcode}${apple.xcodeBuild ? ` (${apple.xcodeBuild})` : ''}`},
    {label: 'Built on macOS', text: apple.buildMachineOs},
];

const appleCapabilityFacts = (apple: AppleInfo): FactSpec[] => [
    {label: 'Devices', text: apple.deviceFamily?.join(', ')},
    {label: 'Orientations', text: apple.orientations?.map(orientationName).join(', ')},
    {label: 'Requires', chips: apple.requiredCapabilities},
    {label: 'Background modes', chips: apple.backgroundModes},
    {label: 'URL schemes', chips: apple.urlSchemes},
    {label: 'File sharing', text: apple.fileSharing ? 'on' : undefined},
    {label: counted('Extensions', apple.extensions), chips: apple.extensions},
    {label: counted('Frameworks', apple.frameworks), chips: apple.frameworks, max: 16},
];

export const appleFacts = (apple: AppleInfo, info: BundleInfo): FactSpec[] =>
    [...appleVersionFacts(apple, info), ...appleCapabilityFacts(apple)];

const uploaded = (b: IAppBundleDetail, now: number): string => `${formatDateTime(b.createdAt)} (${ago(b.createdAt, now)})`;

// the launch activity and executable as the master stores them, for bundles
// whose details were not read from the file
const storedFacts = (b: IAppBundleDetail): FactSpec[] => [
    {label: 'Launch activity', text: b.info?.android ? undefined : b.android?.LaunchActivity, monospace: true},
    {label: 'Executable', text: b.info?.apple ? undefined : b.executable?.Executable, monospace: true},
];

export const fileFacts = (b: IAppBundleDetail, now: number): FactSpec[] => [
    {label: 'File name', text: fileName(b.appPath, b.hash), monospace: true},
    {label: 'Size', text: b.size ? byteFormat(b.size) : undefined},
    {label: 'Installed size', text: b.info?.contents ? byteFormat(b.info.contents.size) : undefined},
    {label: 'SHA-1', text: b.hash, copy: true, monospace: true},
    {label: 'Uploaded', text: uploaded(b, now)},
    {label: 'Details read', text: b.info?.readAt ? formatDateTime(b.info.readAt) : undefined},
    {label: 'Platform', text: b.platform},
    {label: 'Identifier', text: b.identifier, copy: true, monospace: true},
    ...storedFacts(b),
    {label: 'Engine', text: engineLabel(b.info?.engine)},
];
