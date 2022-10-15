import ITestConfigData from './test.config';
import ITestRunData from './test.run';
import { PlatformType } from './platform.type.enum';

export default interface ITestData {
    ID: number,
    AppID: number,
    Name: string,
    TestConfig: ITestConfigData,
    TestRuns: ITestRunData[],
    Last?: ITestRunData | null,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
