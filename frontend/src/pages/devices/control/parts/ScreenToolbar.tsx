import React from 'react';
import {Box, Chip, CircularProgress, IconButton, Stack, ToggleButton, ToggleButtonGroup, Tooltip} from '@mui/material';
import {CameraAltRounded, RefreshRounded} from '@mui/icons-material';
import {ScreenMode} from '../../../../types/device.control.video';
import {ScreenState} from '../useScreen';

const videoHint = (canVideo: boolean, isIOS: boolean): string => {
    if (!canVideo) return 'Video needs a secure context: open the hub over HTTPS or localhost';
    return isIOS
        ? 'Live video from WebDriverAgent (starts it when needed); taps and swipes go through the control API'
        : 'Live video with fast input (Android)';
};

// as tall as the icon buttons beside it; the theme's toggles are form-sized
const toggleSx = {'& .MuiToggleButton-root': {py: 0.5, px: 1.25, height: 36, fontSize: 12, lineHeight: 1}};

const ModeToggle: React.FC<{ screen: ScreenState }> = ({screen}) => {
    const {mode, pick, canVideo, isIOS} = screen.modes;
    return (
        <ToggleButtonGroup size="small" exclusive={true} value={mode} aria-label="Screen" sx={toggleSx}
            onChange={(_, v: ScreenMode | null) => v && pick(v)}>
            <Tooltip title={videoHint(canVideo, isIOS)}>
                <span><ToggleButton value="video" disabled={!canVideo}>Video</ToggleButton></span>
            </Tooltip>
            <Tooltip title="A screenshot every half second; works on every device">
                <ToggleButton value="screenshots">Screenshots</ToggleButton>
            </Tooltip>
            <Tooltip title="No updates; refresh by hand">
                <ToggleButton value="off">Paused</ToggleButton>
            </Tooltip>
        </ToggleButtonGroup>
    );
};

const VideoChips: React.FC<{ screen: ScreenState }> = ({screen}) => {
    const status = screen.modes.mode === 'video' ? screen.video.status : null;
    return (
        <>
            {status?.state === 'connecting' && <Chip size="small"
                label={screen.modes.isIOS ? 'Connecting… (starting WebDriverAgent can take a minute)' : 'Connecting…'}/>}
            {status?.state === 'error' && <Chip size="small" color="error" label={status.message ?? 'Video error'}/>}
            {screen.live && <Tooltip title="Touch, drag, scroll and keys go straight through the video connection">
                <Chip size="small" color="success" variant="outlined" label="Live"/>
            </Tooltip>}
        </>
    );
};

const ScreenToolbar: React.FC<{ screen: ScreenState, busy: boolean }> = ({screen, busy}) => (
    <Stack direction="row" spacing={1} alignItems="center" useFlexGap={true} flexWrap="wrap" sx={{mb: 1}}>
        <ModeToggle screen={screen}/>
        {screen.modes.mode === 'off' && <Tooltip title="Refresh the screen">
            <IconButton size="small" aria-label="Refresh the screen" onClick={screen.source.refresh}><RefreshRounded/></IconButton>
        </Tooltip>}
        <Tooltip title="Save a screenshot">
            <IconButton size="small" aria-label="Save a screenshot" onClick={screen.saveScreenshot} disabled={busy}>
                <CameraAltRounded/>
            </IconButton>
        </Tooltip>
        <Box sx={{flexGrow: 1}}/>
        {busy && <CircularProgress size={18}/>}
        <VideoChips screen={screen}/>
    </Stack>
);

export default ScreenToolbar;
