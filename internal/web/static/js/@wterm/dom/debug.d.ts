import type { TerminalCore, CellData } from "@wterm/core";
export interface TraceEntry {
    ts: number;
    type: "csi" | "sgr" | "osc" | "esc" | "text";
    raw: string;
    params?: number[];
    private?: string;
    final?: string;
}
export interface CellInfo extends CellData {
    charStr: string;
    flagNames: string[];
}
export interface GridSummary {
    rows: number;
    cols: number;
    cursor: {
        row: number;
        col: number;
        visible: boolean;
    };
    altScreen: boolean;
    scrollbackCount: number;
}
export interface PerfStats {
    frameCount: number;
    totalRenderMs: number;
    avgRenderMs: number;
    maxRenderMs: number;
    lastDirtyRows: number;
}
export interface UnhandledEntry {
    final: string;
    private: string;
    paramCount: number;
    params: number[];
}
export declare class DebugAdapter {
    private _traces;
    private _bridge;
    private _perf;
    get traces(): readonly TraceEntry[];
    get perf(): Readonly<PerfStats>;
    setBridge(bridge: TerminalCore): void;
    traceWrite(data: string | Uint8Array): void;
    recordRender(renderMs: number, dirtyRows: number): void;
    resetPerf(): void;
    cell(row: number, col: number): CellInfo | null;
    row(row: number): CellInfo[] | null;
    grid(): GridSummary | null;
    unhandled(): UnhandledEntry[];
    dump(count?: number): void;
    dumpUnhandled(): void;
}
//# sourceMappingURL=debug.d.ts.map