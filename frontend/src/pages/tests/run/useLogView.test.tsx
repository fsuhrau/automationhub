import {describe, expect, it} from 'vitest';
import {act, renderHook} from '@testing-library/react';
import {entry} from './testkit/fixtures';
import {useLogView} from './useLogView';

const log = Array.from({length: 600}, (_, i) => ({...entry({id: i + 1, message: i === 599 ? 'last boom' : `line ${i}`, level: i === 599 ? 'error' : 'info'})}));

describe('useLogView', () => {
    it('shows a page of the log and more on request', () => {
        const {result} = renderHook(() => useLogView(log));
        expect(result.current.limit).toBe(500);
        act(() => result.current.showMore());
        expect(result.current.limit).toBe(1000);
    });

    it('filters by text and errors, and resets the filters', () => {
        const {result} = renderHook(() => useLogView(log));
        act(() => result.current.setSearch('boom'));
        expect(result.current.visible).toHaveLength(1);
        act(() => result.current.setErrorsOnly(true));
        act(() => result.current.toggleHidden('step'));
        expect(result.current.visible).toHaveLength(0);
        act(() => result.current.resetFilters());
        expect(result.current).toMatchObject({search: '', errorsOnly: false});
        expect(result.current.hidden.size).toBe(0);
        expect(result.current.visible).toHaveLength(600);
    });

    it('toggles a hidden source on and off', () => {
        const {result} = renderHook(() => useLogView(log));
        act(() => result.current.toggleHidden('step'));
        expect(result.current.hidden.has('step')).toBe(true);
        act(() => result.current.toggleHidden('step'));
        expect(result.current.hidden.has('step')).toBe(false);
    });

    it('shows an entry: clears the filters, extends the page to it and focuses it', () => {
        const {result} = renderHook(() => useLogView(log));
        act(() => result.current.setSearch('boom'));
        act(() => result.current.showEntry(log[599]));
        expect(result.current).toMatchObject({search: '', focus: 600, limit: 649});
        act(() => result.current.showEntry(log[0]));
        expect(result.current.limit).toBe(649);
    });
});
