import {act, render} from '@testing-library/react';
import React from 'react';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {useTestRunChanges} from './useTestRunChanges';
import {FakeSse} from '../testing/fakeSse';
import {Providers, renderWithProviders} from '../testing/renderWithProviders';

const Watcher: React.FC<{ appId?: number | null, reload: () => void }> = ({appId, reload}) => {
    useTestRunChanges('p1', appId, reload);
    return null;
};

const change = (fields: object = {}): object => ({projectId: 'p1', appId: 3, testRunId: 9, state: 'started', ...fields});

const flush = (ms: number): void => {
    act(() => {
        vi.advanceTimersByTime(ms);
    });
};

const sse = new FakeSse();

beforeEach(() => {
    vi.useFakeTimers();
});
afterEach(() => vi.useRealTimers());

describe('useTestRunChanges', () => {
    it('does not reload without a change', () => {
        const reload = vi.fn();
        renderWithProviders(<Watcher reload={reload}/>, {sse});
        flush(1000);
        expect(reload).not.toHaveBeenCalled();
    });

    it('reloads once after a burst of changes of the project', () => {
        const reload = vi.fn();
        renderWithProviders(<Watcher reload={reload}/>, {sse});
        sse.emit('test_runs', change());
        sse.emit('test_runs', change({state: 'finished'}));
        flush(300);
        expect(reload).toHaveBeenCalledTimes(1);
    });

});

describe('useTestRunChanges filtering', () => {
    it.each([
        ['another project', change({projectId: 'p2'})],
        ['an event without data', null],
    ])('ignores %s', (_name, data) => {
        const reload = vi.fn();
        renderWithProviders(<Watcher reload={reload}/>, {sse});
        sse.emit('test_runs', data);
        flush(1000);
        expect(reload).not.toHaveBeenCalled();
    });

    it('reloads only for the given app', () => {
        const reload = vi.fn();
        renderWithProviders(<Watcher appId={3} reload={reload}/>, {sse});
        sse.emit('test_runs', change({appId: 4}));
        flush(1000);
        expect(reload).not.toHaveBeenCalled();
        sse.emit('test_runs', change({appId: 3}));
        flush(300);
        expect(reload).toHaveBeenCalledTimes(1);
    });

    it('calls the latest reload function', () => {
        const first = vi.fn();
        const second = vi.fn();
        const watcher = (reload: () => void): React.ReactElement => <Providers sse={sse}><Watcher reload={reload}/></Providers>;
        const {rerender} = render(watcher(first));
        sse.emit('test_runs', change());
        rerender(watcher(second));
        flush(300);
        expect(first).not.toHaveBeenCalled();
        expect(second).toHaveBeenCalledTimes(1);
    });
});
