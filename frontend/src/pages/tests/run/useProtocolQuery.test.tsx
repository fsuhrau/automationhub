import {describe, expect, it} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {Location, useLocation} from 'react-router-dom';
import {IBaselineCandidate} from '../../../types/protocol.baseline';
import {wrapperOf} from '../../../testing/renderWithProviders';
import {ProtocolQuery, useProtocolQuery} from './useProtocolQuery';

interface Probe {
    query: ProtocolQuery,
    location: Location,
}

const setup = (url: string): RenderHookResult<Probe, unknown> =>
    renderHook(() => ({query: useProtocolQuery(), location: useLocation()}), {wrapper: wrapperOf({url})});
const candidate = (deviceId?: number): IBaselineCandidate => ({testRunId: 8, deviceId} as IBaselineCandidate);

describe('useProtocolQuery', () => {
    it('reads the tab and the picked run', () => {
        const {result} = setup('/x?t=log&vs=4&vsDevice=9');
        expect(result.current.query).toMatchObject({tab: 'log', vs: '4', vsDevice: '9'});
    });

    it('is the overview without query', () => {
        expect(setup('/x').result.current.query).toMatchObject({tab: 'overview', vs: null, vsDevice: null});
    });

    it('sets the tab, and removes it for the overview', () => {
        const {result} = setup('/x?vs=4');
        act(() => result.current.query.setTab('log'));
        expect(result.current.location.search).toBe('?vs=4&t=log');
        act(() => result.current.query.setTab('overview'));
        expect(result.current.location.search).toBe('?vs=4');
    });

    it('picks a run, naming its device only when it is another one', () => {
        const {result} = setup('/x?vs=4&vsDevice=9');
        act(() => result.current.query.setBaseline(candidate(5), 7));
        expect(result.current.location.search).toBe('?vs=8&vsDevice=5');
        act(() => result.current.query.setBaseline(candidate(7), 7));
        expect(result.current.location.search).toBe('?vs=8');
        act(() => result.current.query.setBaseline(candidate(), 7));
        expect(result.current.location.search).toBe('?vs=8');
    });

    it('clears the pick', () => {
        const {result} = setup('/x?vs=4&vsDevice=9');
        act(() => result.current.query.setBaseline(null, 7));
        expect(result.current.location.search).toBe('');
    });

    it('keeps the tab and the pick for the next test', () => {
        expect(setup('/x?t=log&vs=4&vsDevice=9').result.current.query.keep()).toBe('?t=log&vs=4&vsDevice=9');
        expect(setup('/x').result.current.query.keep()).toBe('');
    });
});
