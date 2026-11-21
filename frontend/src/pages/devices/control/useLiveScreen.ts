import {RefObject, useMemo, useState} from 'react';
import {realVideoSupport} from '../../../services/device.control.video';
import {VideoStatus, videoUrl} from '../../../services/device.video';
import {VideoSupport} from '../../../types/device.control.video';
import {PlatformType} from '../../../types/platform.type.enum';
import {useVideoPlayer} from './useVideoPlayer';

// smaller than on the control page: the screen is shown in a side card
const VIDEO_OPTIONS = {maxSize: 960, maxFps: 20};

export interface LiveScreenOptions {
    projectId: string,
    deviceId: number,
    platformType?: PlatformType,
    support?: VideoSupport,
}

export interface LiveScreen {
    mode: 'video' | 'screenshots',
    canvasRef: RefObject<HTMLCanvasElement | null>,
    status: VideoStatus | null,
    // the screenshot stream returned no image
    broken: boolean,
    onBroken: () => void,
}

// useLiveScreen decides between live video and the screenshot stream for a
// read-only view of a device: video where the browser and device support it,
// screenshots else, and after the device said video is unsupported.
export const useLiveScreen = ({projectId, deviceId, platformType, support}: LiveScreenOptions): LiveScreen => {
    const video = useMemo(() => support ?? realVideoSupport(), [support]);
    // iOS video is MJPEG, which plays without WebCodecs (e.g. over plain http)
    const canVideo = platformType === PlatformType.iOS ? video.decodesMjpeg : video.decodesVideo;
    // the device the video fell back to screenshots for, and the one whose image failed
    const [fellBack, setFellBack] = useState<number | null>(null);
    const [brokenFor, setBrokenFor] = useState<number | null>(null);
    const mode = canVideo && fellBack !== deviceId ? 'video' : 'screenshots';
    const {canvasRef, status} = useVideoPlayer({
        enabled: mode === 'video', createPlayer: video.createPlayer, url: videoUrl(projectId, deviceId, VIDEO_OPTIONS),
        // e.g. a simulator or an editor: use the screenshot stream
        onStatus: s => s.state === 'unsupported' && setFellBack(deviceId),
    });
    return {mode, canvasRef, status, broken: mode === 'screenshots' && brokenFor === deviceId, onBroken: () => setBrokenFor(deviceId)};
};
