import {renderHook} from '@testing-library/react';
import {describe, expect, it} from 'vitest';
import {useUnsavedWarning} from './useUnsavedWarning';

const leave = (): boolean => {
    const event = new Event('beforeunload', {cancelable: true});
    window.dispatchEvent(event);
    return event.defaultPrevented;
};

describe('useUnsavedWarning', () => {
    it('asks before leaving only while dirty', () => {
        const {rerender} = renderHook(({dirty}) => useUnsavedWarning(dirty), {initialProps: {dirty: false}});
        expect(leave()).toBe(false);
        rerender({dirty: true});
        expect(leave()).toBe(true);
        rerender({dirty: false});
        expect(leave()).toBe(false);
    });

    it('stops asking when the page is gone', () => {
        const {unmount} = renderHook(() => useUnsavedWarning(true));
        unmount();
        expect(leave()).toBe(false);
    });
});
