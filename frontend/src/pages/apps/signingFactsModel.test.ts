import {expect, it} from 'vitest';
import {certificate, device, profile} from './appsFixtures';
import {certificateFacts, devicePool, devicesText, entitlementFacts, profileFacts, signingNote, usageFacts} from './signingFactsModel';

it('names usage descriptions by what they ask for', () => {
    expect(usageFacts([{key: 'NSCameraUsageDescription', text: 'Scan'}])).toEqual([{label: 'Camera', text: 'Scan'}]);
});

it('says which devices a profile installs on', () => {
    expect(devicesText({...profile, allDevices: true}, 0)).toBe('every device (enterprise)');
    expect(devicesText({...profile, devices: undefined}, 0)).toBe('none (installs from the App Store or TestFlight)');
    expect(devicesText(profile, 1)).toBe('2 devices, 1 of them in the device pool');
});

it('lists the facts of a profile with the team and the expiry', () => {
    const facts = profileFacts(profile, 1);
    expect(facts.find(f => f.label === 'Team')?.text).toBe('Team (T1)');
    expect(facts.find(f => f.label === 'Expires')?.expiry).toBe(profile.expiresAt);
    expect(profileFacts({...profile, teamId: undefined, teamName: undefined}, 0).find(f => f.label === 'Team')?.text).toBeUndefined();
    expect(profileFacts({...profile, teamId: undefined}, 0).find(f => f.label === 'Team')?.text).toBe('Team');
});

it('sorts the entitlements by name', () => {
    expect(entitlementFacts(profile).map(f => f.label)).toEqual(['aps-environment', 'b.key']);
    expect(entitlementFacts({...profile, entitlements: undefined})).toEqual([]);
});

it('lists the facts of a certificate and says when it signs itself', () => {
    expect(certificateFacts(certificate).find(f => f.label === 'Issuer')?.text).toBe('self-signed');
    expect(certificateFacts({...certificate, issuer: 'CN=CA,O=X'}).find(f => f.label === 'Issuer')?.text).toBe('CA');
});

it('says how the certificates of a bundle came to be', () => {
    expect(signingNote(true, undefined)).toMatch(/provisioning profile/);
    expect(signingNote(false, ['v2', 'v3'])).toBe('APK signature scheme v2, v3');
    expect(signingNote(false, [])).toBe('No signature found');
    expect(signingNote(false, undefined)).toBe('No signature found');
});

it('finds devices by their identifier in any case', () => {
    const pool = devicePool([device(1, {deviceIdentifier: 'AbC'}), device(2, {deviceIdentifier: undefined})]);
    expect(pool.get('abc')?.id).toBe(1);
    expect(pool.get('')?.id).toBe(2);
});
