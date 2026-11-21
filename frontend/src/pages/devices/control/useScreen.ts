import {Point} from '../../../types/device.control.api';
import {VideoSupport} from '../../../types/device.control.video';
import {screenshotUrl} from '../../../services/device.control.service';
import {createDeviceInput} from './deviceInput';
import {fetchScreenshot, saveBlob, screenshotName} from './saveScreenshot';
import {ControlSession} from './useControlSession';
import {PointerHandlers, usePointerInput} from './usePointerInput';
import {ScreenModeState, useScreenMode} from './useScreenMode';
import {ScreenSource, useScreenSource} from './useScreenSource';
import {Typing, useTyping} from './useTyping';
import {useInputError} from './useInputError';
import {VideoInput, useVideoInput} from './useVideoInput';
import {DeviceInput} from './deviceInput';

export interface ScreenState {
    modes: ScreenModeState,
    video: VideoInput,
    source: ScreenSource,
    input: DeviceInput,
    typing: Typing,
    pointer: PointerHandlers,
    // touches, keys and scrolling go straight through the video connection
    live: boolean,
    saveScreenshot: () => void,
}

export interface ScreenOptions {
    support: VideoSupport,
    // set while picking an element on the screen
    onPick: ((el: HTMLElement, at: Point) => void) | null,
}

// useScreen is the screen of the device and all input on it: how it is
// shown, keys and text, taps, swipes and live touches.
export const useScreen = (session: ControlSession, {support, onPick}: ScreenOptions): ScreenState => {
    const {target, service, run, lease: {token, refresh}} = session;
    const modes = useScreenMode(session.device?.platformType, support);
    const onInputError = useInputError(session.report, refresh);
    const video = useVideoInput({
        mode: modes.mode, target, support, lease: token, onInputError,
        onUnsupported: notice => {
            modes.showScreenshots();
            if (notice) session.report(notice);
        },
    });
    const source = useScreenSource(target, modes.mode);
    const input = createDeviceInput({lease: token, service, run, liveInput: video.liveInput});
    const typing = useTyping(input, token !== null);
    const pointer = usePointerInput({
        lease: token, meta: video.meta, service, run, liveInput: video.liveInput, onPick, onGesture: source.afterInput,
    });
    const live = modes.mode === 'video' && video.status?.state === 'playing' && token !== null && !video.meta?.viewOnly;
    const saveScreenshot = (): void => {
        const playing = modes.mode === 'video' && video.status?.state === 'playing';
        run(async () => saveBlob(
            await fetchScreenshot({canvas: playing ? video.canvasRef.current : null, url: screenshotUrl(target.projectId, target.deviceId)}),
            screenshotName(session.name, Date.now()),
        ));
    };
    return {modes, video, source, input, typing, pointer, live, saveScreenshot};
};
