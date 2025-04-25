import ITestRunData from './test.run';

export default interface ITestRunResponseData {
    nextRunId: number,
    prevRunId: number,
    testRun: ITestRunData,
}
