import React from 'react';
import {Box, Paper, Typography} from '@mui/material';
import {ControlSession} from '../useControlSession';
import {TakeControlButton} from './LeaseButtons';

// covers the screen while the user has no control: says why and offers it
const LeaseOverlay: React.FC<{ session: ControlSession }> = ({session}) => {
    const {status, busy} = session;
    return (
        <Box sx={{
            position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center',
            bgcolor: 'rgba(0,0,0,0.45)', lineHeight: 1.5, p: 2,
        }}>
            <Paper sx={{p: 2, maxWidth: 260, textAlign: 'center'}}>
                <Typography variant="subtitle2" gutterBottom={true}>{status.label}</Typography>
                <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>
                    {status.taken
                        ? 'You can watch the screen. Control is possible once the device is free.'
                        : 'Take control to tap, swipe, scroll and type on the device.'}
                </Typography>
                {!status.taken && <TakeControlButton size="small" onClick={session.onLock}
                    disabled={busy || !!session.lease.status?.locked}/>}
            </Paper>
        </Box>
    );
};

export default LeaseOverlay;
