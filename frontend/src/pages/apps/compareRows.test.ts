import {expect, it} from 'vitest';
import {flag, list, number, percent, present, sides, signedBytes, signedCount, value} from './compareRows';

it.each([
    [undefined, undefined, 'same'], ['', undefined, 'same'], [undefined, 'x', 'added'], ['x', '', 'removed'], ['x', 'x', 'same'], ['x', 'y', 'changed'],
] as const)('compares %j with %j as %s', (a, b, change) => {
    expect(value('L', a, b)).toMatchObject({kind: 'value', label: 'L', change});
});

it('keeps empty text out of a row', () => {
    expect(value('L', '', 'x')).toMatchObject({a: undefined, b: 'x'});
    expect(present('')).toBe(false);
});

it('compares numbers as bytes or counts with their difference', () => {
    expect(number('S', {a: 1024, b: 3072}, 'bytes')).toMatchObject({a: '1.00KB', b: '3.00KB', delta: 2048, change: 'changed', from: 1024, unit: 'bytes'});
    expect(number('F', {a: 1000, b: 1000}, 'count')).toMatchObject({a: '1,000', change: 'same', delta: 0});
    expect(number('F', {b: 5}, 'count')).toMatchObject({a: undefined, change: 'added', delta: undefined});
    expect(number('F', {a: 5}, 'count')).toMatchObject({b: undefined, change: 'removed'});
    expect(number('F', {}, 'count')).toMatchObject({change: 'same'});
});

it('compares flags as yes and no', () => {
    expect(flag('D', true, false)).toMatchObject({a: 'yes', b: 'no', change: 'changed'});
    expect(flag('D', undefined, true)).toMatchObject({a: undefined, change: 'added'});
});

it('compares lists as sets, sorted, with what both have', () => {
    expect(list('P', ['b', 'a', 'c'], ['c', 'd'])).toEqual({kind: 'list', label: 'P', added: ['d'], removed: ['a', 'b'], same: 1, change: 'changed'});
    expect(list('P', ['a'], ['a'])).toMatchObject({change: 'same', same: 1});
    expect(list('P', undefined, undefined)).toMatchObject({added: [], removed: [], same: 0, change: 'same'});
});

it('shows changes with their sign', () => {
    expect(signedBytes(0)).toBe('±0');
    expect(signedBytes(2048)).toBe('+2.00KB');
    expect(signedBytes(-1024)).toBe('−1.00KB');
    expect(signedCount(0)).toBe('±0');
    expect(signedCount(1500)).toBe('+1,500');
    expect(signedCount(-3)).toBe('−3');
});

it('says the change relative to the old value when it is worth saying', () => {
    expect(percent(200, 50)).toBe('+25.0 %');
    expect(percent(200, -50)).toBe('−25.0 %');
    expect(percent(0, 5)).toBeUndefined();
    expect(percent(undefined, 5)).toBeUndefined();
    expect(percent(1000000, 1)).toBeUndefined();
});

it('builds the rows of two things of one kind, a missing side being empty', () => {
    const s = sides<{ n?: string, ok?: boolean, size?: number, tags?: string[] }>({n: 'a', ok: true, size: 1, tags: ['x']}, undefined);
    expect(s.text('N', i => i.n)).toMatchObject({a: 'a', b: undefined, change: 'removed'});
    expect(s.flag('Ok', i => i.ok)).toMatchObject({a: 'yes', change: 'removed'});
    expect(s.number('Size', i => i.size, 'count')).toMatchObject({a: '1', change: 'removed'});
    expect(s.list('T', i => i.tags)).toMatchObject({removed: ['x']});
});
