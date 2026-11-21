import {useState} from 'react';
import {screenshotUrl, streamUrl} from '../../../services/device.control.service';
import {DeviceTarget} from '../../../types/device.control.api';
import {ScreenMode} from '../../../types/device.control.video';

const STREAM_FPS = 2;
// a screenshot after a gesture shows its result; the device needs a moment
export const SNAPSHOT_DELAY_MS = 700;

export interface ScreenSource {
    // the image to show in the screenshot modes
    src: string,
    // take a new screenshot (paused mode)
    refresh: () => void,
    // call after input: in paused mode the screen is refreshed shortly
    afterInput: () => void,
}

// useScreenSource says where the screen image comes from when there is no
// video: the screenshot stream, or in paused mode one screenshot at a time.
export const useScreenSource = ({projectId, deviceId}: DeviceTarget, mode: ScreenMode): ScreenSource => {
    const [snapshot, setSnapshot] = useState(() => screenshotUrl(projectId, deviceId));
    const refresh = (): void => setSnapshot(screenshotUrl(projectId, deviceId));
    return {
        src: mode === 'screenshots' ? streamUrl(projectId, deviceId, STREAM_FPS) : snapshot,
        refresh,
        afterInput: () => {
            if (mode === 'off') setTimeout(refresh, SNAPSHOT_DELAY_MS);
        },
    };
};
