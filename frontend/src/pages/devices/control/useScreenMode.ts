import {useState} from 'react';
import {PlatformType} from '../../../types/platform.type.enum';
import {ScreenMode, VideoSupport} from '../../../types/device.control.video';

export interface ScreenModeState {
    mode: ScreenMode,
    // the user picks a mode; there is no automatic switch back after that
    pick: (mode: ScreenMode) => void,
    // live video isn't available: show screenshots
    showScreenshots: () => void,
    canVideo: boolean,
    isIOS: boolean,
}

// useScreenMode chooses how the screen is shown: live video where the
// browser can decode the device's stream, else screenshots, until the user
// or a failed video picks another. iOS video is MJPEG, which plays without
// WebCodecs (e.g. over plain http).
export const useScreenMode = (platform: PlatformType | undefined, support: VideoSupport): ScreenModeState => {
    const isIOS = platform === PlatformType.iOS;
    const canVideo = isIOS ? support.decodesMjpeg : support.decodesVideo;
    const [chosen, setChosen] = useState<ScreenMode | null>(null);
    const mode = chosen ?? (canVideo ? 'video' : 'screenshots');
    return {mode, pick: setChosen, showScreenshots: () => setChosen('screenshots'), canVideo, isIOS};
};
