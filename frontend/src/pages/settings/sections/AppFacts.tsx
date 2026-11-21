import React from 'react';
import {Box, Link, Stack, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import PlatformTypeIcon from '../../../components/PlatformTypeIcon';
import {CopyButton} from '../../../components/common';
import {routes} from '../../../lib/routes';
import {IAppData} from '../../../types/app';
import {identifierLabel, platformLabel} from '../settingsModel';

const Label: React.FC<{ children: string }> = ({children}) =>
    <Typography variant="body2" color="text.secondary">{children}</Typography>;

const IdentifierFact: React.FC<{ app: IAppData }> = ({app}) => <>
    <Label>{identifierLabel(app.platform)}</Label>
    <Stack direction="row" spacing={0.5} alignItems="center">
        <Typography variant="body2" sx={{fontFamily: 'monospace', wordBreak: 'break-all'}}>{app.identifier}</Typography>
        <CopyButton text={app.identifier} label="Copy"/>
    </Stack>
</>;

// AppFacts is what the hub knows about the app, and links to its pages
const AppFacts: React.FC<{ app: IAppData, projectIdentifier: string }> = ({app, projectIdentifier}) => {
    const appRoutes = routes.project(projectIdentifier).app(app.id);
    return <Box sx={{display: 'grid', gridTemplateColumns: 'max-content 1fr', columnGap: 2, rowGap: 0.75, alignItems: 'center'}}>
        <Label>Platform</Label>
        <Stack direction="row" spacing={1} alignItems="center">
            <PlatformTypeIcon platformType={app.platform} fontSize="small"/>
            <Typography variant="body2">{platformLabel(app.platform)}</Typography>
        </Stack>
        {app.identifier && <IdentifierFact app={app}/>}
        <Label>Hub id</Label>
        <Typography variant="body2">{app.id}</Typography>
        <Label>Go to</Label>
        <Stack direction="row" spacing={2}>
            <Link component={RouterLink} to={appRoutes.tests()} variant="body2">Tests</Link>
            <Link component={RouterLink} to={appRoutes.bundles()} variant="body2">Bundles</Link>
        </Stack>
    </Box>;
};

export default AppFacts;
