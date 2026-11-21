import React from 'react';
import {Grid} from '@mui/material';
import IDeviceData from '../../../types/device';
import {BundleInfo} from '../../../types/bundle.info';
import {AndroidSection} from './AndroidSection';
import {AppleSection, UsageSection} from './AppleSection';
import {PermissionsSection} from './PermissionsSection';
import {ProfileSection} from './ProfileSection';

interface PlatformCellsProps {
    info: BundleInfo,
    devices: IDeviceData[] | null,
    now: number,
}

const half = {xs: 12, md: 6};

// PlatformCells are the sections of what the manifest or Info.plist says
export const PlatformCells: React.FC<PlatformCellsProps> = ({info, devices, now}) => {
    const {android, apple} = info;
    return <>
        {android && <>
            <Grid size={half}><AndroidSection android={android} now={now}/></Grid>
            <Grid size={half}><PermissionsSection android={android}/></Grid>
        </>}
        {apple && <>
            <Grid size={half}><AppleSection apple={apple} info={info} now={now}/></Grid>
            <Grid size={half}><UsageSection apple={apple} now={now}/></Grid>
            {apple.profile && <Grid size={half}><ProfileSection profile={apple.profile} devices={devices ?? []} now={now}/></Grid>}
        </>}
    </>;
};
