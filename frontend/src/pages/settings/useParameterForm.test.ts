import {expect, it} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {AppParameter} from '../../types/app';
import {useParameterForm} from './useParameterForm';

const choice: AppParameter = {id: 2, name: 'Market', type: {type: 'option', options: ['eu'], defaultValue: 'eu'}};

it('useParameterForm: starts from the parameter, or empty for a new one', () => {
    expect(renderHook(() => useParameterForm(null, [])).result.current.form.name).toBe('');
    expect(renderHook(() => useParameterForm(choice, [])).result.current.form.name).toBe('Market');
});

it('useParameterForm: adds typed options, commas separating several, and clears the input', () => {
    const {result} = renderHook(() => useParameterForm(null, []));
    act(() => result.current.set({kind: 'option'}));
    act(() => result.current.setOption('a, b,,a'));
    act(() => result.current.addOption());
    expect(result.current.options).toEqual(['a', 'b']);
    expect(result.current.option).toBe('');
});

it('useParameterForm: shows the errors only after a submit was tried', () => {
    const {result} = renderHook(() => useParameterForm(null, []));
    expect(result.current.errors).toEqual({});
    let ok = true;
    act(() => {
        ok = result.current.validate();
    });
    expect(ok).toBe(false);
    expect(result.current.errors.name).toBe('A name is needed.');
});

it('useParameterForm: validates against the siblings, ignoring the edited parameter itself', () => {
    const {result} = renderHook(() => useParameterForm(choice, [choice]));
    let ok = false;
    act(() => {
        ok = result.current.validate();
    });
    expect(ok).toBe(true);
});
