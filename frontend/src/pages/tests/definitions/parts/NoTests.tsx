import React from 'react';
import {Button, Stack, Typography} from '@mui/material';
import {AddRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';

// NoTests is the empty state of the tests list
export const NoTests: React.FC<{ newTestUrl: string }> = ({newTestUrl}) =>
    <Stack alignItems="center" spacing={1.5} sx={{py: 6, textAlign: 'center'}}>
        <Typography variant="subtitle1">No tests yet</Typography>
        <Typography variant="body2" color="text.secondary">
            A test says what to run (for Unity: all tests, a category or picked tests) and on which devices.
        </Typography>
        <Button variant="outlined" startIcon={<AddRounded/>} component={RouterLink} to={newTestUrl}>Create a test</Button>
    </Stack>;
