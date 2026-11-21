import React from 'react';
import Grid from '@mui/material/Grid';
import {IProjectOverview} from '../../../types/overview.project';
import {byteFormat} from '../../tests/value_formatter';
import {routes} from '../../../lib/routes';
import {Readiness} from '../dashboardModel';
import Tile from './Tile';

interface DashboardTilesProps {
    projectIdentifier: string,
    deviceCount: number,
    readiness: Readiness,
    appCount: number,
    overview: IProjectOverview | null,
}

const deviceNote = (deviceCount: number, inUse: number): string => {
    if (inUse) return `${inUse} in use`;
    return deviceCount ? 'none in use' : 'no devices yet';
};

// what the overview adds to the devices and apps
const OverviewTiles: React.FC<{ overview: IProjectOverview }> = ({overview}) => <>
    <Grid size={{xs: 6, md: 3}}>
        <Tile label="Tests" value={overview.tests} note={overview.failing.length ? `${overview.failing.length} need attention` : undefined}/>
    </Grid>
    <Grid size={{xs: 6, md: 3}}>
        <Tile label="Bundles" value={overview.bundles} note={`${byteFormat(overview.bundleStorage)} stored`}/>
    </Grid>
</>;

// devices and apps are known without the overview, so they show even when it fails
const DashboardTiles: React.FC<DashboardTilesProps> = ({projectIdentifier, deviceCount, readiness, appCount, overview}) => {
    const project = routes.project(projectIdentifier);
    return <Grid container={true} spacing={1.5} sx={{mb: 2}}>
        <Grid size={{xs: 6, md: 3}}>
            <Tile label="Devices ready" value={deviceCount ? `${readiness.ready} of ${deviceCount}` : '0'}
                note={deviceNote(deviceCount, readiness.inUse)} to={project.devices()}/>
        </Grid>
        <Grid size={{xs: 6, md: 3}}>
            <Tile label="Apps" value={overview?.apps ?? appCount} note="platforms of the game" to={project.settings('apps')}/>
        </Grid>
        {overview && <OverviewTiles overview={overview}/>}
    </Grid>;
};

export default DashboardTiles;
