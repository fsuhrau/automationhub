// The device edit form: its rows, what makes it invalid (the same rules the
// master checks), whether anything changed, and the changes of its rows.
import IDeviceData from '../../types/device';
import {DeviceUpdate} from '../../types/device.requests';

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

const comparable = (f: EditForm): string => JSON.stringify({
    alias: f.alias.trim(),
    parameters: f.parameters.map(p => [p.id ?? null, p.key.trim(), p.value]),
});

export const isDirty = (initial: EditForm, current: EditForm): boolean => comparable(initial) !== comparable(current);

// toRequest is the body of the device update
export const toRequest = (f: EditForm): DeviceUpdate => ({
    alias: f.alias.trim(),
    customParameter: f.parameters.map(p => ({...(p.id ? {id: p.id} : {}), key: p.key.trim(), value: p.value})),
});

let rowCounter = 0;

// a key for a new row, unique for the session
export const newRowId = (): string => `row-${rowCounter++}`;

// the form of a stored device, one row per custom parameter
export const formOf = (device: IDeviceData, rowId: () => string = newRowId): EditForm => ({
    alias: device.alias ?? '',
    parameters: (device.customParameter ?? []).map(p => ({rowId: rowId(), id: p.id, key: p.key, value: p.value})),
});

export const setAlias = (form: EditForm, alias: string): EditForm => ({...form, alias});

export const setRow = (form: EditForm, rowId: string, change: Partial<ParameterRow>): EditForm =>
    ({...form, parameters: form.parameters.map(p => p.rowId === rowId ? {...p, ...change} : p)});

export const addRow = (form: EditForm, rowId: string): EditForm =>
    ({...form, parameters: [...form.parameters, {rowId, key: '', value: ''}]});

export const removeRow = (form: EditForm, rowId: string): EditForm =>
    ({...form, parameters: form.parameters.filter(p => p.rowId !== rowId)});

// the listed device keeps its live state; only what was edited changes
export const applyEdit = (listed: IDeviceData, saved: IDeviceData): IDeviceData =>
    ({...listed, alias: saved.alias, customParameter: saved.customParameter});
