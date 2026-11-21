import React, {useState} from 'react';
import {Box, Button, Stack, Typography} from '@mui/material';
import Grid from '@mui/material/Grid';
import {EditRounded, PlayArrowRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import ITestData from '../../types/test';
import {routes} from '../../lib/routes';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import RunTestDialog from '../../components/RunTestDialog';
import {describeConfig} from './testsListModel';
import {RunStats} from './definitions/parts/RunStats';
import {RunsSection} from './definitions/parts/RunsSection';
import {WhatItRunsCard} from './definitions/parts/WhatItRunsCard';
import {WhereItRunsCard} from './definitions/parts/WhereItRunsCard';
import {useAppRoutes} from './definitions/useAppRoutes';
import {useTestRuns} from './definitions/useTestRuns';

// ShowTestPage shows a test: how it has been doing, what it runs where, and
// its runs; it starts and edits the test.
const ShowTestPage: React.FC<{ test: ITestData }> = ({test}) => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const testRoutes = useAppRoutes().test(test.id);
    const app = state.apps?.find(a => a.id === appId);
    const runs = useTestRuns(projectIdentifier, appId, test.id);
    const [runDialog, setRunDialog] = useState(false);
    // older masters don't summarize runs; then there is nothing to sum up
    const summarized = runs.runs?.some(r => r.status !== undefined) ? runs.runs : null;

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{test.name}</Typography>
                    <Typography variant="body2" color="text.secondary">{describeConfig(test.testConfig)}</Typography>
                </Box>
                <Stack direction="row" spacing={1} sx={{ml: {md: 'auto'}}}>
                    <Button variant="outlined" startIcon={<EditRounded/>} component={RouterLink} to={testRoutes.edit()}>Edit</Button>
                    <Button variant="contained" startIcon={<PlayArrowRounded/>} disabled={!app} onClick={() => setRunDialog(true)}>Run</Button>
                </Stack>
            </Stack>
            <Grid container={true} spacing={2}>
                <Grid size={{xs: 12, lg: 8}}>
                    <Stack spacing={2}>
                        {summarized && <RunStats runs={summarized}/>}
                        <RunsSection runs={runs} testRoutes={testRoutes} canRun={!!app} onRun={() => setRunDialog(true)}/>
                    </Stack>
                </Grid>
                <Grid size={{xs: 12, lg: 4}}>
                    <Stack spacing={2}>
                        <WhatItRunsCard config={test.testConfig} editUrl={testRoutes.edit()}/>
                        <WhereItRunsCard config={test.testConfig} projectRoutes={routes.project(projectIdentifier)}/>
                    </Stack>
                </Grid>
            </Grid>
            {app && <RunTestDialog projectIdentifier={projectIdentifier} app={app} test={runDialog ? test : null}
                onClose={() => setRunDialog(false)}/>}
        </Box>
    );
};

export default ShowTestPage;
