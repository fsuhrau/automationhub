import {expect, it} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {useNameField} from './useNameField';

it('useNameField: is not invalid before a submit was tried', () => {
    const {result} = renderHook(() => useNameField());
    expect(result.current.invalid).toBe(false);
});

it('useNameField: is invalid after a submit with a blank name', () => {
    const {result} = renderHook(() => useNameField());
    act(() => result.current.setName('  '));
    let ok = true;
    act(() => {
        ok = result.current.validate();
    });
    expect(ok).toBe(false);
    expect(result.current.invalid).toBe(true);
});

it('useNameField: validates a filled name and exposes it trimmed', () => {
    const {result} = renderHook(() => useNameField());
    act(() => result.current.setName(' ci '));
    expect(result.current.trimmed).toBe('ci');
    let ok = false;
    act(() => {
        ok = result.current.validate();
    });
    expect(ok).toBe(true);
});

it('useNameField: resets the name and the tried state', () => {
    const {result} = renderHook(() => useNameField());
    act(() => {
        result.current.validate();
    });
    act(() => result.current.reset());
    expect(result.current).toMatchObject({name: '', invalid: false});
});
