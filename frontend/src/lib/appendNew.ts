export interface Identified {
    id: string | number,
}

// appendNew appends the items whose id is not in the list yet and keeps
// only the last `limit` entries. It returns the same list if nothing is new.
export const appendNew = <T extends Identified>(list: T[], items: T[], limit = Infinity): T[] => {
    const known = new Set(list.map(e => e.id));
    const fresh = items.filter(item => !known.has(item.id) && known.add(item.id));
    if (fresh.length === 0) return list;
    const next = list.concat(fresh);
    return next.length > limit ? next.slice(next.length - limit) : next;
};
