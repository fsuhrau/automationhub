import ITestRunData from './test.run';
import {RunLifecycle} from '../pages/tests/testRunModel';

export default interface ITestRunResponseData {
    nextRunId: number,
    prevRunId: number,
    testRun: ITestRunData,
    // whether the run still executes; missing on older masters
    state?: RunLifecycle,
}
