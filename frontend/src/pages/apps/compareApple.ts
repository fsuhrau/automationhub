import {AppleInfo, ProvisioningProfile} from '../../types/bundle.info';
import {CompareSection} from '../../types/bundle.compare';
import {entitlementValue, orientationName, PROFILE_KINDS, usageName} from './bundleDetailModel';
import {sides} from './compareRows';

const xcode = (p: AppleInfo): string | undefined => p.xcode && `${p.xcode}${p.xcodeBuild ? ` (${p.xcodeBuild})` : ''}`;

// appleSection compares what the Info.plist of two bundles says
export const appleSection = (a: AppleInfo | undefined, b: AppleInfo | undefined): CompareSection => {
    const s = sides(a, b);
    return {
        title: a?.platform === 'macosx' || b?.platform === 'macosx' ? 'macOS' : 'iOS', rows: [
            s.text('Bundle name', i => i.bundleName),
            s.text('Executable', i => i.executable),
            s.text('Minimum OS', i => i.minOs),
            s.text('SDK', i => i.sdk),
            s.text('Xcode', xcode),
            s.text('Built on macOS', i => i.buildMachineOs),
            s.flag('File sharing', i => i.fileSharing),
            s.list('Devices', i => i.deviceFamily),
            s.list('Orientations', i => i.orientations?.map(orientationName)),
            s.list('Requires', i => i.requiredCapabilities),
            s.list('Background modes', i => i.backgroundModes),
            s.list('URL schemes', i => i.urlSchemes),
            s.list('Permissions', i => i.usageDescriptions?.map(u => `${usageName(u.key)}: ${u.text}`)),
            s.list('Extensions', i => i.extensions),
            s.list('Frameworks', i => i.frameworks),
        ],
    };
};

const kind = (p: ProvisioningProfile): string => PROFILE_KINDS[p.kind] ?? p.kind;
const devices = (p: ProvisioningProfile): string => p.allDevices ? 'every device' : `${p.devices?.length ?? 0}`;
const day = (iso: string): string => new Date(iso).toISOString().slice(0, 10);
const entitlements = (p: ProvisioningProfile): string[] =>
    Object.entries(p.entitlements ?? {}).map(([k, v]) => `${k} = ${entitlementValue(v)}`);

// profileSection compares the provisioning profiles of two bundles
export const profileSection = (a: ProvisioningProfile | undefined, b: ProvisioningProfile | undefined): CompareSection => {
    const s = sides(a, b);
    return {
        title: 'Provisioning profile', rows: [
            s.text('Name', i => i.name),
            s.text('Kind', kind),
            s.text('Team', i => i.teamName),
            s.text('UUID', i => i.uuid),
            s.text('Expires', i => day(i.expiresAt)),
            s.text('Devices', devices),
            s.list('Device UDIDs', i => i.devices),
            s.list('Entitlements', entitlements),
        ],
    };
};
