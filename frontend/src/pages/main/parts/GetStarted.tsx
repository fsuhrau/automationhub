import React from 'react';
import {Button, Stack, Typography} from '@mui/material';
import {AddRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {IAppData} from '../../../types/app';
import {Section} from '../../../components/common';
import {routes} from '../../../lib/routes';

// a project without apps: they come first
export const GetStartedWithApps: React.FC<{ projectIdentifier: string }> = ({projectIdentifier}) => (
    <Section title="Get started">
        <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>
            Add the game's apps first (one per platform). Then upload a bundle and create a test.
        </Typography>
        <Button variant="contained" startIcon={<AddRounded/>} component={RouterLink} to={routes.project(projectIdentifier).settings('apps')}>
            Add an app
        </Button>
    </Section>
);

// a project with apps but no tests: one button per app
export const GetStartedWithTests: React.FC<{ projectIdentifier: string, apps: IAppData[] }> = ({projectIdentifier, apps}) => (
    <Section title="Get started">
        <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>
            The project has no tests yet. A test says what to run and on which devices.
        </Typography>
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap={true}>
            {apps.map(a => <Button key={a.id} variant="outlined" startIcon={<AddRounded/>} component={RouterLink}
                to={routes.project(projectIdentifier).app(a.id).newTest()}>Test for {a.name}</Button>)}
        </Stack>
    </Section>
);
