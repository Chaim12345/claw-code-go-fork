import type { TerminalCore } from "@wterm/core";
export declare class InputHandler {
    private element;
    private textarea;
    private onData;
    private getBridge;
    private composing;
    private _onKeyDown;
    private _onPaste;
    private _onCompositionStart;
    private _onCompositionEnd;
    private _onInput;
    private _onFocus;
    private _onBlur;
    constructor(element: HTMLElement, onData: (data: string) => void, getBridge: () => TerminalCore | null);
    focus(): void;
    destroy(): void;
    private handleKeyDown;
    private handlePaste;
    private handleCompositionStart;
    private handleCompositionEnd;
    private handleInput;
    private keyToSequence;
}
//# sourceMappingURL=input.d.ts.map