// The rows of a comparison: one attribute, one list or one number of two
// bundles, and how to show a change.
import {Change, ListRow, ValueRow} from '../../types/bundle.compare';
import {byteFormat} from '../tests/value_formatter';

type Unit = 'bytes' | 'count';

export const present = (v: string | undefined): v is string => v !== undefined && v !== '';

const changeOf = (a: string | undefined, b: string | undefined): Change => {
    if (!present(a) && !present(b)) return 'same';
    if (!present(a)) return 'added';
    if (!present(b)) return 'removed';
    return a === b ? 'same' : 'changed';
};

export const value = (label: string, a: string | undefined, b: string | undefined): ValueRow =>
    ({kind: 'value', label, a: present(a) ? a : undefined, b: present(b) ? b : undefined, change: changeOf(a, b)});

const showNumber = (n: number | undefined, unit: Unit): string | undefined => {
    if (n === undefined) return undefined;
    return unit === 'bytes' ? byteFormat(n) : n.toLocaleString();
};

const numberChange = (a: number | undefined, b: number | undefined): Change => {
    if (a === undefined && b === undefined) return 'same';
    if (a === undefined) return 'added';
    if (b === undefined) return 'removed';
    return a === b ? 'same' : 'changed';
};

export interface NumberPair {
    a?: number,
    b?: number,
}

export const number = (label: string, {a, b}: NumberPair, unit: Unit): ValueRow => ({
    kind: 'value', label, a: showNumber(a, unit), b: showNumber(b, unit), unit, from: a,
    change: numberChange(a, b),
    delta: a !== undefined && b !== undefined ? b - a : undefined,
});

const yesNo = (v: boolean | undefined): string | undefined => v === undefined ? undefined : v ? 'yes' : 'no';

export const flag = (label: string, a: boolean | undefined, b: boolean | undefined): ValueRow =>
    value(label, yesNo(a), yesNo(b));

// list compares two lists as sets; a missing list is an empty one
export const list = (label: string, a: string[] | undefined, b: string[] | undefined): ListRow => {
    const inA = new Set(a ?? []);
    const inB = new Set(b ?? []);
    const added = [...inB].filter(x => !inA.has(x)).sort();
    const removed = [...inA].filter(x => !inB.has(x)).sort();
    return {
        kind: 'list', label, added, removed, same: [...inA].filter(x => inB.has(x)).length,
        change: added.length === 0 && removed.length === 0 ? 'same' : 'changed',
    };
};

const sign = (delta: number): string => delta > 0 ? '+' : '−';

// signedBytes is a size change with its sign, like "+1.20MB"
export const signedBytes = (delta: number): string =>
    delta === 0 ? '±0' : `${sign(delta)}${byteFormat(Math.abs(delta))}`;

// signedCount is a count change with its sign, like "+3"
export const signedCount = (delta: number): string =>
    delta === 0 ? '±0' : `${sign(delta)}${Math.abs(delta).toLocaleString()}`;

// percent is the change relative to a, or undefined when a is 0
export const percent = (a: number | undefined, delta: number): string | undefined => {
    if (!a) return undefined;
    const p = delta / a * 100;
    if (Math.abs(p) < 0.05) return undefined;
    return `${sign(p)}${Math.abs(p).toFixed(1)} %`;
};

export interface Sides<I> {
    text: (label: string, pick: (i: I) => string | undefined) => ValueRow,
    flag: (label: string, pick: (i: I) => boolean | undefined) => ValueRow,
    number: (label: string, pick: (i: I) => number | undefined, unit: Unit) => ValueRow,
    list: (label: string, pick: (i: I) => string[] | undefined) => ListRow,
}

const on = <I, R>(i: I | undefined, pick: (i: I) => R): R | undefined => i === undefined ? undefined : pick(i);

// sides builds the rows of two things of the same kind (the android info of
// both bundles) from what to read of each; a side that is missing is empty
export const sides = <I, >(a: I | undefined, b: I | undefined): Sides<I> => ({
    text: (label, pick) => value(label, on(a, pick), on(b, pick)),
    flag: (label, pick) => flag(label, on(a, pick), on(b, pick)),
    number: (label, pick, unit) => number(label, {a: on(a, pick), b: on(b, pick)}, unit),
    list: (label, pick) => list(label, on(a, pick), on(b, pick)),
});
