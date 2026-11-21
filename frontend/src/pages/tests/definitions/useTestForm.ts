import {useCallback, useMemo, useState} from 'react';
import ITestData from '../../../types/test';
import {TestType} from '../../../types/test.type.enum';
import {formOf, FormErrors, isDirty, newForm, SaveState, TestForm, validate} from '../testEditModel';

export interface TestFormState {
    form: TestForm,
    set: (change: Partial<TestForm>) => void,
    errors: FormErrors,
    valid: boolean,
    // the form differs from the test it started from
    dirty: boolean,
    // a save was tried, so the errors are shown
    tried: boolean,
    markTried: () => void,
    saveState: SaveState,
    // the test runs Unity tests (a new test always does)
    unity: boolean,
}

// useTestForm is the state of the edit form: the fields, what is wrong with
// them, and whether they changed. Without a test it starts a new one.
export const useTestForm = (test?: ITestData): TestFormState => {
    const unity = !test || (test.testConfig.type === TestType.Unity && !!test.testConfig.unity);
    const initial = useMemo(() => test ? formOf(test) : newForm(), [test]);
    const [form, setForm] = useState<TestForm>(initial);
    const [tried, setTried] = useState(false);
    const set = useCallback((change: Partial<TestForm>) => setForm(f => ({...f, ...change})), []);
    const markTried = useCallback(() => setTried(true), []);
    const errors = validate(form, unity);
    const valid = Object.keys(errors).length === 0;
    const dirty = isDirty(initial, form);
    return {form, set, errors, valid, dirty, tried, markTried, unity, saveState: {creating: !test, dirty, tried, valid}};
};
