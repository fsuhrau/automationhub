import {afterEach, describe, expect, it, vi} from 'vitest';
import {act, fireEvent, screen, waitFor} from '@testing-library/react';
import {fakeSupport} from './control/fakes';
import {BASE, controlled, loaded, PageWorld, renderPage} from './control/pageKit';

afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
});

const image = (): HTMLImageElement => screen.getByAltText('device screen') as HTMLImageElement;

// the screenshot is 1000x2000 device pixels, shown at 100x200 css pixels
const showScreenshots = (): void => {
    fireEvent.click(screen.getByText('Screenshots'));
    Object.defineProperty(image(), 'naturalWidth', {value: 1000});
    Object.defineProperty(image(), 'naturalHeight', {value: 2000});
    image().getBoundingClientRect = () => ({left: 0, top: 0, width: 100, height: 200}) as DOMRect;
};

const drag = (from: [number, number], to: [number, number]): void => {
    fireEvent.pointerDown(image(), {clientX: from[0], clientY: from[1], pointerId: 1});
    fireEvent.pointerUp(image(), {clientX: to[0], clientY: to[1], pointerId: 1});
};

const sentTo = (page: PageWorld, path: string): unknown[] => page.api.to('POST', `${BASE}/${path}`).map(c => c.data);

describe('the control page: screenshots', () => {
    it('streams screenshots and taps through the REST API with the lease', async () => {
        const page = await controlled();
        showScreenshots();
        expect(image()).toHaveAttribute('src', '/api/p1/device/7/stream?fps=2');
        drag([50, 100], [51, 100]);
        await waitFor(() => expect(sentTo(page, 'tap')).toEqual([{x: 500, y: 1000}]));
        expect(page.api.to('POST', `${BASE}/tap`)[0].config?.headers).toEqual({'X-Device-Lease': 'tok-1'});
    });

    it('swipes when the pointer was dragged', async () => {
        const page = await controlled();
        showScreenshots();
        drag([10, 20], [60, 120]);
        await waitFor(() => expect(sentTo(page, 'swipe')).toHaveLength(1));
        expect(sentTo(page, 'swipe')[0]).toMatchObject({fromX: 100, fromY: 200, toX: 600, toY: 1200});
    });

    it('sends navigation, volume and menu keys through the REST API', async () => {
        const page = await controlled();
        showScreenshots();
        fireEvent.click(screen.getByRole('button', {name: 'Recent apps'}));
        fireEvent.click(screen.getByRole('button', {name: 'Volume up'}));
        fireEvent.click(screen.getByRole('button', {name: 'Volume down'}));
        fireEvent.click(screen.getByRole('button', {name: 'More keys'}));
        fireEvent.click(await screen.findByRole('menuitem', {name: 'Enter'}));
        await waitFor(() => expect(sentTo(page, 'key')).toEqual([
            {key: 'app_switch'}, {key: 'volume_up'}, {key: 'volume_down'}, {key: 'enter'},
        ]));
    });

    it('closes the menu of more keys without sending a key', async () => {
        const page = await controlled();
        fireEvent.click(screen.getByRole('button', {name: 'More keys'}));
        fireEvent.keyDown(await screen.findByRole('menu'), {key: 'Escape'});
        await waitFor(() => expect(screen.queryByRole('menu')).not.toBeInTheDocument());
        expect(sentTo(page, 'key')).toEqual([]);
    });
});

describe('the control page: typing and suggestions', () => {
    it('sends typed text through the REST API', async () => {
        const page = await controlled();
        showScreenshots();
        fireEvent.change(screen.getByLabelText('Text to type on the device'), {target: {value: 'abc'}});
        fireEvent.click(screen.getByRole('button', {name: 'Type on the device'}));
        await waitFor(() => expect(sentTo(page, 'text')).toEqual([{text: 'abc'}]));
        await waitFor(() => expect(screen.getByLabelText('Text to type on the device')).toHaveValue(''));
    });

    it('suggests the faster video mode on Android while screenshots show', async () => {
        await controlled();
        expect(screen.queryByText(/Video mode is faster/)).not.toBeInTheDocument();
        showScreenshots();
        expect(screen.getByText(/Video mode is faster/)).toBeInTheDocument();
    });
});

describe('the control page: paused screen and screenshot', () => {
    it('shows one screenshot when paused and a new one on refresh', async () => {
        renderPage();
        await loaded();
        fireEvent.click(screen.getByText('Paused'));
        const first = image().getAttribute('src');
        expect(first).toMatch(/^\/api\/p1\/device\/7\/screenshot\?t=\d+$/);
        await act(() => new Promise(resolve => setTimeout(resolve, 5)));
        fireEvent.click(screen.getByRole('button', {name: 'Refresh the screen'}));
        expect(image().getAttribute('src')).not.toBe(first);
    });

    it('disables the video mode where the browser cannot decode it', async () => {
        renderPage({support: fakeSupport({decodesVideo: false})});
        await loaded();
        expect(screen.getByRole('button', {name: 'Video'})).toBeDisabled();
        expect(image()).toHaveAttribute('src', '/api/p1/device/7/stream?fps=2');
    });
});

describe('the control page: saving the screenshot', () => {
    it('saves the screenshot of the hub', async () => {
        const createObjectURL = vi.fn(() => 'blob:shot');
        vi.stubGlobal('URL', {createObjectURL, revokeObjectURL: vi.fn()});
        const fetchMock = vi.fn(async () => ({ok: true, blob: async () => new Blob(['png'])}));
        vi.stubGlobal('fetch', fetchMock);
        vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);
        renderPage({support: fakeSupport({decodesVideo: false})});
        await loaded();
        fireEvent.click(screen.getByRole('button', {name: 'Save a screenshot'}));
        await waitFor(() => expect(createObjectURL).toHaveBeenCalled());
        expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/^\/api\/p1\/device\/7\/screenshot\?t=/), {credentials: 'include'});
    });

    it('saves the current frame of a playing video', async () => {
        vi.stubGlobal('URL', {createObjectURL: vi.fn(() => 'blob:frame'), revokeObjectURL: vi.fn()});
        const fetchMock = vi.fn();
        vi.stubGlobal('fetch', fetchMock);
        const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);
        const {support} = renderPage();
        await loaded();
        support.players[0].canvas.toBlob = cb => cb(new Blob(['frame']));
        act(() => support.players[0].callbacks.onStatus({state: 'playing'}));
        fireEvent.click(screen.getByRole('button', {name: 'Save a screenshot'}));
        await waitFor(() => expect(click).toHaveBeenCalled());
        expect(fetchMock).not.toHaveBeenCalled();
    });

    it('shows why saving failed', async () => {
        vi.stubGlobal('fetch', vi.fn(async () => ({ok: false, status: 502})));
        renderPage({support: fakeSupport({decodesVideo: false})});
        await loaded();
        fireEvent.click(screen.getByRole('button', {name: 'Save a screenshot'}));
        expect(await screen.findByRole('alert')).toHaveTextContent('screenshot failed: 502');
    });
});
