import {useEffect} from 'react';
import {useApiClient} from '../../api/ApiClientProvider';
import {useLoader} from '../../hooks/useLoader';
import {useTestRunChanges} from '../../hooks/useTestRunChanges';
import {errorMessage} from '../../lib/errorMessage';
import {OverviewMissing, overviewService} from '../../services/overview.service';
import {IProjectOverview} from '../../types/overview.project';

// runs change while the page is open; refresh while something runs
const REFRESH_WHILE_RUNNING_MS = 15000;

export interface OverviewState {
    overview: IProjectOverview | null,
    // why it could not be loaded
    error: string | null,
    // the master predates the overview
    outdated: boolean,
    retry: () => void,
}

// useOverview loads the overview of the project and keeps it current
export const useOverview = (projectIdentifier: string): OverviewState => {
    const api = useApiClient();
    const {data, error, retry} = useLoader(() => overviewService(api).get(projectIdentifier), [api, projectIdentifier]);
    // a run started or finished, e.g. from MCP or the CLI
    useTestRunChanges(projectIdentifier, null, retry);
    const running = (data?.running.length ?? 0) > 0;
    useEffect(() => {
        if (!running) return;
        const timer = setInterval(retry, REFRESH_WHILE_RUNNING_MS);
        return () => clearInterval(timer);
    }, [running, retry]);

    const outdated = error instanceof OverviewMissing;
    return {overview: data, error: error && !outdated ? errorMessage(error) : null, outdated, retry};
};
