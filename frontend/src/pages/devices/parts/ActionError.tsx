import React from 'react';
import {Alert} from '@mui/material';

// ActionError tells that an action failed; it shows nothing without a message
export const ActionError: React.FC<{ message: string | null }> = ({message}) =>
    message ? <Alert severity="error" sx={{mb: 2}}>{message}</Alert> : null;
