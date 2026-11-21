import {IRunSummary} from './test';

export interface ITestRunListEntry extends Omit<IRunSummary, 'status'> {
    // undefined when an older master doesn't summarize runs
    status?: IRunSummary['status'],
    parameter: string,
    bundle: { id: number, name: string, version: string } | null,
    endedAt: Date | null,
}

export interface ITestRunsPage {
    runs: ITestRunListEntry[],
    // older runs exist; load them with before = the last run's id
    more: boolean,
}
