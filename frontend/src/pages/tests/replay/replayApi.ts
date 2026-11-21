import {ApiClient} from '../../../api/ApiClient';
import ITestRunResponseData from '../../../types/test.run.response';

export interface RunKey {
    projectId: string,
    appId: number,
    testId: string,
    runId: string,
}

export const fetchReplayRun = (api: ApiClient, k: RunKey): Promise<ITestRunResponseData> =>
    api.get<ITestRunResponseData>(`/${k.projectId}/app/${k.appId}/test/${k.testId}/run/${k.runId}`);
