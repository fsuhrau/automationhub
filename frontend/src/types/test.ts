import ITestConfigData from './test.config';
import ITestRunData from './test.run';

export default interface ITestData {
    ID?: number,
    CompanyID: number,
    Name: string,
    TestConfig: ITestConfigData,
    TestRuns: ITestRunData[],
    Last?: ITestRunData | null,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
