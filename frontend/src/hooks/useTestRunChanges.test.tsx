import {act, render} from '@testing-library/react';
import React from 'react';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {useTestRunChanges} from './useTestRunChanges';
import {resetSse, sendSse} from '../app/testing/fakeSse';

vi.mock('react-hooks-sse', async () => (await import('../app/testing/fakeSse')).fakeSseModule);

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

beforeEach(() => {
    vi.useFakeTimers();
    resetSse();
});
afterEach(() => vi.useRealTimers());

describe('useTestRunChanges', () => {
    it('does not reload without a change', () => {
        const reload = vi.fn();
        render(<Watcher reload={reload}/>);
        flush(1000);
        expect(reload).not.toHaveBeenCalled();
    });

    it('reloads once after a burst of changes of the project', () => {
        const reload = vi.fn();
        render(<Watcher reload={reload}/>);
        sendSse('test_runs', change());
        sendSse('test_runs', change({state: 'finished'}));
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
        render(<Watcher reload={reload}/>);
        sendSse('test_runs', data);
        flush(1000);
        expect(reload).not.toHaveBeenCalled();
    });

    it('reloads only for the given app', () => {
        const reload = vi.fn();
        render(<Watcher appId={3} reload={reload}/>);
        sendSse('test_runs', change({appId: 4}));
        flush(1000);
        expect(reload).not.toHaveBeenCalled();
        sendSse('test_runs', change({appId: 3}));
        flush(300);
        expect(reload).toHaveBeenCalledTimes(1);
    });

    it('calls the latest reload function', () => {
        const first = vi.fn();
        const second = vi.fn();
        const {rerender} = render(<Watcher reload={first}/>);
        sendSse('test_runs', change());
        rerender(<Watcher reload={second}/>);
        flush(300);
        expect(first).not.toHaveBeenCalled();
        expect(second).toHaveBeenCalledTimes(1);
    });
});
