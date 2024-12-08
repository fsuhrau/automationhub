import { TestType } from './test.type.enum';
import IAppFunctionData from './app.function';
import { TestExecutionType } from './test.execution.type.enum';
import { PlatformType } from './platform.type.enum';
import { UnityTestCategory } from "./unity.test.category.type.enum";

export default interface ICreateTestData {
    Name: string,
    TestType: TestType,
    UnityTestCategoryType: UnityTestCategory,
    ExecutionType: TestExecutionType,
    Categories: string[],
    UnitySelectedTests: IAppFunctionData[],
    AllDevices: boolean,
    SelectedDevices: number[],
}