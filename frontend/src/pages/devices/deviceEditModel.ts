// The device edit form: its rows, what makes it invalid (the same rules the
// master checks), and whether anything changed.

export interface ParameterRow {
    // stable key for React; rows are edited, added and removed by it
    rowId: string,
    // the stored parameter's id, absent for new rows
    id?: number,
    key: string,
    value: string,
}

export interface EditForm {
    alias: string,
    parameters: ParameterRow[],
}

// keyErrors returns an error per row index that has one: a missing key, or
// a key that an earlier row already has
export const keyErrors = (rows: ParameterRow[]): Map<number, string> => {
    const errors = new Map<number, string>();
    const seen = new Set<string>();
    rows.forEach((r, i) => {
        const key = r.key.trim();
        if (key === '') {
            errors.set(i, 'A key is needed');
        } else if (seen.has(key)) {
            errors.set(i, 'This key is already there');
        }
        seen.add(key);
    });
    return errors;
};

const comparable = (f: EditForm) => JSON.stringify({
    alias: f.alias.trim(),
    parameters: f.parameters.map(p => [p.id ?? null, p.key.trim(), p.value]),
});

export const isDirty = (initial: EditForm, current: EditForm): boolean => comparable(initial) !== comparable(current);

// toRequest is the body of the device update
export const toRequest = (f: EditForm) => ({
    alias: f.alias.trim(),
    customParameter: f.parameters.map(p => ({...(p.id ? {id: p.id} : {}), key: p.key.trim(), value: p.value})),
});
