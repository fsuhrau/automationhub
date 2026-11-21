import {errorMessage} from '../../../lib/errorMessage';
import {useLoader} from '../../../hooks/useLoader';
import {useTestRunChanges} from '../../../hooks/useTestRunChanges';
import ITestData from '../../../types/test';
import {useTestService} from '../../../services/test.service';

export interface TestsList {
    // null while loading
    tests: ITestData[] | null,
    error: string | null,
    reload: () => void,
}

// useTestsList loads the tests of an app and reloads them when a run of the
// app started or finished, e.g. from MCP or the CLI
export const useTestsList = (projectId: string, appId: number | null): TestsList => {
    const service = useTestService();
    const {data, error, retry} = useLoader(
        () => appId === null ? null : service.list({projectId, appId}), [service, projectId, appId]);
    useTestRunChanges(projectId, appId, retry);
    return {tests: data, error: error ? errorMessage(error) : null, reload: retry};
};
