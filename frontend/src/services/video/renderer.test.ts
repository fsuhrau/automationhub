import {describe, expect, it} from 'vitest';
import {fakeCanvas} from './fakes';
import {createRenderer, toVideoPoint} from './renderer';

describe('createRenderer', () => {
    it('resizes the canvas to the frame and draws it', () => {
        const {canvas, drawImage} = fakeCanvas();
        const source = {} as CanvasImageSource;
        createRenderer(canvas).draw(source, 40, 30);
        expect([canvas.width, canvas.height]).toEqual([40, 30]);
        expect(drawImage).toHaveBeenCalledWith(source, 0, 0);
    });

    it.each([[40, 31], [41, 30]])('resizes when only the size %ix%i differs', (w, h) => {
        const {canvas} = fakeCanvas(40, 30);
        createRenderer(canvas).draw({} as CanvasImageSource, w, h);
        expect([canvas.width, canvas.height]).toEqual([w, h]);
    });

    it('keeps the canvas size when the frame has the same size', () => {
        const {canvas, drawImage} = fakeCanvas(40, 30);
        createRenderer(canvas).draw({} as CanvasImageSource, 40, 30);
        expect([canvas.width, canvas.height]).toEqual([40, 30]);
        expect(drawImage).toHaveBeenCalled();
    });

    it('draws nothing without a 2d context', () => {
        const canvas = {width: 1, height: 1, getContext: () => null} as unknown as HTMLCanvasElement;
        expect(() => createRenderer(canvas).draw({} as CanvasImageSource, 2, 2)).not.toThrow();
    });
});

describe('toVideoPoint', () => {
    // the rect is 200x100 at (10, 20)
    it.each([
        {name: 'the middle', at: [110, 70], expected: {x: 50, y: 25}},
        {name: 'the top left corner', at: [10, 20], expected: {x: 0, y: 0}},
        {name: 'left of and above the canvas', at: [-50, -50], expected: {x: 0, y: 0}},
        {name: 'right of and below the canvas', at: [500, 500], expected: {x: 99, y: 49}},
    ])('maps $name', ({at, expected}) => {
        const {canvas} = fakeCanvas(100, 50);
        expect(toVideoPoint(canvas, at[0], at[1])).toEqual({...expected, width: 100, height: 50});
    });
});
