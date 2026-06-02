import type { TerminalCore } from "@wterm/core";
export declare class Renderer {
    private container;
    private rows;
    private cols;
    private rowEls;
    private prevCursorRow;
    private prevCursorCol;
    private prevContainerBg;
    private prevRowBg;
    private _scrollbackRowEls;
    private _renderedScrollbackCount;
    constructor(container: HTMLElement);
    setup(cols: number, rows: number): void;
    private _buildRowContent;
    private _buildScrollbackRowEl;
    private syncScrollback;
    render(core: TerminalCore): void;
}
//# sourceMappingURL=renderer.d.ts.map