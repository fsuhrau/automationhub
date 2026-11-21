import {describe, expect, it} from 'vitest';
import {ToArray} from './enum_to_array';

enum Color {
    Red,
    Green = 5,
}

describe('ToArray', () => {
    it('lists the members of an enum as id and name, in id order', () => {
        expect(ToArray(Color)).toEqual([{id: '0', name: 'Red'}, {id: '5', name: 'Green'}]);
    });

    it('lists nothing for an enum without numbers', () => {
        expect(ToArray({A: 'a'})).toEqual([]);
    });
});
