import {expect, it} from 'vitest';
import {bundle} from './appsFixtures';
import {pendingPair, selectionOf, swapped, withPick} from './compareSelectionModel';

const list = [bundle(5), bundle(3), bundle(1)];

it('compares the newest bundle with the one before it by default', () => {
    expect(selectionOf(list, null, null)).toEqual({a: 3, b: 5});
});

it('compares the picked bundle with the one before it', () => {
    expect(selectionOf(list, null, '3')).toEqual({a: 1, b: 3});
});

it('compares the oldest bundle with the first other one in the list', () => {
    expect(selectionOf(list, null, '1')).toEqual({a: 5, b: 1});
});

it('takes both bundles from the url', () => {
    expect(selectionOf(list, '1', '5')).toEqual({a: 1, b: 5});
    expect(selectionOf(null, '1', '5')).toEqual({a: 1, b: 5});
    expect(selectionOf(null, '2', null)).toEqual({a: 2, b: undefined});
});

it('has nothing to pick before the bundles are there', () => {
    expect(selectionOf(null, null, null)).toEqual({b: undefined});
    expect(selectionOf([], null, null)).toEqual({b: undefined});
    expect(selectionOf(null, null, '4')).toEqual({b: 4});
});

it('builds the url selection of a pick and of a swap', () => {
    expect(withPick({a: 1, b: 2}, 'a', 9)).toEqual({a: '9', b: '2'});
    expect(withPick({}, 'b', 9)).toEqual({a: '', b: '9'});
    expect(swapped({a: 1, b: 2})).toEqual({a: '2', b: '1'});
    expect(swapped({})).toEqual({a: '', b: ''});
});

it('waits for the details only when there are two bundles to compare', () => {
    expect(pendingPair(list, false, false)).toBe(true);
    expect(pendingPair(list, true, false)).toBe(false);
    expect(pendingPair(list, false, true)).toBe(false);
    expect(pendingPair([bundle(1)], false, false)).toBe(false);
    expect(pendingPair(null, false, false)).toBe(false);
});
