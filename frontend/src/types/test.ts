import ITestConfigData from './test.config';
import ITestRunData from './test.run';
import { PlatformType } from './platform.type.enum';

export default interface ITestData {
    id: number,
    appId: number,
    name: string,
    testConfig: ITestConfigData,
    testRuns: ITestRunData[],
    last?: ITestRunData | null,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
