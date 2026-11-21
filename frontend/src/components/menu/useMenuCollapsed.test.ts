import {act, renderHook} from '@testing-library/react';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {useMenuCollapsed} from './useMenuCollapsed';

beforeEach(() => localStorage.clear());
afterEach(() => vi.restoreAllMocks());

describe('useMenuCollapsed', () => {
    it('starts expanded', () => {
        expect(renderHook(() => useMenuCollapsed()).result.current.collapsed).toBe(false);
    });

    it('starts as it was left', () => {
        localStorage.setItem('mainMenuCollapsed', 'true');
        expect(renderHook(() => useMenuCollapsed()).result.current.collapsed).toBe(true);
    });

    it('toggles, expands and remembers', () => {
        const {result} = renderHook(() => useMenuCollapsed());
        act(() => result.current.toggle());
        expect(result.current.collapsed).toBe(true);
        expect(localStorage.getItem('mainMenuCollapsed')).toBe('true');
        act(() => result.current.expand());
        expect(result.current.collapsed).toBe(false);
        expect(localStorage.getItem('mainMenuCollapsed')).toBe('false');
    });

    it('still works when the browser storage is blocked', () => {
        vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('blocked');
        });
        vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('blocked');
        });
        const {result} = renderHook(() => useMenuCollapsed());
        expect(result.current.collapsed).toBe(false);
        act(() => result.current.toggle());
        expect(result.current.collapsed).toBe(true);
    });
});
