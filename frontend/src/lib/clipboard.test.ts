import {describe, expect, it, vi} from 'vitest';
import {copyText} from './clipboard';

describe('copyText', () => {
    it('writes the text and reports success', async () => {
        const writeText = vi.fn().mockResolvedValue(undefined);
        expect(await copyText('abc', {writeText})).toBe(true);
        expect(writeText).toHaveBeenCalledWith('abc');
    });

    it('reports failure when the write is denied', async () => {
        const writeText = vi.fn().mockRejectedValue(new Error('denied'));
        expect(await copyText('abc', {writeText})).toBe(false);
    });

    it('reports failure without a clipboard', async () => {
        expect(await copyText('abc', undefined)).toBe(false);
    });

    it('uses the browser clipboard by default', async () => {
        const writeText = vi.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, 'clipboard', {value: {writeText}, configurable: true});
        expect(await copyText('x')).toBe(true);
        expect(writeText).toHaveBeenCalledWith('x');
    });
});
