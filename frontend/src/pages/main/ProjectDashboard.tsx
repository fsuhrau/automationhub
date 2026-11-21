import React from 'react';
import {Alert, Box, Button, LinearProgress, Typography} from '@mui/material';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import {useNow} from '../../hooks/useNow';
import {OverviewMissing} from '../../services/overview.service';
import {IOverviewRun} from '../../types/overview.run';
import {IProjectOverview} from '../../types/overview.project';
import {IAppData} from '../../types/app';
import PlatformTypeIcon from '../../components/PlatformTypeIcon';
import {routes} from '../../lib/routes';
import {readiness, stageOf} from './dashboardModel';
import {OverviewState, useOverview} from './useOverview';
import DashboardTiles from './parts/DashboardTiles';
import {GetStartedWithApps, GetStartedWithTests} from './parts/GetStarted';
import RunSections from './parts/RunSections';

const REFRESH_TIME_MS = 30000;

const AppName: React.FC<{ app: IAppData }> = ({app}) => <>
    <PlatformTypeIcon platformType={app.platform} sx={{fontSize: 12, verticalAlign: 'middle', mr: 0.5}}/>{app.name}
</>;

// why the overview is missing, or that it is on its way
const OverviewStatus: React.FC<Pick<OverviewState, 'overview' | 'error' | 'outdated' | 'retry'>> = ({overview, error, outdated, retry}) => <>
    {error && <Alert severity="error" sx={{mb: 2}} action={<Button color="inherit" size="small" onClick={retry}>Retry</Button>}>{error}</Alert>}
    {outdated && <Alert severity="info" sx={{mb: 2}}>{new OverviewMissing().message}</Alert>}
    {!overview && !error && !outdated && <LinearProgress sx={{mb: 2}}/>}
</>;

interface BodyProps {
    overview: IProjectOverview,
    projectIdentifier: string,
    apps: IAppData[],
}

// the overview's first section: the next step for a new project, else the runs
const OverviewBody: React.FC<BodyProps> = ({overview, projectIdentifier, apps}) => {
    const now = useNow(REFRESH_TIME_MS);
    const stage = stageOf(overview);
    if (stage === 'noApps') return <GetStartedWithApps projectIdentifier={projectIdentifier}/>;
    if (stage === 'noTests') return <GetStartedWithTests projectIdentifier={projectIdentifier} apps={apps}/>;
    const appName = (id: number): React.ReactNode => {
        const app = apps.find(a => a.id === id);
        return app ? <AppName app={app}/> : `app ${id}`;
    };
    const runUrl = (r: IOverviewRun): string => routes.project(projectIdentifier).app(r.appId).test(r.testId).run(r.id).show();
    return <RunSections overview={overview} runUrl={runUrl} appName={appName} now={now}/>;
};

// ProjectDashboard is the project's home: what needs attention, what runs,
// what ran, and whether devices are ready
const ProjectDashboard: React.FC = () => {
    const {project, projectIdentifier} = useProjectContext();
    const {state} = useHubState();
    const status = useOverview(projectIdentifier);
    const apps = state.apps ?? [];
    const devices = state.devices ?? [];

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Typography component="h1" variant="h5">{project.name}</Typography>
            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
                How the project's tests are doing, what runs now, and whether devices are ready.
            </Typography>
            <OverviewStatus {...status}/>
            <DashboardTiles projectIdentifier={projectIdentifier} deviceCount={devices.length} readiness={readiness(devices)}
                appCount={apps.length} overview={status.overview}/>
            {status.overview && <OverviewBody overview={status.overview} projectIdentifier={projectIdentifier} apps={apps}/>}
        </Box>
    );
};

export default ProjectDashboard;
