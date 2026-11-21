import React from 'react';
import {Alert} from '@mui/material';
import {FormErrors} from '../../testEditModel';

// ErrorSummary lists what keeps the form from saving, once a save was tried
export const ErrorSummary: React.FC<{ errors: FormErrors, tried: boolean }> = ({errors, tried}) => {
    const messages = Object.values(errors);
    if (!tried || messages.length === 0) return null;
    return <Alert severity="error" sx={{mb: 2}}>{messages.join(' ')}</Alert>;
};
