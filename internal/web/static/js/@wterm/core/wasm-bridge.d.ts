import type { CellData, CursorState, UnhandledSequence, TerminalCore } from "./terminal-core.js";
export declare class WasmBridge implements TerminalCore {
    private exports;
    private memory;
    private gridPtr;
    private dirtyPtr;
    private writeBufferPtr;
    private cellSize;
    private maxCols;
    private encoder;
    private decoder;
    private _dv;
    constructor(instance: WebAssembly.Instance);
    static load(url?: string): Promise<WasmBridge>;
    init(cols: number, rows: number): void;
    private _updatePointers;
    writeString(str: string): void;
    writeRaw(data: Uint8Array): void;
    getCell(row: number, col: number): CellData;
    isDirtyRow(row: number): boolean;
    clearDirty(): void;
    getCursor(): CursorState;
    getCols(): number;
    getRows(): number;
    cursorKeysApp(): boolean;
    bracketedPaste(): boolean;
    usingAltScreen(): boolean;
    getTitle(): string | null;
    getResponse(): string | null;
    getScrollbackCount(): number;
    getScrollbackCell(offset: number, col: number): CellData;
    getScrollbackLineLen(offset: number): number;
    getUnhandledSequences(): UnhandledSequence[];
    resize(cols: number, rows: number): void;
}
//# sourceMappingURL=wasm-bridge.d.ts.map