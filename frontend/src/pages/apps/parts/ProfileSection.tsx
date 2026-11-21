import React, {useState} from 'react';
import {Box, Button, Chip, Stack, Typography} from '@mui/material';
import {Section} from '../../../components/common';
import IDeviceData from '../../../types/device';
import {ProvisioningProfile} from '../../../types/bundle.info';
import {PROFILE_KINDS} from '../bundleDetailModel';
import {devicePool, DevicePool, entitlementFacts, profileFacts} from '../signingFactsModel';
import {Facts} from './Facts';
import {mono} from './mono';

interface ProfileDevicesProps {
    udids: string[],
    pool: DevicePool,
}

const UdidRow: React.FC<{ udid: string, device?: IDeviceData }> = ({udid, device}) =>
    <Stack direction="row" spacing={1} alignItems="center" sx={{py: 0.25}}>
        <Typography variant="body2" sx={{...mono, flex: '0 1 auto'}}>{udid}</Typography>
        {device && <Chip size="small" color="primary" variant="outlined" label={device.alias || device.name}/>}
    </Stack>;

const ProfileDevices: React.FC<ProfileDevicesProps> = ({udids, pool}) => {
    const [showAll, setShowAll] = useState(false);
    const inPool = udids.filter(u => pool.has(u.toLowerCase()));
    return <Box sx={{mt: 1}}>
        {(showAll ? udids : inPool).map(u => <UdidRow key={u} udid={u} device={pool.get(u.toLowerCase())}/>)}
        {udids.length > inPool.length && <Button size="small" onClick={() => setShowAll(s => !s)}>
            {showAll ? 'Show only devices in the pool' : `Show all ${udids.length} devices`}
        </Button>}
    </Box>;
};

const Entitlements: React.FC<{ profile: ProvisioningProfile, now: number }> = ({profile, now}) => {
    const [shown, setShown] = useState(false);
    const facts = entitlementFacts(profile);
    if (facts.length === 0) return null;
    return <Box sx={{mt: 1}}>
        <Button size="small" onClick={() => setShown(s => !s)}>{shown ? 'Hide entitlements' : `Show ${facts.length} entitlements`}</Button>
        {shown && <Facts facts={facts} now={now}/>}
    </Box>;
};

interface ProfileSectionProps {
    profile: ProvisioningProfile,
    devices: IDeviceData[],
    now: number,
}

export const ProfileSection: React.FC<ProfileSectionProps> = ({profile, devices, now}) => {
    const pool = devicePool(devices);
    const udids = profile.devices ?? [];
    const inPool = udids.filter(u => pool.has(u.toLowerCase())).length;
    return <Section title="Provisioning profile" description={PROFILE_KINDS[profile.kind] ?? profile.kind} fullHeight={true}>
        <Facts facts={profileFacts(profile, inPool)} now={now}/>
        {udids.length > 0 && <ProfileDevices udids={udids} pool={pool}/>}
        <Entitlements profile={profile} now={now}/>
    </Section>;
};
