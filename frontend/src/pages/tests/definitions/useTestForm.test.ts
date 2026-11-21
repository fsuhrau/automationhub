import {act, renderHook} from '@testing-library/react';
import {describe, expect, it} from 'vitest';
import {TestType} from '../../../types/test.type.enum';
import {UnityTestCategory} from '../../../types/unity.test.category.type.enum';
import {aTest} from './harness/fixtures';
import {useTestForm} from './useTestForm';

describe('useTestForm', () => {
    it('starts a new Unity test that may be created', () => {
        const {result} = renderHook(() => useTestForm());
        expect(result.current.unity).toBe(true);
        expect(result.current.saveState).toEqual({creating: true, dirty: false, tried: false, valid: false});
        expect(result.current.errors).toEqual({name: 'The test needs a name.'});
    });

    it('starts from a stored test, which is not dirty until a field changes', () => {
        const test = aTest();
        const {result} = renderHook(() => useTestForm(test));
        expect(result.current.dirty).toBe(false);
        act(() => result.current.set({mode: UnityTestCategory.RunAllOfCategory}));
        expect(result.current.dirty).toBe(true);
        expect(result.current.errors.categories).toBeDefined();
    });

    it('marks a save as tried', () => {
        const {result} = renderHook(() => useTestForm());
        act(() => result.current.markTried());
        expect(result.current.tried).toBe(true);
    });

    it('knows tests of other types are not Unity tests', () => {
        const {result} = renderHook(() => useTestForm(aTest({}, {type: TestType.Cocos})));
        expect(result.current.unity).toBe(false);
    });
});
