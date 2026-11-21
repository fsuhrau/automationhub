// The websocket of the video stream with reconnect and exponential backoff.

export const RECONNECT_MIN_MS = 1000;
export const RECONNECT_MAX_MS = 10000;
const WS_OPEN = 1;

export interface TransportHandlers {
    onConnecting: () => void,
    onText: (text: string) => void,
    onBinary: (data: Uint8Array) => void,
    // the socket closed on its own (not through stop)
    onDisconnect: () => void,
}

export interface TransportOptions {
    url: string,
    createSocket: (url: string) => WebSocket,
    handlers: TransportHandlers,
}

export interface Transport {
    readonly isOpen: boolean,
    send: (text: string) => boolean,
    resetBackoff: () => void,
    // no reconnect after the socket closes
    giveUp: () => void,
    // closes the socket; it reconnects through the usual backoff
    drop: () => void,
    stop: () => void,
}

class SocketTransport implements Transport {
    private ws: WebSocket | null = null;
    private timer: ReturnType<typeof setTimeout> | null = null;
    private retryMs = RECONNECT_MIN_MS;
    private stopped = false;
    private gaveUp = false;

    constructor(private readonly options: TransportOptions) {
        this.connect();
    }

    get isOpen(): boolean {
        return this.ws?.readyState === WS_OPEN;
    }

    send(text: string): boolean {
        if (!this.ws || this.ws.readyState !== WS_OPEN) return false;
        this.ws.send(text);
        return true;
    }

    resetBackoff(): void {
        this.retryMs = RECONNECT_MIN_MS;
    }

    giveUp(): void {
        this.gaveUp = true;
    }

    drop(): void {
        this.ws?.close();
    }

    stop(): void {
        this.stopped = true;
        if (this.timer) clearTimeout(this.timer);
        const socket = this.ws;
        this.ws = null;
        socket?.close();
    }

    private connect(): void {
        const {url, createSocket, handlers} = this.options;
        handlers.onConnecting();
        const socket = createSocket(url);
        socket.binaryType = 'arraybuffer';
        socket.onmessage = event => this.onMessage(event.data);
        socket.onclose = () => this.onClose(socket);
        this.ws = socket;
    }

    private onMessage(data: unknown): void {
        const {onText, onBinary} = this.options.handlers;
        if (typeof data === 'string') onText(data);
        else onBinary(new Uint8Array(data as ArrayBuffer));
    }

    private onClose(socket: WebSocket): void {
        if (this.ws !== socket) return;
        this.ws = null;
        this.options.handlers.onDisconnect();
        if (this.stopped || this.gaveUp) return;
        this.timer = setTimeout(() => this.connect(), this.retryMs);
        this.retryMs = Math.min(this.retryMs * 2, RECONNECT_MAX_MS);
    }
}

export const createTransport = (options: TransportOptions): Transport => new SocketTransport(options);
