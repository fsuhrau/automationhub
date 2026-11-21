import React from 'react';
import {Grid} from '@mui/material';
import {routes} from '../../../lib/routes';
import {IAppBundleDetail} from '../../../types/app';
import IDeviceData from '../../../types/device';
import {BundleInfo} from '../../../types/bundle.info';
import {BundleScope} from '../useBundleScope';
import {CertificatesSection} from './CertificatesSection';
import {ContentsSection} from './ContentsSection';
import {DevicesSection} from './DevicesSection';
import {FileSection} from './FileSection';
import {PlatformCells} from './PlatformCells';
import {RunsSection} from './RunsSection';

interface BundleSectionsProps {
    detail: IAppBundleDetail,
    scope: BundleScope,
    // null while loading
    devices: IDeviceData[] | null,
    now: number,
}

const half = {xs: 12, md: 6};

interface InfoCellsProps extends BundleSectionsProps {
    info: BundleInfo,
}

// InfoCells are the sections of what the master read from the file
const InfoCells: React.FC<InfoCellsProps> = ({info, scope, devices, now}) => {
    const deviceUrl = (id: number): string => routes.project(scope.projectId).device(id).show();
    const {signing} = info;
    return <>
        <Grid size={half}><DevicesSection info={info} devices={devices} deviceUrl={deviceUrl}/></Grid>
        <PlatformCells info={info} devices={devices} now={now}/>
        {(signing || info.android) && <Grid size={half}>
            <CertificatesSection certificates={signing?.certificates ?? []} schemes={signing?.schemes} apple={!!info.apple} now={now}/>
        </Grid>}
        {info.contents && <Grid size={12}><ContentsSection contents={info.contents}/></Grid>}
    </>;
};

// BundleSections lay out everything the page knows about a bundle
export const BundleSections: React.FC<BundleSectionsProps> = props => {
    const {detail, scope, now} = props;
    return <Grid container={true} spacing={2}>
        <Grid size={half}><FileSection detail={detail} now={now}/></Grid>
        {detail.info && <InfoCells {...props} info={detail.info}/>}
        <Grid size={12}>
            <RunsSection detail={detail} now={now} runUrl={(testId, runId) => scope.routes.test(testId).run(runId).show()}/>
        </Grid>
    </Grid>;
};
