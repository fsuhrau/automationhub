import ITestParameterData from "./test.parameter";
import ITestLogData from "./test.log";
import ITestResultData from "./test.result";

export default interface ITestRunData {
    id?: number | null,
    TestID: number,
    Parameter: ITestParameterData,
    Log: ITestLogData,
    Result: ITestResultData
}