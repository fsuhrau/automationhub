import ITestConfigUnityData from './test.config.unity';
import { TestExecutionType } from './test.execution.type.enum';
import { TestType } from './test.type.enum';

export default interface ITestConfigData {
    ID?: number,
    TestID: number,
    ExecutionType: TestExecutionType,
    Type: TestType,
    Unity?: ITestConfigUnityData | null,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
