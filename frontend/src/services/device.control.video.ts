import {DeviceVideoPlayer, mjpegDecodingAvailable, videoDecodingAvailable} from './device.video';
import {VideoSupport} from '../types/device.control.video';

// realVideoSupport asks the browser what it can decode and plays with
// DeviceVideoPlayer.
export const realVideoSupport = (): VideoSupport => ({
    decodesVideo: videoDecodingAvailable(),
    decodesMjpeg: mjpegDecodingAvailable(),
    createPlayer: (url, canvas, callbacks) =>
        new DeviceVideoPlayer(url, canvas, callbacks.onStatus, callbacks.onMeta, callbacks.onInputError),
});
