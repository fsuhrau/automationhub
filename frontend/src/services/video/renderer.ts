// Draws decoded frames into the canvas and maps client positions to pixels.

export interface Renderer {
    draw: (source: CanvasImageSource, width: number, height: number) => void,
}

export interface VideoPoint {
    x: number,
    y: number,
    width: number,
    height: number,
}

export const createRenderer = (canvas: HTMLCanvasElement): Renderer => {
    const ctx = canvas.getContext('2d');
    return {
        draw: (source, width, height) => {
            if (canvas.width !== width || canvas.height !== height) {
                canvas.width = width;
                canvas.height = height;
            }
            ctx?.drawImage(source, 0, 0);
        },
    };
};

const clamp = (v: number, max: number): number => Math.min(Math.max(Math.round(v), 0), max - 1);

// maps a client position to pixels of the current video frame
export const toVideoPoint = (canvas: HTMLCanvasElement, clientX: number, clientY: number): VideoPoint => {
    const rect = canvas.getBoundingClientRect();
    const {width, height} = canvas;
    return {
        x: clamp((clientX - rect.left) * width / rect.width, width),
        y: clamp((clientY - rect.top) * height / rect.height, height),
        width,
        height,
    };
};
