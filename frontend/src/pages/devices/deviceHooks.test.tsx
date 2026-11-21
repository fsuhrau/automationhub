import React from 'react';
import {describe, expect, it, vi} from 'vitest';
import {act, renderHook, screen, waitFor} from '@testing-library/react';
import {useScreenshot} from './useScreenshot';
import {useDeviceEditForm} from './useDeviceEditForm';
import {useDevicePool} from './useDevicePool';
import {useUnlockDevice} from './useUnlockDevice';
import {useHubState} from '../../hooks/HubStateProvider';
import {device, fakeApi, renderPage} from './testKit';

describe('useScreenshot', () => {
    it('takes a new screenshot on refresh and forgets that the last one was broken', () => {
        const now = vi.spyOn(Date, 'now').mockReturnValueOnce(1).mockReturnValueOnce(2);
        const {result} = renderHook(() => useScreenshot('p1', 3));
        expect(result.current.src).toBe('/api/p1/device/3/screenshot?t=1');
        act(() => result.current.markBroken());
        expect(result.current.broken).toBe(true);
        act(() => result.current.refresh());
        expect(result.current).toMatchObject({src: '/api/p1/device/3/screenshot?t=2', broken: false});
        now.mockRestore();
    });
});

describe('useDeviceEditForm', () => {
    const stored = device({alias: 'A', customParameter: [{id: 1, key: 'K', value: 'V'}]});

    it('starts clean with the stored values', () => {
        const {result} = renderHook(() => useDeviceEditForm(stored));
        expect(result.current.dirty).toBe(false);
        expect(result.current.form.alias).toBe('A');
    });

    it('follows alias and row changes, and reports key errors', () => {
        const {result} = renderHook(() => useDeviceEditForm(stored));
        act(() => result.current.setAlias('B'));
        act(() => result.current.addRow());
        expect(result.current.dirty).toBe(true);
        expect(result.current.errors.get(1)).toBe('A key is needed');
        const rowId = result.current.form.parameters[1].rowId;
        act(() => result.current.setRow(rowId, {key: 'N'}));
        expect(result.current.errors.size).toBe(0);
        act(() => result.current.removeRow(rowId));
        expect(result.current.form.parameters).toHaveLength(1);
    });
});

const Probe: React.FC<{ onHooks: (h: { pool: ReturnType<typeof useDevicePool>, unlock: ReturnType<typeof useUnlockDevice> }) => void }> = ({onHooks}) => {
    const pool = useDevicePool();
    const unlock = useUnlockDevice();
    const {state} = useHubState();
    onHooks({pool, unlock});
    return <div data-testid="count">{state.devices?.map(d => `${d.id}:${d.isLocked}`).join(',') ?? 'none'}</div>;
};

describe('useDevicePool and useUnlockDevice', () => {
    it('put the loaded devices into the hub state, and a device free on unlock', async () => {
        const api = fakeApi({'GET /p1/devices': [device({id: 1, isLocked: true})], 'POST /p1/device/1/unlock': {}});
        let hooks!: Parameters<React.ComponentProps<typeof Probe>['onHooks']>[0];
        renderPage(<Probe onHooks={h => {
            hooks = h;
        }}/>, {api, route: 'x', url: '/project/p1/x'});
        await waitFor(() => expect(screen.getByTestId('count')).toHaveTextContent('1:true'));
        await act(async () => {
            await hooks.unlock.run(1);
        });
        expect(screen.getByTestId('count')).toHaveTextContent('1:false');
        expect(hooks.pool.error).toBeNull();
    });
});
