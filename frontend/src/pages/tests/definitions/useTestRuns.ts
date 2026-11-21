import {useCallback, useEffect, useState} from 'react';
import {useAsyncAction} from '../../../hooks/useAsyncAction';
import {useTestRunChanges} from '../../../hooks/useTestRunChanges';
import {useTestRunService} from '../../../services/test.run.service';
import {ITestRunListEntry} from '../../../types/test.run.list';

export interface TestRuns {
    // null until the first page is loaded
    runs: ITestRunListEntry[] | null,
    // older runs exist
    more: boolean,
    loading: boolean,
    error: string | null,
    reload: () => void,
    loadOlder: () => void,
}

// useTestRuns pages through the runs of a test, newest first, and reloads
// them when a run of the app started or finished
export const useTestRuns = (projectId: string, appId: number | null, testId: number): TestRuns => {
    const service = useTestRunService();
    const [runs, setRuns] = useState<ITestRunListEntry[] | null>(null);
    const [more, setMore] = useState(false);
    const fetchPage = useCallback(async (before?: number): Promise<void> => {
        if (appId === null) return;
        const page = await service.runsPage({projectId, appId, testId}, {before});
        setRuns(prev => before && prev ? [...prev, ...page.runs] : page.runs);
        setMore(page.more);
    }, [service, projectId, appId, testId]);
    const {run, busy, error} = useAsyncAction(fetchPage);

    const reload = useCallback(() => void run(), [run]);
    useEffect(reload, [reload]);
    useTestRunChanges(projectId, appId, reload);
    const loadOlder = (): void => void run(runs?.[runs.length - 1]?.id);
    return {runs, more, loading: busy, error, reload, loadOlder};
};
