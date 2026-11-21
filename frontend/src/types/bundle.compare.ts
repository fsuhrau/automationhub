// What comparing two bundles produces; see pages/apps/bundleCompareModel.ts.

export type Change = 'same' | 'changed' | 'added' | 'removed';

// ValueRow is one attribute of both bundles
export interface ValueRow {
    kind: 'value',
    label: string,
    a?: string,
    b?: string,
    change: Change,
    // a as a number and b minus a, for numbers like sizes
    from?: number,
    delta?: number,
    // how to show the delta
    unit?: 'bytes' | 'count',
}

// ListRow is a list attribute, like the permissions: what only one of the
// bundles has
export interface ListRow {
    kind: 'list',
    label: string,
    added: string[],
    removed: string[],
    same: number,
    change: Change,
}

export type Row = ValueRow | ListRow;

export interface CompareSection {
    title: string,
    rows: Row[],
}

// SizeRow is one part of the bundle contents in both bundles
export interface SizeRow {
    path: string,
    a?: number,
    b?: number,
    // b minus a; unknown when a list is partial and one side lacks the entry
    delta?: number,
    change: Change,
}

export interface Comparison {
    sections: CompareSection[],
    // the parts of the contents, biggest change first
    parts: SizeRow[],
    // the largest files of both, biggest change first
    files: SizeRow[],
    changes: number,
}
