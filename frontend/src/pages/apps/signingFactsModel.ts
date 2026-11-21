// The facts about how a bundle is signed and what it asks permission for.
import {formatDateTime} from '../../components/DateTime';
import IDeviceData from '../../types/device';
import {Certificate, ProvisioningProfile, UsageDescription} from '../../types/bundle.info';
import {FactSpec} from '../../types/bundle.fact';
import {commonName, entitlementValue, usageName} from './bundleDetailModel';

export const usageFacts = (usage: UsageDescription[]): FactSpec[] => usage.map(u => ({label: usageName(u.key), text: u.text}));

const teamOf = (p: ProvisioningProfile): string | undefined => p.teamName && `${p.teamName}${p.teamId ? ` (${p.teamId})` : ''}`;

// devicesText says which devices a profile installs on
export const devicesText = (p: ProvisioningProfile, inPool: number): string => {
    const udids = p.devices ?? [];
    if (p.allDevices) return 'every device (enterprise)';
    if (udids.length === 0) return 'none (installs from the App Store or TestFlight)';
    return `${udids.length} devices, ${inPool} of them in the device pool`;
};

export const profileFacts = (p: ProvisioningProfile, inPool: number): FactSpec[] => [
    {label: 'Name', text: p.name},
    {label: 'Team', text: teamOf(p)},
    {label: 'App ID', text: p.appIdName},
    {label: 'UUID', text: p.uuid, copy: true, monospace: true},
    {label: 'Created', text: formatDateTime(p.createdAt)},
    {label: 'Expires', expiry: p.expiresAt},
    {label: 'Devices', text: devicesText(p, inPool)},
];

// entitlementFacts are the entitlements of a profile, sorted by name
export const entitlementFacts = (p: ProvisioningProfile): FactSpec[] =>
    Object.entries(p.entitlements ?? {}).sort(([a], [b]) => a.localeCompare(b))
        .map(([label, v]) => ({label, text: entitlementValue(v), monospace: true}));

export const certificateFacts = (c: Certificate): FactSpec[] => [
    {label: 'Subject', text: c.subject},
    {label: 'Issuer', text: c.issuer === c.subject ? 'self-signed' : commonName(c.issuer)},
    {label: 'Valid from', text: formatDateTime(c.notBefore)},
    {label: 'Valid until', expiry: c.notAfter},
    {label: 'SHA-1', text: c.sha1, copy: true, monospace: true},
    {label: 'SHA-256', text: c.sha256, copy: true, monospace: true},
];

// signingNote says how the certificates of a bundle came to be
export const signingNote = (apple: boolean, schemes: string[] | undefined): string => {
    if (apple) return 'Certificates the provisioning profile allows to sign the app';
    return schemes?.length ? `APK signature scheme ${schemes.join(', ')}` : 'No signature found';
};

export type DevicePool = Map<string, IDeviceData>;

// devicePool finds devices of the project by their identifier (the udid)
export const devicePool = (devices: IDeviceData[]): DevicePool =>
    new Map(devices.map(d => [(d.deviceIdentifier ?? '').toLowerCase(), d]));
