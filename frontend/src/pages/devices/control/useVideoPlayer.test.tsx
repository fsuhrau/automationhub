import React, {useEffect} from 'react';
import {describe, expect, it, vi} from 'vitest';
import {act, render, RenderResult} from '@testing-library/react';
import {VideoPlayerState, useVideoPlayer, VideoPlayerOptions} from './useVideoPlayer';
import {fakeSupport, FakeSupport} from './fakes';

// renders a canvas the hook can play on
const Harness: React.FC<{ options: VideoPlayerOptions, onState: (s: VideoPlayerState) => void }> = ({options, onState}) => {
    const {canvasRef, playerRef, status, meta} = useVideoPlayer(options);
    useEffect(() => onState({canvasRef, playerRef, status, meta}));
    return <canvas ref={canvasRef}/>;
};

interface Setup {
    support: FakeSupport,
    view: RenderResult,
    withUrl: (url: string) => void,
    current: () => VideoPlayerState,
}

const setup = (over: Partial<VideoPlayerOptions> = {}): Setup => {
    const support = fakeSupport();
    let state: VideoPlayerState | undefined;
    const options: VideoPlayerOptions = {enabled: true, url: 'ws://x/1', createPlayer: support.createPlayer, ...over};
    const capture = (s: VideoPlayerState): void => {
        state = s;
    };
    const view = render(<Harness options={options} onState={capture}/>);
    const withUrl = (url: string): void => view.rerender(<Harness options={{...options, url}} onState={capture}/>);
    return {support, view, withUrl, current: () => state as VideoPlayerState};
};

describe('useVideoPlayer', () => {
    it('plays the url on the canvas', () => {
        const {support} = setup();
        expect(support.players).toHaveLength(1);
        expect(support.players[0].url).toBe('ws://x/1');
        expect(support.players[0].canvas.tagName).toBe('CANVAS');
    });

    it('plays nothing while disabled', () => {
        const {support} = setup({enabled: false});
        expect(support.players).toHaveLength(0);
    });

    it('reports status and meta', () => {
        const onStatus = vi.fn();
        const {support, current} = setup({onStatus});
        const meta = {codec: 'h264', width: 1, height: 1, screenWidth: 1, screenHeight: 1};
        act(() => {
            support.players[0].callbacks.onStatus({state: 'playing'});
            support.players[0].callbacks.onMeta(meta);
        });
        expect(current().status).toEqual({state: 'playing'});
        expect(current().meta).toEqual(meta);
        expect(onStatus).toHaveBeenCalledWith({state: 'playing'});
    });

    it('takes status without a listener', () => {
        const {support, current} = setup();
        act(() => support.players[0].callbacks.onStatus({state: 'error', message: 'x'}));
        expect(current().status?.state).toBe('error');
    });
});

describe('useVideoPlayer input errors and lifecycle', () => {
    it('passes input errors on', () => {
        const onInputError = vi.fn();
        const {support} = setup({onInputError});
        support.players[0].callbacks.onInputError('bad');
        expect(onInputError).toHaveBeenCalledWith('bad');
    });

    it('takes input errors without a listener', () => {
        const {support} = setup();
        expect(() => support.players[0].callbacks.onInputError('bad')).not.toThrow();
    });

    it('exposes the player while it plays and closes it on unmount', () => {
        const {support, view, current} = setup();
        expect(current().playerRef.current).toBe(support.players[0]);
        view.unmount();
        expect(support.players[0].closed).toBe(true);
    });

    it('starts a new player when the url changes and forgets the old stream meta', () => {
        const {support, withUrl, current} = setup();
        act(() => support.players[0].callbacks.onMeta({codec: 'h264', width: 1, height: 1, screenWidth: 1, screenHeight: 1}));
        withUrl('ws://x/2');
        expect(support.players.map(p => [p.url, p.closed])).toEqual([['ws://x/1', true], ['ws://x/2', false]]);
        expect(current().meta).toBeNull();
    });
});
