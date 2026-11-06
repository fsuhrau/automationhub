import ITestConfigData from './test.config';
import { PlatformType } from './platform.type.enum';

export default interface ITestData {
    id: number,
    appId: number,
    name: string,
    testConfig: ITestConfigData,
    // the latest runs, newest first (in the tests list)
    recentRuns?: IRunSummary[],
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}

export type RunSummaryStatus = 'running' | 'passed' | 'failed' | 'unstable' | 'interrupted' | 'not_run' | 'empty';

// a run in the tests list: when it ran and how it went
export interface IRunSummary {
    id: number,
    createdAt: Date,
    state: 'running' | 'finished' | 'not_running',
    status: RunSummaryStatus,
    passed: number,
    failed: number,
    unstable: number,
    open: number,
}
