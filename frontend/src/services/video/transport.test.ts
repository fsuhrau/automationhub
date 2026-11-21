import {afterEach, beforeEach, expect, it, Mock, vi} from 'vitest';
import {createFakeBrowser, FakeBrowser, FakeSocket} from './fakes';
import {createTransport, RECONNECT_MAX_MS, Transport, TransportHandlers} from './transport';

let browser: FakeBrowser;
let handlers: { [K in keyof TransportHandlers]: Mock<TransportHandlers[K]> };
let transport: Transport;

beforeEach(() => {
    vi.useFakeTimers();
    browser = createFakeBrowser();
    handlers = {
        onConnecting: vi.fn<() => void>(),
        onText: vi.fn<(t: string) => void>(),
        onBinary: vi.fn<(d: Uint8Array) => void>(),
        onDisconnect: vi.fn<() => void>(),
    };
    transport = createTransport({url: 'ws://x', createSocket: browser.deps.createSocket, handlers});
});
afterEach(() => vi.useRealTimers());

const socket = (i = 0): FakeSocket => browser.sockets[i];

it('connects on creation with an arraybuffer socket', () => {
    expect(handlers.onConnecting).toHaveBeenCalledTimes(1);
    expect(socket().url).toBe('ws://x');
    expect(socket().binaryType).toBe('arraybuffer');
});

it('passes text and binary messages on', () => {
    socket().emit('hello');
    socket().emit(new Uint8Array([1, 2]).buffer);
    expect(handlers.onText).toHaveBeenCalledWith('hello');
    expect(handlers.onBinary).toHaveBeenCalledWith(new Uint8Array([1, 2]));
});

it('sends only while the socket is open', () => {
    expect(transport.isOpen).toBe(false);
    expect(transport.send('a')).toBe(false);
    socket().open();
    expect(transport.isOpen).toBe(true);
    expect(transport.send('b')).toBe(true);
    expect(socket().sent).toEqual(['b']);
});

it('reconnects after the socket closes, doubling the delay up to the maximum', () => {
    const delays: number[] = [];
    for (let i = 0; i < 6; i++) {
        socket(i).serverClose();
        const before = browser.sockets.length;
        let waited = 0;
        while (browser.sockets.length === before) {
            vi.advanceTimersByTime(500);
            waited += 500;
        }
        delays.push(waited);
    }
    expect(delays).toEqual([1000, 2000, 4000, 8000, RECONNECT_MAX_MS, RECONNECT_MAX_MS]);
    expect(handlers.onDisconnect).toHaveBeenCalledTimes(6);
    expect(handlers.onConnecting).toHaveBeenCalledTimes(7);
});

it('starts the backoff over after resetBackoff', () => {
    socket(0).serverClose();
    vi.advanceTimersByTime(1000);
    socket(1).serverClose();
    transport.resetBackoff();
    vi.advanceTimersByTime(2000);
    socket(2).serverClose();
    vi.advanceTimersByTime(999);
    expect(browser.sockets).toHaveLength(3);
    vi.advanceTimersByTime(1);
    expect(browser.sockets).toHaveLength(4);
});

it('does not reconnect after giveUp', () => {
    transport.giveUp();
    socket().serverClose();
    vi.advanceTimersByTime(60000);
    expect(browser.sockets).toHaveLength(1);
    expect(handlers.onDisconnect).toHaveBeenCalled();
});

it('drop closes the socket and then reconnects', () => {
    transport.drop();
    vi.advanceTimersByTime(1000);
    expect(browser.sockets).toHaveLength(2);
});

it('drop without a socket does nothing', () => {
    socket().serverClose();
    expect(() => transport.drop()).not.toThrow();
});

it('stop closes the socket without telling the handlers and cancels a pending reconnect', () => {
    socket().serverClose();
    transport.stop();
    vi.advanceTimersByTime(60000);
    expect(browser.sockets).toHaveLength(1);
});

it('stop closes the open socket and ignores its close event', () => {
    socket().open();
    transport.stop();
    expect(socket().readyState).toBe(3);
    expect(handlers.onDisconnect).not.toHaveBeenCalled();
    vi.advanceTimersByTime(60000);
    expect(browser.sockets).toHaveLength(1);
});

it('ignores the close event of an old socket', () => {
    const old = socket();
    old.serverClose();
    vi.advanceTimersByTime(1000);
    old.onclose?.();
    expect(handlers.onDisconnect).toHaveBeenCalledTimes(1);
});
