import React from 'react';
import {Box, CircularProgress, Stack, Typography} from '@mui/material';
import {BrokenImageRounded} from '@mui/icons-material';
import {streamUrl} from '../services/device.control.service';
import {VideoStatus} from '../services/device.video';
import {useLiveScreen} from '../pages/devices/control/useLiveScreen';
import {VideoSupport} from '../types/device.control.video';
import {PlatformType} from '../types/platform.type.enum';

const SCREENSHOT_FPS = 2;

interface DeviceLiveScreenProps {
    projectId: string,
    deviceId: number,
    platformType?: PlatformType,
    label: string,
    maxHeight?: number | string,
    // what the browser can decode and how to play it; tests fake it
    support?: VideoSupport,
}

const NoScreen: React.FC = () => (
    <Stack alignItems="center" justifyContent="center" spacing={1}
        sx={{height: 200, bgcolor: 'action.hover', borderRadius: 1, color: 'text.secondary', p: 2, textAlign: 'center'}}>
        <BrokenImageRounded/>
        <Typography variant="body2">The device returned no screen.</Typography>
    </Stack>
);

const Connecting: React.FC<{ status: VideoStatus | null }> = ({status}) => (
    <Stack alignItems="center" justifyContent="center" spacing={1}
        sx={{position: 'absolute', inset: 0, color: 'common.white', lineHeight: 1.5, p: 2, textAlign: 'center'}}>
        <CircularProgress size={20} color="inherit"/>
        <Typography variant="caption">
            {status?.state === 'error' ? `Reconnecting… ${status.message ?? ''}` : 'Connecting to the screen…'}
        </Typography>
    </Stack>
);

// DeviceLiveScreen shows the screen of a device read-only: the live video
// where the browser and device support it, else a screenshot stream. Watching
// needs no lease, so it also works while a test run holds the device.
const DeviceLiveScreen: React.FC<DeviceLiveScreenProps> = ({label, maxHeight = 480, ...options}) => {
    const {projectId, deviceId} = options;
    const {mode, canvasRef, status, broken, onBroken} = useLiveScreen(options);
    const style: React.CSSProperties = {display: 'block', maxWidth: '100%', maxHeight, margin: '0 auto'};

    if (broken) return <NoScreen/>;
    return (
        <Box sx={{position: 'relative', borderRadius: 1, overflow: 'hidden', bgcolor: '#000', lineHeight: 0}}>
            {mode === 'video'
                ? <canvas ref={canvasRef} width={360} height={640} role="img" aria-label={`Live screen of ${label}`} style={style}/>
                : <img key={deviceId} src={streamUrl(projectId, deviceId, SCREENSHOT_FPS)} alt={`Live screen of ${label}`}
                        draggable={false} onError={onBroken} style={style}/>}
            {mode === 'video' && status?.state !== 'playing' && <Connecting status={status}/>}
        </Box>
    );
};

export default DeviceLiveScreen;
