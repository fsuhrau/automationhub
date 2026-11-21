import {IAppBinaryData} from '../../types/app';
import {previousOf} from './bundleCompareModel';

export interface Selection {
    a?: number,
    b?: number,
}

// selectionOf says which two bundles to compare. Without ids in the url it
// is the newest bundle against the one before it; with only the second, that
// one against the one before it (the oldest against the next).
export const selectionOf = (bundles: IAppBinaryData[] | null, a: string | null, b: string | null): Selection => {
    const to = Number(b) || bundles?.[0]?.id;
    if (Number(a)) return {a: Number(a), b: to};
    if (!bundles || !to) return {b: to};
    return {a: (previousOf(bundles, to) ?? bundles.find(x => x.id !== to))?.id, b: to};
};

// withPick is the url selection after picking a bundle for one side
export const withPick = (current: Selection, which: 'a' | 'b', id: number): Record<'a' | 'b', string> =>
    ({a: String(current.a ?? ''), b: String(current.b ?? ''), [which]: String(id)});

// swapped is the url selection with both sides exchanged
export const swapped = ({a, b}: Selection): Record<'a' | 'b', string> => ({a: String(b ?? ''), b: String(a ?? '')});

// pendingPair says the details of two bundles to compare are still on their way
export const pendingPair = (bundles: IAppBinaryData[] | null, loaded: boolean, failed: boolean): boolean =>
    (bundles?.length ?? 0) >= 2 && !loaded && !failed;
