import {SizeEntry} from '../../types/bundle.info';
import {SizeRow} from '../../types/bundle.compare';

const merge = (a: SizeEntry[], b: SizeEntry[]): Map<string, SizeRow> => {
    const byPath = new Map<string, SizeRow>();
    for (const e of a) byPath.set(e.path, {path: e.path, a: e.size, delta: -e.size, change: 'removed'});
    for (const e of b) {
        const row = byPath.get(e.path);
        if (!row) {
            byPath.set(e.path, {path: e.path, b: e.size, delta: e.size, change: 'added'});
            continue;
        }
        row.b = e.size;
        row.delta = e.size - (row.a ?? 0);
        row.change = row.delta === 0 ? 'same' : 'changed';
    }
    return byPath;
};

const weight = (r: SizeRow): number => r.delta === undefined ? -1 : Math.abs(r.delta);

// sizes compares two size lists by path. A complete list (the parts) misses
// an entry when it isn't there; a partial one (the largest files) only when
// it is smaller, so the change of such an entry is unknown.
export const sizes = (a: SizeEntry[] | undefined, b: SizeEntry[] | undefined, partial = false): SizeRow[] => {
    const rows = [...merge(a ?? [], b ?? []).values()];
    if (partial) {
        rows.filter(r => r.change === 'added' || r.change === 'removed').forEach(r => {
            r.delta = undefined;
        });
    }
    return rows.sort((x, y) => weight(y) - weight(x) || x.path.localeCompare(y.path));
};
