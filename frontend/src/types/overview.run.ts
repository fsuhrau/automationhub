import {IRunSummary} from './test';

// a run on the project overview, with its test and app
export interface IOverviewRun extends IRunSummary {
    testId: number,
    testName: string,
    appId: number,
}
