// Pointer events for hook tests: jsdom has no layout or pointer capture.
import {vi} from 'vitest';
import type {PointerEvent} from 'react';

export interface ScreenElement {
    el: HTMLCanvasElement,
    parent: HTMLElement,
}

// a canvas of 100x200 css px at (10, 10) in a focusable parent
export const screenElement = (): ScreenElement => {
    const parent = document.createElement('div');
    parent.tabIndex = 0;
    const el = document.createElement('canvas');
    parent.appendChild(el);
    document.body.appendChild(parent);
    el.width = 360;
    el.height = 640;
    el.getBoundingClientRect = () => ({left: 10, top: 10, width: 100, height: 200}) as DOMRect;
    el.setPointerCapture = vi.fn();
    return {el, parent};
};

export interface PointerInit {
    id?: number,
}

// a pointer event at (x, y); id tells pointers apart
export const pointer = (el: HTMLElement, {x, y}: { x: number, y: number }, {id = 1}: PointerInit = {}): PointerEvent<HTMLElement> =>
    ({currentTarget: el, clientX: x, clientY: y, pointerId: id, preventDefault: vi.fn()}) as unknown as PointerEvent<HTMLElement>;
