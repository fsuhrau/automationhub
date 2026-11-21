import {useCallback, useState} from 'react';
import {screenshotUrl} from '../../services/device.control.service';

export interface Screenshot {
    src: string,
    broken: boolean,
    refresh: () => void,
    markBroken: () => void,
}

// useScreenshot is the current screenshot of a device; refresh takes a new one
export function useScreenshot(projectId: string, deviceId: number): Screenshot {
    const [src, setSrc] = useState(() => screenshotUrl(projectId, deviceId));
    const [broken, setBroken] = useState(false);
    const refresh = useCallback(() => {
        setBroken(false);
        setSrc(screenshotUrl(projectId, deviceId));
    }, [projectId, deviceId]);
    const markBroken = useCallback(() => setBroken(true), []);
    return {src, broken, refresh, markBroken};
}
