import {describe, expect, it, vi} from 'vitest';
import {act, fireEvent, screen} from '@testing-library/react';
import {controlled, loaded, renderPage} from './control/controlHarness';

const stage = (): HTMLElement => screen.getByLabelText('Device screen');
const canvas = (): HTMLCanvasElement => stage().querySelector('canvas') as HTMLCanvasElement;
const image = (): HTMLImageElement => screen.getByAltText('device screen') as HTMLImageElement;

describe('the control page: live video', () => {
    it('plays the video of the device and shows how it connects', async () => {
        const {support} = renderPage();
        await loaded();
        expect(support.players[0].url).toContain('/api/p1/device/7/video');
        act(() => support.players[0].callbacks.onStatus({state: 'connecting'}));
        expect(screen.getByText('Connecting…')).toBeInTheDocument();
        act(() => support.players[0].callbacks.onStatus({state: 'error', message: 'socket closed'}));
        expect(screen.getByText('socket closed')).toBeInTheDocument();
    });

    it('says when touch, keys and scrolling go straight through the video', async () => {
        const {support} = await controlled();
        expect(screen.queryByText('Live')).not.toBeInTheDocument();
        act(() => support.players[0].callbacks.onStatus({state: 'playing'}));
        expect(screen.getByText('Live')).toBeInTheDocument();
    });

    it('sends a touch over the video socket', async () => {
        const {support} = await controlled();
        canvas().setPointerCapture = vi.fn();
        fireEvent.pointerDown(canvas(), {clientX: 10, clientY: 20, pointerId: 1});
        fireEvent.pointerUp(canvas(), {clientX: 10, clientY: 20, pointerId: 1});
        expect(support.players[0].sent.map(s => [s.lease, (s.input as { action: string }).action])).toEqual([['tok-1', 'down'], ['tok-1', 'up']]);
    });
});

describe('the control page: live input', () => {
    it('sends keys and typed text over the video socket', async () => {
        const {support, api} = await controlled();
        fireEvent.click(screen.getByRole('button', {name: 'Home'}));
        fireEvent.change(screen.getByLabelText('Text to type on the device'), {target: {value: 'hello'}});
        fireEvent.keyDown(screen.getByLabelText('Text to type on the device'), {key: 'Enter'});
        expect(support.players[0].sent.map(s => s.input)).toEqual([{kind: 'key', action: 'press', key: 'home'}, {kind: 'text', text: 'hello'}]);
        expect(screen.getByLabelText('Text to type on the device')).toHaveValue('');
        expect(api.calls.filter(c => c.url.endsWith('/key') || c.url.endsWith('/text'))).toHaveLength(0);
    });

    it('types what the keyboard sends while the screen has focus', async () => {
        const {support} = await controlled();
        expect(screen.getByText('Click the screen to type on it with your keyboard.')).toBeInTheDocument();
        fireEvent.focus(stage());
        expect(screen.getByText(/Your keyboard types on the device/)).toBeInTheDocument();
        fireEvent.keyDown(stage(), {key: 'a'});
        fireEvent.blur(stage());
        expect(support.players[0].sent.map(s => s.input)).toEqual([{kind: 'text', text: 'a'}]);
    });

    it('falls back to screenshots and says why when the device has no video', async () => {
        const {support} = renderPage();
        await loaded();
        act(() => support.players[0].callbacks.onStatus({state: 'unsupported', message: 'WebDriverAgent did not start'}));
        expect(await screen.findByText('Live video unavailable, showing screenshots: WebDriverAgent did not start')).toBeInTheDocument();
        expect(image()).toHaveAttribute('src', '/api/p1/device/7/stream?fps=2');
        expect(support.players[0].closed).toBe(true);
    });
});
