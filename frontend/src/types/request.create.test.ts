import { TestType } from './test.type.enum';
import IAppFunctionData from './app.function';
import { TestExecutionType } from './test.execution.type.enum';
import { PlatformType } from './platform.type.enum';
import { UnityTestCategory } from "./unity.test.category.type.enum";

export default interface ICreateTestData {
    name: string,
    testType: TestType,
    unityTestCategoryType: UnityTestCategory,
    executionType: TestExecutionType,
    categories: string[],
    unitySelectedTests: IAppFunctionData[],
    allDevices: boolean,
    selectedDevices: number[],
}