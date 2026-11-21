import React from 'react';
import {Alert, Box, Button} from '@mui/material';
import {ArrowBackRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';

interface LoadFailureProps {
    message: string,
    onRetry: () => void,
    // the list to go back to
    back: string,
}

export const LoadFailure: React.FC<LoadFailureProps> = ({message, onRetry, back}) =>
    <Box sx={{mt: 2}}>
        <Alert severity="error" action={<Button color="inherit" size="small" onClick={onRetry}>Retry</Button>}>{message}</Alert>
        <Button sx={{mt: 2}} startIcon={<ArrowBackRounded/>} component={RouterLink} to={back}>All bundles</Button>
    </Box>;
