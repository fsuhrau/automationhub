import React from 'react';
import {act, fireEvent, render, screen, waitFor} from '@testing-library/react';
import {afterEach, beforeEach, expect, it, vi} from 'vitest';
import ReplayVideo from './ReplayVideo';
import {FrameIo} from '../pages/tests/replay/video/frameLoader';
import {ReplaySegment, segmentsOf} from '../pages/tests/replayModel';
import {recording, T0} from '../pages/tests/replay/testFixtures';

const segmentsOfCodec = (over: Parameters<typeof recording>[0] = {}): ReplaySegment[] => segmentsOf([recording(over)]).get(5) ?? [];

const props = (segments: ReplaySegment[], over: Partial<React.ComponentProps<typeof ReplayVideo>> = {}): React.ComponentProps<typeof ReplayVideo> =>
    ({segments, position: T0 + 1000, playing: false, rate: 1, label: 'Screen', ...over});

it('ReplayVideo messages: says when the device has no recording', () => {
    render(<ReplayVideo {...props([])}/>);
    expect(screen.getByText('This device has no recording.')).toBeInTheDocument();
});

it('ReplayVideo messages: says when the time is between recordings', () => {
    render(<ReplayVideo {...props(segmentsOfCodec(), {position: T0 + 20000})}/>);
    expect(screen.getByText('No video at this point of the run.')).toBeInTheDocument();
});

it('ReplayVideo messages: keeps the aspect ratio of the next recording before it starts', () => {
    const later = segmentsOfCodec({startedAt: new Date(T0 + 5000).toISOString(), width: 300, height: 100});
    const {container} = render(<ReplayVideo {...props(later, {position: T0})}/>);
    expect(container.firstChild).toHaveStyle({aspectRatio: '300 / 100'});
});

let currentTime = 0;
let readyState = 4;
const play = vi.fn();
const pause = vi.fn();

beforeEach(() => {
    currentTime = 0;
    readyState = 4;
    play.mockReset().mockResolvedValue(undefined);
    pause.mockReset();
    vi.spyOn(HTMLMediaElement.prototype, 'play').mockImplementation(play);
    vi.spyOn(HTMLMediaElement.prototype, 'pause').mockImplementation(pause);
    vi.spyOn(HTMLMediaElement.prototype, 'readyState', 'get').mockImplementation(() => readyState);
    vi.spyOn(HTMLMediaElement.prototype, 'currentTime', 'get').mockImplementation(() => currentTime);
    vi.spyOn(HTMLMediaElement.prototype, 'currentTime', 'set').mockImplementation(t => {
        currentTime = t;
    });
});
afterEach(() => {
    vi.restoreAllMocks();
});

const video = (): HTMLVideoElement => screen.getByLabelText('Screen') as HTMLVideoElement;

it('ReplayVideo with an mp4 recording: shows the file with a spinner until its metadata loaded', () => {
    render(<ReplayVideo {...props(segmentsOfCodec())}/>);
    expect(video()).toHaveAttribute('src', '/api/data/a%20b.mp4');
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
    fireEvent.loadedMetadata(video());
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
});

it('ReplayVideo with an mp4 recording: jumps to the position when paused and off', () => {
    render(<ReplayVideo {...props(segmentsOfCodec())}/>);
    fireEvent.loadedMetadata(video());
    expect(currentTime).toBe(1);
});

it('ReplayVideo with an mp4 recording: lets a playing video drift a little, but not much', () => {
    const {rerender} = render(<ReplayVideo {...props(segmentsOfCodec(), {playing: true})}/>);
    currentTime = 1.3;
    fireEvent.loadedMetadata(video());
    expect(currentTime).toBe(1.3);
    rerender(<ReplayVideo {...props(segmentsOfCodec(), {playing: true, position: T0 + 3000})}/>);
    expect(currentTime).toBe(3);
});

it('ReplayVideo with an mp4 recording: does not seek before the video has metadata', () => {
    readyState = 0;
    render(<ReplayVideo {...props(segmentsOfCodec())}/>);
    expect(currentTime).toBe(0);
});

it('ReplayVideo with an mp4 recording: plays, pauses and sets the speed with the clock', () => {
    const {rerender} = render(<ReplayVideo {...props(segmentsOfCodec(), {playing: true, rate: 2})}/>);
    expect(play).toHaveBeenCalled();
    expect(video().playbackRate).toBe(2);
    rerender(<ReplayVideo {...props(segmentsOfCodec(), {playing: false, rate: 2})}/>);
    expect(pause).toHaveBeenCalled();
});

it('ReplayVideo with an mp4 recording: survives a refused play', () => {
    play.mockRejectedValue(new Error('blocked'));
    expect(() => render(<ReplayVideo {...props(segmentsOfCodec(), {playing: true})}/>)).not.toThrow();
});

it('ReplayVideo with an mp4 recording: says when the file cannot be played', () => {
    render(<ReplayVideo {...props(segmentsOfCodec())}/>);
    fireEvent.error(video());
    expect(screen.getByText("This recording can't be played.")).toBeInTheDocument();
});

const mjpeg = (over: Parameters<typeof recording>[0] = {}): ReplaySegment[] => segmentsOfCodec({codec: 'mjpeg', indexFile: 'a.idx', ...over});

const respond = (body: string, ok = true): Response =>
    ({ok, status: 206, statusText: 'bad', text: () => Promise.resolve(body), blob: () => Promise.resolve(new Blob(['jpeg']))} as unknown as Response);

const fakeIo = (index: Response): FrameIo => {
    let n = 0;
    return {
        fetch: vi.fn((url: string) => Promise.resolve(url.endsWith('.idx') ? index : respond(''))),
        createObjectUrl: vi.fn(() => `blob:${n++}`),
        revokeObjectUrl: vi.fn(),
    };
};

const INDEX = '0 0 10\n500 10 10\n900 20 10\n';

it('ReplayVideo with an mjpeg recording: shows the frame at the position', async () => {
    const io = fakeIo(respond(INDEX));
    render(<ReplayVideo {...props(mjpeg(), {position: T0 + 600, io})}/>);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
    const img = await screen.findByAltText('Screen');
    expect(img).toHaveAttribute('src', 'blob:0');
    expect(io.fetch).toHaveBeenCalledWith('/api/data/a%20b.mp4', {credentials: 'same-origin', headers: {Range: 'bytes=10-19'}});
});

it('ReplayVideo with an mjpeg recording: fetches the next frames ahead while playing', async () => {
    const io = fakeIo(respond(INDEX));
    render(<ReplayVideo {...props(mjpeg(), {position: T0 + 100, playing: true, io})}/>);
    await screen.findByAltText('Screen');
    await waitFor(() => expect(io.createObjectUrl).toHaveBeenCalledTimes(3));
});

it('ReplayVideo with an mjpeg recording: follows the position to the next frame', async () => {
    const io = fakeIo(respond(INDEX));
    const {rerender} = render(<ReplayVideo {...props(mjpeg(), {position: T0 + 100, io})}/>);
    await screen.findByAltText('Screen');
    rerender(<ReplayVideo {...props(mjpeg(), {position: T0 + 950, io})}/>);
    await waitFor(() => expect(screen.getByAltText('Screen')).toHaveAttribute('src', 'blob:1'));
});

it('ReplayVideo with an mjpeg recording: waits while a frame fails to load', async () => {
    const io = fakeIo(respond(INDEX));
    (io.fetch as ReturnType<typeof vi.fn>).mockImplementation((url: string) =>
        Promise.resolve(url.endsWith('.idx') ? respond(INDEX) : respond('', false)));
    render(<ReplayVideo {...props(mjpeg(), {position: T0 + 100, io})}/>);
    await waitFor(() => expect(io.fetch).toHaveBeenCalledTimes(2));
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
});

it('ReplayVideo with an mjpeg recording: frees the frames on unmount', async () => {
    const io = fakeIo(respond(INDEX));
    const {unmount} = render(<ReplayVideo {...props(mjpeg(), {position: T0 + 100, io})}/>);
    await screen.findByAltText('Screen');
    unmount();
    expect(io.revokeObjectUrl).toHaveBeenCalledWith('blob:0');
});

it.each([
    ['has no index', mjpeg({indexFile: undefined}), respond(INDEX)],
    ['has a failing index', mjpeg(), respond('', false)],
])('says so when the recording %s', async (_, segments, index) => {
    render(<ReplayVideo {...props(segments, {io: fakeIo(index)})}/>);
    expect(await screen.findByText("This recording can't be played.")).toBeInTheDocument();
});

it('ReplayVideo with an mjpeg recording: ignores an index that arrives after unmount', async () => {
    let resolve: (r: Response) => void = () => undefined;
    const io = fakeIo(respond(INDEX));
    (io.fetch as ReturnType<typeof vi.fn>).mockReturnValue(new Promise<Response>(r => {
        resolve = r;
    }));
    const {unmount} = render(<ReplayVideo {...props(mjpeg(), {io})}/>);
    unmount();
    await act(async () => resolve(respond(INDEX)));
    expect(io.createObjectUrl).not.toHaveBeenCalled();
});

it('ReplayVideo with an mjpeg recording: uses the browser fetch by default', async () => {
    const fetchFn = vi.fn().mockResolvedValue(respond('', false));
    vi.stubGlobal('fetch', fetchFn);
    render(<ReplayVideo {...props(mjpeg())}/>);
    expect(await screen.findByText("This recording can't be played.")).toBeInTheDocument();
    expect(fetchFn).toHaveBeenCalledWith('/api/data/a.idx', {credentials: 'same-origin'});
    vi.unstubAllGlobals();
});
