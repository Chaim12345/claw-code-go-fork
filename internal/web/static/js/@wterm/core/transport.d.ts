export interface WebSocketTransportOptions {
    url?: string;
    reconnect?: boolean;
    maxReconnectDelay?: number;
    onData?: (data: Uint8Array | string) => void;
    onOpen?: () => void;
    onClose?: () => void;
    onError?: (event: Event) => void;
}
export declare class WebSocketTransport {
    url: string | null;
    reconnect: boolean;
    maxReconnectDelay: number;
    onData: ((data: Uint8Array | string) => void) | null;
    onOpen: (() => void) | null;
    onClose: (() => void) | null;
    onError: ((event: Event) => void) | null;
    private _ws;
    private _reconnectTimer;
    private _reconnectDelay;
    private _closed;
    private _buffer;
    constructor(options?: WebSocketTransportOptions);
    connect(url?: string): void;
    send(data: string | Uint8Array): void;
    close(): void;
    get connected(): boolean;
    private _flushBuffer;
    private _scheduleReconnect;
}
//# sourceMappingURL=transport.d.ts.map