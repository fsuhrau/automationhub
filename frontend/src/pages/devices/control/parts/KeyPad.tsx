import React, {ReactNode, useState} from 'react';
import {IconButton, Menu, MenuItem, Stack, Tooltip} from '@mui/material';
import {
    ArrowBackIosNewRounded,
    CircleOutlined,
    CropSquareRounded,
    MoreHorizRounded,
    VolumeDownRounded,
    VolumeUpRounded,
} from '@mui/icons-material';

interface KeyProps {
    disabled: boolean,
    onKey: (key: string) => void,
}

// the keys of a phone's navigation bar, then the rarer ones
const NAV_KEYS: { key: string, label: string, icon: ReactNode }[] = [
    {key: 'back', label: 'Back', icon: <ArrowBackIosNewRounded fontSize="small"/>},
    {key: 'home', label: 'Home', icon: <CircleOutlined fontSize="small"/>},
    {key: 'app_switch', label: 'Recent apps', icon: <CropSquareRounded fontSize="small"/>},
];
const MORE_KEYS = [
    {key: 'enter', label: 'Enter'},
    {key: 'delete', label: 'Delete (backspace)'},
    {key: 'tab', label: 'Tab'},
    {key: 'menu', label: 'Menu'},
    {key: 'search', label: 'Search'},
    {key: 'power', label: 'Power (turns the screen off or on)'},
];

const MoreKeys: React.FC<KeyProps> = ({disabled, onKey}) => {
    const [anchor, setAnchor] = useState<HTMLElement | null>(null);
    return (
        <>
            <Tooltip title="More keys"><span>
                <IconButton size="small" aria-label="More keys" disabled={disabled}
                    onClick={e => setAnchor(e.currentTarget)}><MoreHorizRounded fontSize="small"/></IconButton>
            </span></Tooltip>
            <Menu anchorEl={anchor} open={!!anchor} onClose={() => setAnchor(null)}>
                {MORE_KEYS.map(k => <MenuItem key={k.key} onClick={() => {
                    setAnchor(null);
                    onKey(k.key);
                }}>{k.label}</MenuItem>)}
            </Menu>
        </>
    );
};

// the navigation bar below the screen, where a phone has it in portrait
export const NavKeys: React.FC<KeyProps> = ({disabled, onKey}) => (
    <Stack direction="row" alignItems="center" justifyContent="center" spacing={0.5} sx={{mt: 1}}
        role="group" aria-label="Navigation keys">
        {NAV_KEYS.map(k => <Tooltip key={k.key} title={k.label}><span>
            <IconButton aria-label={k.label} disabled={disabled} onClick={() => onKey(k.key)} sx={{mx: 1}}>{k.icon}</IconButton>
        </span></Tooltip>)}
        <MoreKeys disabled={disabled} onKey={onKey}/>
    </Stack>
);

// volume on the right edge
export const VolumeKeys: React.FC<KeyProps> = ({disabled, onKey}) => (
    <Stack spacing={0.5} sx={{mt: 10}} role="group" aria-label="Volume keys">
        <Tooltip title="Volume up" placement="right"><span>
            <IconButton size="small" aria-label="Volume up" disabled={disabled}
                onClick={() => onKey('volume_up')}><VolumeUpRounded fontSize="small"/></IconButton>
        </span></Tooltip>
        <Tooltip title="Volume down" placement="right"><span>
            <IconButton size="small" aria-label="Volume down" disabled={disabled}
                onClick={() => onKey('volume_down')}><VolumeDownRounded fontSize="small"/></IconButton>
        </span></Tooltip>
    </Stack>
);
