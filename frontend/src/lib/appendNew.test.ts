import {describe, expect, it} from 'vitest';
import {appendNew} from './appendNew';

const e = (id: number | string): { id: number | string } => ({id});

describe('appendNew', () => {
    it('appends items with new ids', () => {
        expect(appendNew([e(1)], [e(2), e(3)])).toEqual([e(1), e(2), e(3)]);
    });

    it('ignores known ids and repeats inside the batch', () => {
        expect(appendNew([e(1)], [e(1), e(2), e(2)])).toEqual([e(1), e(2)]);
    });

    it('returns the same list when nothing is new', () => {
        const list = [e('a')];
        expect(appendNew(list, [e('a')])).toBe(list);
    });

    it('keeps the last entries within the limit', () => {
        expect(appendNew([e(1), e(2)], [e(3)], 2)).toEqual([e(2), e(3)]);
    });
});
