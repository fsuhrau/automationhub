import ITestRunData from './test.run';
import {RunLifecycle} from './run.lifecycle';

export default interface ITestRunResponseData {
    nextRunId: number,
    prevRunId: number,
    testRun: ITestRunData,
    // whether the run still executes; missing on older masters
    state?: RunLifecycle,
    // devices a running run holds now; missing on older masters
    devices?: number[],
}
