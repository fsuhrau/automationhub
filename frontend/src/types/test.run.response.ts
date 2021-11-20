import ITestRunData from './test.run';

export default interface ITestRunResponseData {
    NextRunId: number,
    PrevRunId: number,
    TestRun: ITestRunData,
}
