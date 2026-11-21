import {useEffect, useMemo, useState} from 'react';
import IDeviceData from '../../types/device';
import {
    addRow, EditForm, formOf, isDirty, keyErrors, newRowId, ParameterRow, removeRow, setAlias, setRow,
} from './deviceEditModel';

export interface DeviceEditForm {
    form: EditForm,
    errors: Map<number, string>,
    dirty: boolean,
    setAlias: (alias: string) => void,
    setRow: (rowId: string, change: Partial<ParameterRow>) => void,
    addRow: () => void,
    removeRow: (rowId: string) => void,
}

// closing or reloading the tab with unsaved changes asks first
const useLeaveWarning = (active: boolean): void => {
    useEffect(() => {
        if (!active) return;
        const warn = (e: BeforeUnloadEvent): void => e.preventDefault();
        window.addEventListener('beforeunload', warn);
        return () => window.removeEventListener('beforeunload', warn);
    }, [active]);
};

// useDeviceEditForm holds the form of a device with what is wrong in it,
// whether it changed, and the changes the page makes
export function useDeviceEditForm(device: IDeviceData): DeviceEditForm {
    const initial = useMemo(() => formOf(device), [device]);
    const [form, setForm] = useState<EditForm>(initial);
    const dirty = isDirty(initial, form);
    useLeaveWarning(dirty);
    return {
        form,
        errors: keyErrors(form.parameters),
        dirty,
        setAlias: alias => setForm(f => setAlias(f, alias)),
        setRow: (rowId, change) => setForm(f => setRow(f, rowId, change)),
        addRow: () => {
            const rowId = newRowId();
            setForm(f => addRow(f, rowId));
        },
        removeRow: rowId => setForm(f => removeRow(f, rowId)),
    };
}
