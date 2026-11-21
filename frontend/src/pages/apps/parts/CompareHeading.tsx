import React from 'react';
import {Box, Button, Stack, Typography} from '@mui/material';
import {ArrowBackRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';

export const CompareHeading: React.FC<{ appName?: string, back: string }> = ({appName, back}) =>
    <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
        <Box sx={{minWidth: 0}}>
            <Typography component="h1" variant="h5">Compare bundles</Typography>
            <Typography variant="body2" color="text.secondary">
                What changed from one build of {appName ?? 'the app'} to another. Archived bundles keep their details and compare too.
            </Typography>
        </Box>
        <Button sx={{ml: {md: 'auto'}}} startIcon={<ArrowBackRounded/>} component={RouterLink} to={back}>All bundles</Button>
    </Stack>;
