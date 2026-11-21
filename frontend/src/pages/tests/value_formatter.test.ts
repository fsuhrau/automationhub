import {describe, expect, it} from 'vitest';
import {byteFormat, fixedTwoFormat, kFormat} from './value_formatter';

describe('kFormat', () => {
    it.each([[999, '999.00'], [1000, '1.00k'], [12345, '12.35k']])('formats %d as %s', (value, text) => {
        expect(kFormat(value)).toBe(text);
    });
});

describe('fixedTwoFormat', () => {
    it('keeps two decimals', () => {
        expect(fixedTwoFormat(1.5)).toBe('1.50');
    });
});

describe('byteFormat', () => {
    it.each([
        [512, '512B'],
        [2048, '2.00KB'],
        [3 * 1024 * 1024, '3.00MB'],
        [5 * 1024 * 1024 * 1024, '5.00GB'],
    ])('formats %d bytes as %s', (bytes, text) => {
        expect(byteFormat(bytes)).toBe(text);
    });
});
