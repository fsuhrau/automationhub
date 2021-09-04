import ITestParameterData from './test.parameter';
import ITestLogData from './test.log';
import ITestResultData from './test.result';

export default interface ITestRunData {
    ID?: number,
    TestID: number,
    Parameter: ITestParameterData,
    Log: ITestLogData,
    Result: ITestResultData
}
