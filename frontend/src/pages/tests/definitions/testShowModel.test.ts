import {describe, expect, it} from 'vitest';
import {aFunction} from './testing/fixtures';
import {methodsByClass, runResults} from './testShowModel';

describe('runResults', () => {
    it('lists the counts that are not zero', () => {
        expect(runResults({passed: 3, failed: 1, unstable: 2, open: 4})).toBe('3 passed · 1 failed · 2 unstable · 4 open');
        expect(runResults({passed: 3, failed: 0, unstable: 0, open: 0})).toBe('3 passed');
    });

    it('is empty without counts', () => {
        expect(runResults({passed: 0, failed: 0, unstable: 0, open: 0})).toBe('');
    });
});

describe('methodsByClass', () => {
    it('groups by class, falling back to the assembly and to nothing', () => {
        const groups = methodsByClass([
            aFunction('A', 'X'), aFunction('B', ''), {assembly: '', class: '', method: 'C'}, aFunction('D', 'X'),
        ]);
        expect(groups).toEqual([['X', ['A', 'D']], ['Game', ['B']], ['', ['C']]]);
    });
});
