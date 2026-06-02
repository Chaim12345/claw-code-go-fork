import { type TerminalCore } from "@wterm/core";
import { DebugAdapter } from "./debug.js";
export interface WTermOptions {
    cols?: number;
    rows?: number;
    /**
     * A pre-constructed terminal core. When provided, `wasmUrl` is ignored and
     * this core is used directly instead of loading the built-in Zig WASM binary.
     */
    core?: TerminalCore;
    wasmUrl?: string;
    autoResize?: boolean;
    cursorBlink?: boolean;
    debug?: boolean;
    onData?: (data: string) => void;
    onTitle?: (title: string) => void;
    onResize?: (cols: number, rows: number) => void;
}
export declare class WTerm {
    element: HTMLElement;
    cols: number;
    rows: number;
    bridge: TerminalCore | null;
    autoResize: boolean;
    debug: DebugAdapter | null;
    private _coreOption;
    private wasmUrl;
    private _debugEnabled;
    private renderer;
    private input;
    private rafId;
    private _renderTimer;
    private resizeObserver;
    private _destroyed;
    private _shouldScrollToBottom;
    private _rowHeight;
    private _onClickFocus;
    onData: ((data: string) => void) | null;
    onTitle: ((title: string) => void) | null;
    onResize: ((cols: number, rows: number) => void) | null;
    private _container;
    constructor(element: HTMLElement, options?: WTermOptions);
    init(): Promise<this>;
    private _isScrolledToBottom;
    private _scrollToBottom;
    write(data: string | Uint8Array): void;
    resize(cols: number, rows: number): void;
    focus(): void;
    private _scheduleRender;
    private _initialRender;
    private _doRender;
    private _lockHeight;
    private _setRowHeight;
    private _measureCharSize;
    private _setupResizeObserver;
    destroy(): void;
}
//# sourceMappingURL=wterm.d.ts.map