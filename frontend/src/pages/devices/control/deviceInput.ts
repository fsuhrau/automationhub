import {DeviceControlApi} from '../../../types/device.control.api';
import {VIDEO_TEXT_MAX_BYTES} from '../../../services/device.video';
import {VideoPlayerLike} from '../../../types/device.control.video';
import {Runner} from './useGuardedAction';

export interface InputDeps {
    lease: string | null,
    service: DeviceControlApi,
    run: Runner,
    liveInput: () => VideoPlayerLike | null,
}

export interface DeviceInput {
    sendKey: (key: string) => void,
    // done runs once the text is on its way
    typeText: (text: string, done?: () => void) => void,
}

// createDeviceInput sends keys and text to the device: over the video
// socket when it can take input, else through the REST API. Without a lease
// nothing is sent.
export const createDeviceInput = ({lease, service, run, liveInput}: InputDeps): DeviceInput => ({
    sendKey: key => {
        const player = liveInput();
        if (player) {
            player.sendInput({kind: 'key', action: 'press', key}, lease);
        } else if (lease) {
            run(() => service.key(lease, key));
        }
    },
    typeText: (text, done) => {
        const player = liveInput();
        if (player && new TextEncoder().encode(text).length <= VIDEO_TEXT_MAX_BYTES) {
            player.sendInput({kind: 'text', text}, lease);
            done?.();
        } else if (lease) {
            run(async () => {
                await service.text(lease, text);
                done?.();
            });
        }
    },
});
