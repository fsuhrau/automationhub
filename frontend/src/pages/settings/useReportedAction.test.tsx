import React from 'react';
import {expect, it} from 'vitest';
import {act, renderHook, screen} from '@testing-library/react';
import {ErrorProvider} from '../../ErrorProvider';
import {useReportedAction} from './useReportedAction';

const wrapper = ({children}: { children: React.ReactNode }): React.ReactElement => <ErrorProvider>{children}</ErrorProvider>;

it('useReportedAction: shows a failure in the error toast and keeps it on the action', async () => {
    const {result} = renderHook(() => useReportedAction(async () => {
        throw new Error('boom');
    }), {wrapper});
    await act(() => result.current.run());
    expect(result.current.error).toBe('boom');
    expect(screen.getByRole('alert')).toHaveTextContent('boom');
});

it('useReportedAction: shows nothing when the action works', async () => {
    const {result} = renderHook(() => useReportedAction(async () => 5), {wrapper});
    await expect(act(() => result.current.run())).resolves.toBe(5);
    expect(screen.queryByRole('alert')).toBeNull();
});
