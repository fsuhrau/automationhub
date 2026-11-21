import {formatDateTime} from '../../../components/DateTime';

// a file name for a screenshot of a device
export const screenshotName = (deviceName: string, now: number): string =>
    `${deviceName}-${formatDateTime(now, 'YYYYMMDD-HHmmss')}.png`.replace(/[^\w.-]+/g, '_');

export const saveBlob = (blob: Blob, name: string): void => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = name;
    a.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
};

export interface ScreenshotSource {
    // the canvas of a playing video, whose current frame is the screenshot
    canvas: HTMLCanvasElement | null,
    // otherwise the hub's screenshot
    url: string,
}

const canvasBlob = (canvas: HTMLCanvasElement | null): Promise<Blob | null> =>
    canvas ? new Promise(resolve => canvas.toBlob(resolve, 'image/png')) : Promise.resolve(null);

export const fetchScreenshot = async ({canvas, url}: ScreenshotSource): Promise<Blob> => {
    const frame = await canvasBlob(canvas);
    if (frame) return frame;
    const response = await fetch(url, {credentials: 'include'});
    if (!response.ok) throw new Error(`screenshot failed: ${response.status}`);
    return response.blob();
};
