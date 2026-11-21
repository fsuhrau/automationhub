import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {downloadText} from './downloadText';

describe('downloadText', () => {
    beforeEach(() => {
        vi.useFakeTimers();
        URL.createObjectURL = vi.fn(() => 'blob:text');
        URL.revokeObjectURL = vi.fn();
    });

    afterEach(() => {
        vi.useRealTimers();
        vi.restoreAllMocks();
    });

    it('clicks a link to the text with the file name and frees the url later', async () => {
        let download = '';
        let href = '';
        vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) {
            download = this.download;
            href = this.href;
        });
        downloadText('log.txt', 'hello');
        expect(await (URL.createObjectURL as ReturnType<typeof vi.fn>).mock.calls[0][0].text()).toBe('hello');
        expect([download, href]).toEqual(['log.txt', 'blob:text']);
        expect(URL.revokeObjectURL).not.toHaveBeenCalled();
        vi.advanceTimersByTime(1000);
        expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:text');
    });
});
