import React from 'react';
import {Chip, Typography} from '@mui/material';
import {Change} from '../../../types/bundle.compare';

export const CHANGE_COLOR: Record<Exclude<Change, 'same'>, 'success' | 'error' | 'warning'> = {
    changed: 'warning', added: 'success', removed: 'error',
};

// edgeColor marks a row by how it changed
export const edgeColor = (change: Change): string => change === 'same' ? 'transparent' : `${CHANGE_COLOR[change]}.main`;

export const ChangeMark: React.FC<{ change: Change }> = ({change}) =>
    change === 'same'
        ? <Typography variant="caption" color="text.disabled">same</Typography>
        : <Chip size="small" color={CHANGE_COLOR[change]} variant="outlined" label={change}/>;
