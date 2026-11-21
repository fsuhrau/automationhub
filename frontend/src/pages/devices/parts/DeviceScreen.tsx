import React from 'react';
import {Box, ButtonBase, IconButton, Stack, Tooltip, Typography} from '@mui/material';
import {BrokenImageRounded, RefreshRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {Section} from '../../../components/common';
import {DeviceRoutes} from '../../../lib/routes';
import {Availability} from '../deviceDetailsModel';
import {Screenshot} from '../useScreenshot';

interface DeviceScreenProps {
    name: string,
    avail: Availability,
    links: DeviceRoutes,
    shot: Screenshot,
}

const Refresh: React.FC<{ onClick: () => void }> = ({onClick}) =>
    <Tooltip title="Take a new screenshot">
        <IconButton size="small" aria-label="Take a new screenshot" onClick={onClick}><RefreshRounded fontSize="small"/></IconButton>
    </Tooltip>;

const NoScreen: React.FC<{ text: string }> = ({text}) =>
    <Stack alignItems="center" justifyContent="center" spacing={1}
        sx={{height: 200, bgcolor: 'action.hover', borderRadius: 1, color: 'text.secondary', p: 2, textAlign: 'center'}}>
        <BrokenImageRounded/>
        <Typography variant="body2">{text}</Typography>
    </Stack>;

const Image: React.FC<Omit<DeviceScreenProps, 'avail'>> = ({name, links, shot}) =>
    <ButtonBase component={RouterLink} to={links.control()} aria-label="Open live control"
        sx={{display: 'block', borderRadius: 1, overflow: 'hidden', width: '100%'}}>
        <Box component="img" src={shot.src} alt={`Screen of ${name}`} onError={shot.markBroken}
             // a screenshot takes a moment; keep the space meanwhile
            sx={{display: 'block', width: '100%', minHeight: 200, maxHeight: 440, objectFit: 'contain', bgcolor: 'action.hover'}}/>
    </ButtonBase>;

const missingText = (avail: Availability): string =>
    avail.canControl ? 'The device returned no screenshot.' : `No screen while the device is ${avail.label.toLowerCase()}.`;

// DeviceScreen shows the screenshot of a device that is up, linking to live control
export const DeviceScreen: React.FC<DeviceScreenProps> = ({name, avail, links, shot}) =>
    <Section title="Screen" action={avail.canControl && <Refresh onClick={shot.refresh}/>}>
        {avail.canControl && !shot.broken
            ? <Image name={name} links={links} shot={shot}/>
            : <NoScreen text={missingText(avail)}/>}
        {avail.canControl && <Typography variant="caption" color="text.secondary">Click it for live control.</Typography>}
    </Section>;
