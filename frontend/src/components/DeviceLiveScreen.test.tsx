import type {ReactElement} from 'react';
import {describe, expect, it} from 'vitest';
import {act, fireEvent, render, screen} from '@testing-library/react';
import {fakeSupport, FakeSupport} from '../pages/devices/control/fakes';
import {PlatformType} from '../types/platform.type.enum';
import DeviceLiveScreen from './DeviceLiveScreen';

const show = (support: FakeSupport, platformType = PlatformType.Android, deviceId = 7): ReactElement =>
    <DeviceLiveScreen projectId="p1" deviceId={deviceId} platformType={platformType} label="Pixel" support={support}/>;

describe('DeviceLiveScreen with video', () => {
    it('plays the video of the device on a canvas', () => {
        const support = fakeSupport();
        render(show(support));
        expect(screen.getByRole('img', {name: 'Live screen of Pixel'})).toBeInTheDocument();
        expect(support.players[0].url).toContain('/api/p1/device/7/video?maxSize=960&maxFps=20');
    });

    it('says it connects until the video plays', () => {
        const support = fakeSupport();
        render(show(support));
        expect(screen.getByText('Connecting to the screen…')).toBeInTheDocument();
        act(() => support.players[0].callbacks.onStatus({state: 'error', message: 'closed'}));
        expect(screen.getByText('Reconnecting… closed')).toBeInTheDocument();
        act(() => support.players[0].callbacks.onStatus({state: 'playing'}));
        expect(screen.queryByText(/Connecting|Reconnecting/)).not.toBeInTheDocument();
    });

    it('shows the screenshot stream when the device has no video', () => {
        const support = fakeSupport();
        render(show(support));
        act(() => support.players[0].callbacks.onStatus({state: 'unsupported'}));
        expect(screen.getByAltText('Live screen of Pixel')).toHaveAttribute('src', '/api/p1/device/7/stream?fps=2');
        expect(support.players[0].closed).toBe(true);
    });

    it('tries video again for another device', () => {
        const support = fakeSupport();
        const view = render(show(support));
        act(() => support.players[0].callbacks.onStatus({state: 'unsupported'}));
        view.rerender(show(support, PlatformType.Android, 8));
        expect(support.players.map(p => p.url.includes('/device/8/'))).toEqual([false, true]);
    });
});

describe('DeviceLiveScreen without video', () => {
    it('streams screenshots where the browser cannot decode video', () => {
        const support = fakeSupport({decodesVideo: false});
        render(show(support));
        expect(screen.getByAltText('Live screen of Pixel')).toBeInTheDocument();
        expect(support.players).toHaveLength(0);
    });

    it('plays iOS video as MJPEG even without WebCodecs', () => {
        const support = fakeSupport({decodesVideo: false});
        render(show(support, PlatformType.iOS));
        expect(support.players).toHaveLength(1);
    });

    it('says when the device returns no screen, and tries again for another device', () => {
        const support = fakeSupport({decodesVideo: false});
        const view = render(show(support));
        fireEvent.error(screen.getByAltText('Live screen of Pixel'));
        expect(screen.getByText('The device returned no screen.')).toBeInTheDocument();
        view.rerender(show(support, PlatformType.Android, 8));
        expect(screen.getByAltText('Live screen of Pixel')).toBeInTheDocument();
    });

    it('asks the browser when no support is given', () => {
        render(<DeviceLiveScreen projectId="p1" deviceId={7} label="Pixel"/>);
        expect(screen.getByAltText('Live screen of Pixel')).toBeInTheDocument();
    });
});
