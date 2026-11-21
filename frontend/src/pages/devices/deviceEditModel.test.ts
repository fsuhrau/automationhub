import {describe, expect, it} from 'vitest';
import {
    addRow, applyEdit, EditForm, formOf, isDirty, keyErrors, newRowId, removeRow, setAlias, setRow, toRequest,
} from './deviceEditModel';
import {device} from './testKit';

const form: EditForm = {
    alias: 'a',
    parameters: [{rowId: 'r1', id: 3, key: 'PIN', value: '1'}, {rowId: 'r2', key: 'LANG', value: 'de'}],
};

describe('keyErrors', () => {
    it('flags missing and repeated keys by row', () => {
        const rows = [
            {rowId: 'a', key: ' ', value: ''}, {rowId: 'b', key: 'X', value: ''}, {rowId: 'c', key: ' X ', value: ''},
        ];
        expect([...keyErrors(rows)]).toEqual([[0, 'A key is needed'], [2, 'This key is already there']]);
        expect(keyErrors(form.parameters).size).toBe(0);
    });
});

describe('isDirty', () => {
    it('ignores spaces around alias and key but sees every other change', () => {
        expect(isDirty(form, {...form, alias: ' a '})).toBe(false);
        expect(isDirty(form, setAlias(form, 'b'))).toBe(true);
        expect(isDirty(form, setRow(form, 'r1', {value: '2'}))).toBe(true);
        expect(isDirty(form, removeRow(form, 'r2'))).toBe(true);
    });
});

describe('toRequest', () => {
    it('trims, keeps stored ids and leaves ids off new rows', () => {
        expect(toRequest({...form, alias: ' a '})).toEqual({
            alias: 'a', customParameter: [{id: 3, key: 'PIN', value: '1'}, {key: 'LANG', value: 'de'}],
        });
    });
});

describe('formOf', () => {
    it('makes a row per custom parameter with fresh row ids', () => {
        const d = device({alias: 'x', customParameter: [{id: 3, key: 'K', value: 'V'}]});
        let n = 0;
        expect(formOf(d, () => `id${n++}`)).toEqual({alias: 'x', parameters: [{rowId: 'id0', id: 3, key: 'K', value: 'V'}]});
    });

    it('copes with a device without alias and parameters', () => {
        const bare = {...device(), alias: undefined, customParameter: undefined} as unknown as ReturnType<typeof device>;
        expect(formOf(bare)).toEqual({alias: '', parameters: []});
    });

    it('hands out unique row ids by default', () => {
        expect(newRowId()).not.toBe(newRowId());
    });
});

describe('row changes', () => {
    it('changes one row only', () => {
        expect(setRow(form, 'r2', {key: 'L'}).parameters.map(p => p.key)).toEqual(['PIN', 'L']);
    });

    it('adds an empty row at the end and removes by id', () => {
        expect(addRow(form, 'r3').parameters.at(-1)).toEqual({rowId: 'r3', key: '', value: ''});
        expect(removeRow(form, 'r1').parameters.map(p => p.rowId)).toEqual(['r2']);
    });

    it('does not change the form it gets', () => {
        addRow(form, 'r3');
        expect(form.parameters).toHaveLength(2);
    });
});

describe('applyEdit', () => {
    it('takes alias and parameters from the saved device and the rest from the list', () => {
        const listed = device({alias: 'old', isLocked: true});
        const saved = device({alias: 'new', customParameter: [{key: 'K', value: 'V'}], isLocked: false});
        expect(applyEdit(listed, saved)).toEqual({...listed, alias: 'new', customParameter: saved.customParameter});
    });
});
