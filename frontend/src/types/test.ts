import ITestConfigData from "./test.config";
import ITestRunData from "./test.run";

export default interface ITestData {
    id?: number | null,
    CompanyID: number,
    Name: string,
    TestConfig: ITestConfigData,
    TestRuns: ITestRunData[],
    Last: ITestRunData,
}