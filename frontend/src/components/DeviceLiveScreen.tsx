import React, {useEffect, useMemo, useRef, useState} from 'react';
import {Box, CircularProgress, Stack, Typography} from '@mui/material';
import {BrokenImageRounded} from '@mui/icons-material';
import {streamUrl} from '../services/device.control.service';
import {DeviceVideoPlayer, mjpegDecodingAvailable, videoDecodingAvailable, VideoStatus, videoUrl} from '../services/device.video';
import {PlatformType} from '../types/platform.type.enum';

// smaller than on the control page: the screen is shown in a side card
const VIDEO_OPTIONS = {maxSize: 960, maxFps: 20};
const SCREENSHOT_FPS = 2;

interface DeviceLiveScreenProps {
    projectId: string,
    deviceId: number,
    platformType?: PlatformType,
    label: string,
    maxHeight?: number | string,
}

// DeviceLiveScreen shows the screen of a device read-only: the live video
// where the browser and device support it, else a screenshot stream. Watching
// needs no lease, so it also works while a test run holds the device.
const DeviceLiveScreen: React.FC<DeviceLiveScreenProps> = ({projectId, deviceId, platformType, label, maxHeight = 480}) => {
    // iOS video is MJPEG, which plays without WebCodecs (e.g. over plain http)
    const canVideo = useMemo(() => platformType === PlatformType.iOS ? mjpegDecodingAvailable() : videoDecodingAvailable(), [platformType]);
    const [mode, setMode] = useState<'video' | 'screenshots'>(canVideo ? 'video' : 'screenshots');
    const [status, setStatus] = useState<VideoStatus | null>(null);
    const [imageBroken, setImageBroken] = useState(false);
    const canvasRef = useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        setMode(canVideo ? 'video' : 'screenshots');
        setImageBroken(false);
    }, [deviceId, canVideo]);

    useEffect(() => {
        if (mode !== 'video' || !canvasRef.current) return;
        setStatus(null);
        const player = new DeviceVideoPlayer(
            videoUrl(projectId, deviceId, VIDEO_OPTIONS),
            canvasRef.current,
            s => {
                setStatus(s);
                // e.g. a simulator or an editor: use the screenshot stream
                if (s.state === 'unsupported') setMode('screenshots');
            },
            () => undefined,
        );
        return () => player.close();
    }, [mode, projectId, deviceId]);

    const style: React.CSSProperties = {display: 'block', maxWidth: '100%', maxHeight, margin: '0 auto'};

    if (mode === 'screenshots' && imageBroken) {
        return <Stack alignItems="center" justifyContent="center" spacing={1}
                      sx={{height: 200, bgcolor: 'action.hover', borderRadius: 1, color: 'text.secondary', p: 2, textAlign: 'center'}}>
            <BrokenImageRounded/>
            <Typography variant="body2">The device returned no screen.</Typography>
        </Stack>;
    }

    const waiting = mode === 'video' && status?.state !== 'playing';
    return <Box sx={{position: 'relative', borderRadius: 1, overflow: 'hidden', bgcolor: '#000', lineHeight: 0}}>
        {mode === 'video'
            ? <canvas ref={canvasRef} width={360} height={640} role="img" aria-label={`Live screen of ${label}`} style={style}/>
            : <img key={deviceId} src={streamUrl(projectId, deviceId, SCREENSHOT_FPS)} alt={`Live screen of ${label}`}
                   draggable={false} onError={() => setImageBroken(true)} style={style}/>}
        {waiting && <Stack alignItems="center" justifyContent="center" spacing={1}
                           sx={{position: 'absolute', inset: 0, color: 'common.white', lineHeight: 1.5, p: 2, textAlign: 'center'}}>
            <CircularProgress size={20} color="inherit"/>
            <Typography variant="caption">
                {status?.state === 'error' ? `Reconnecting… ${status.message ?? ''}` : 'Connecting to the screen…'}
            </Typography>
        </Stack>}
    </Box>;
};

export default DeviceLiveScreen;
