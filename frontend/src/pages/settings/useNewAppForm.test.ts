import {expect, it} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {PlatformType} from '../../types/platform.type.enum';
import {useNewAppForm} from './useNewAppForm';

it('useNewAppForm: starts with an Android app and shows no errors yet', () => {
    const {result} = renderHook(() => useNewAppForm());
    expect(result.current.platform).toBe(PlatformType.Android);
    expect(result.current.errors).toEqual({});
});

it('useNewAppForm: shows the errors after a failed submit', () => {
    const {result} = renderHook(() => useNewAppForm());
    let ok = true;
    act(() => {
        ok = result.current.validate();
    });
    expect(ok).toBe(false);
    expect(Object.keys(result.current.errors)).toEqual(['name', 'identifier']);
});

it('useNewAppForm: builds the trimmed request of a valid form', () => {
    const {result} = renderHook(() => useNewAppForm());
    act(() => {
        result.current.setName(' Game ');
        result.current.setIdentifier(' com.x ');
        result.current.setPlatform(PlatformType.iOS);
    });
    let ok = false;
    act(() => {
        ok = result.current.validate();
    });
    expect(ok).toBe(true);
    expect(result.current.request()).toEqual({name: 'Game', platform: PlatformType.iOS, identifier: 'com.x'});
});
