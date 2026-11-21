import {expect, it} from 'vitest';
import {SizeEntry} from '../../types/bundle.info';
import {sizes} from './compareSizes';

const entry = (path: string, size: number): SizeEntry => ({path, size, compressed: size});

it('compares complete lists by path, biggest change first', () => {
    const rows = sizes([entry('a', 10), entry('b', 100), entry('c', 5)], [entry('a', 40), entry('b', 100), entry('d', 7)]);
    expect(rows.map(r => [r.path, r.delta, r.change])).toEqual([
        ['a', 30, 'changed'], ['d', 7, 'added'], ['c', -5, 'removed'], ['b', 0, 'same'],
    ]);
});

it('does not know the change of an entry that one partial list lacks', () => {
    const rows = sizes([entry('a', 10), entry('gone', 9)], [entry('a', 20), entry('new', 8)], true);
    expect(rows.map(r => [r.path, r.delta])).toEqual([['a', 10], ['gone', undefined], ['new', undefined]]);
});

it('sorts by path when the changes are the same, and takes missing lists as empty', () => {
    expect(sizes([entry('b', 1), entry('a', 1)], [entry('b', 1), entry('a', 1)]).map(r => r.path)).toEqual(['a', 'b']);
    expect(sizes(undefined, undefined)).toEqual([]);
});
