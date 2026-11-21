import React from 'react';
import {Chip, Stack} from '@mui/material';
import {ErrorOutlineRounded, WarningAmberRounded} from '@mui/icons-material';
import {formatDateTime} from '../../../components/DateTime';
import {expiry} from '../bundleDetailModel';

const DAY = 86400000;

// ExpiryText is a date that says when it has passed or is near
export const ExpiryText: React.FC<{ date: string, now: number }> = ({date, now}) => {
    const state = expiry(date, now);
    return <Stack direction="row" spacing={0.75} alignItems="center" component="span">
        <span>{formatDateTime(date)}</span>
        {state === 'expired' && <Chip size="small" color="error" icon={<ErrorOutlineRounded/>} label="expired"/>}
        {state === 'soon' && <Chip size="small" color="warning" icon={<WarningAmberRounded/>}
            label={`${Math.ceil((new Date(date).valueOf() - now) / DAY)} days left`}/>}
    </Stack>;
};
