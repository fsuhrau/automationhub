import * as React from 'react';
import {Link as RouterLink} from 'react-router-dom';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import SwapHorizRoundedIcon from '@mui/icons-material/SwapHorizRounded';
import {routes} from '../../lib/routes';

interface SwitchProjectButtonProps {
    onNavigate?: () => void,
    // where the tooltip opens; the open menu shows it below
    placement?: 'right',
}

const SwitchProjectButton: React.FC<SwitchProjectButtonProps> = ({onNavigate, placement}) => (
    <Tooltip title="Switch project" placement={placement}>
        <IconButton size="small" component={RouterLink} to={routes.projects()} aria-label="Switch project" onClick={onNavigate}>
            <SwapHorizRoundedIcon fontSize="small"/>
        </IconButton>
    </Tooltip>
);

export default SwitchProjectButton;
