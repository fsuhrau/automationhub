import {useEffect, useRef} from 'react';
import {useSSE} from 'react-hooks-sse';

// sent on the "test_runs" channel whenever a run started or finished, whether
// from this UI, the CLI or MCP
export interface TestRunChange {
    projectId: string,
    appId: number,
    testId?: number,
    testRunId: number,
    state: 'started' | 'finished',
}

const RELOAD_DEBOUNCE_MS = 300;

// does the change belong to the project (and, if given, the app)?
const concerns = (c: TestRunChange | null, projectIdentifier: string, appId: number | null | undefined): boolean =>
    c !== null && c.projectId === projectIdentifier && (appId == null || c.appId === appId);

// useTestRunChanges calls reload when a run of the project (and, if given, of
// the app) started or finished. A few runs starting at once reload only once.
export const useTestRunChanges = (projectIdentifier: string, appId: number | null | undefined, reload: () => void): void => {
    const changes = useSSE<number, TestRunChange | null>('test_runs', 0, {
        stateReducer: (count, action) => concerns(action.data, projectIdentifier, appId) ? count + 1 : count,
    });
    const latest = useRef(reload);
    useEffect(() => {
        latest.current = reload;
    });
    useEffect(() => {
        if (changes === 0) return;
        const timer = setTimeout(() => latest.current(), RELOAD_DEBOUNCE_MS);
        return () => clearTimeout(timer);
    }, [changes]);
};
