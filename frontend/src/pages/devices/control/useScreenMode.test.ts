import {describe, expect, it} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {PlatformType} from '../../../types/platform.type.enum';
import {fakeSupport} from './fakes';
import {ScreenModeState, useScreenMode} from './useScreenMode';

const modeOf = (platform: PlatformType | undefined, over = {}): RenderHookResult<ScreenModeState, unknown> =>
    renderHook(() => useScreenMode(platform, fakeSupport(over)));

describe('useScreenMode', () => {
    it('plays video where the browser decodes it', () => {
        const {result} = modeOf(PlatformType.Android);
        expect(result.current).toMatchObject({mode: 'video', canVideo: true, isIOS: false});
    });

    it('shows screenshots where it does not', () => {
        const {result} = modeOf(PlatformType.Android, {decodesVideo: false});
        expect(result.current).toMatchObject({mode: 'screenshots', canVideo: false});
    });

    it('plays iOS video as MJPEG even without WebCodecs', () => {
        const {result} = modeOf(PlatformType.iOS, {decodesVideo: false});
        expect(result.current).toMatchObject({mode: 'video', canVideo: true, isIOS: true});
    });

    it('shows screenshots on iOS without MJPEG decoding', () => {
        const {result} = modeOf(PlatformType.iOS, {decodesMjpeg: false});
        expect(result.current.mode).toBe('screenshots');
    });

    it('keeps the mode the user picks', () => {
        const {result} = modeOf(PlatformType.Android);
        act(() => result.current.pick('off'));
        expect(result.current.mode).toBe('off');
    });

    it('falls back to screenshots', () => {
        const {result} = modeOf(PlatformType.Android);
        act(() => result.current.showScreenshots());
        expect(result.current.mode).toBe('screenshots');
    });
});
