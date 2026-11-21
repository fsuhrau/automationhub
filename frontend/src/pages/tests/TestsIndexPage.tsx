import React, {useState} from 'react';
import {Box, Button, Card, CardContent, Stack, Typography} from '@mui/material';
import {AddRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import ITestData from '../../types/test';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useApplicationContext} from '../../hooks/ApplicationProvider';
import {useHubState} from '../../hooks/HubStateProvider';
import RunTestDialog from '../../components/RunTestDialog';
import {TestsListBody} from './definitions/parts/TestsListBody';
import {useAppRoutes} from './definitions/useAppRoutes';
import {useTestsList} from './definitions/useTestsList';

// TestsIndexPage lists the tests of the app picked in the menu with how
// they are doing, and starts them.
const TestsIndexPage: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {appId} = useApplicationContext();
    const {state} = useHubState();
    const appRoutes = useAppRoutes();
    const app = state.apps?.find(a => a.id === appId);
    const {tests, error, reload} = useTestsList(projectIdentifier, appId);
    const [running, setRunning] = useState<ITestData | null>(null);

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
                <Box sx={{minWidth: 0}}>
                    <Typography component="h1" variant="h5">Tests</Typography>
                    <Typography variant="body2" color="text.secondary">
                        Test setups of {app?.name ?? 'this app'}: what runs, on which devices. Open one for its runs, or start it.
                    </Typography>
                </Box>
                <Button variant="contained" startIcon={<AddRounded/>} sx={{ml: {md: 'auto'}}}
                    component={RouterLink} to={appRoutes.newTest()}>New test</Button>
            </Stack>
            <Card variant="outlined">
                <CardContent>
                    <TestsListBody tests={tests} error={error} onRetry={reload} appRoutes={appRoutes} canRun={!!app} onRun={setRunning}/>
                </CardContent>
            </Card>
            {app && <RunTestDialog projectIdentifier={projectIdentifier} app={app} test={running} onClose={() => setRunning(null)}/>}
        </Box>
    );
};

export default TestsIndexPage;
