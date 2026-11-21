import React from 'react';
import {Button, Tooltip} from '@mui/material';
import {LockOpenRounded, TouchAppRounded} from '@mui/icons-material';

interface LeaseButtonProps {
    onClick: () => void,
    disabled: boolean,
}

export const TakeControlButton: React.FC<LeaseButtonProps & { size?: 'small' }> = ({onClick, disabled, size}) => (
    <Button variant="contained" size={size} startIcon={<TouchAppRounded/>} onClick={onClick} disabled={disabled}>
        Take control
    </Button>
);

export const ReleaseButton: React.FC<LeaseButtonProps> = ({onClick, disabled}) => (
    <Tooltip title="Give the device back so tests and others can use it">
        <Button variant="outlined" startIcon={<LockOpenRounded/>} onClick={onClick} disabled={disabled}>Release</Button>
    </Tooltip>
);
