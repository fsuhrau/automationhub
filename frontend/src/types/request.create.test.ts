import { TestType } from './test.type.enum';
import IAppFunctionData from './app.function';
import { TestExecutionType } from './test.execution.type.enum';

export default interface ICreateTestData {
    Name: string,
    TestType: TestType,
    UnityAllTests: boolean,
    ExecutionType: TestExecutionType,
    UnitySelectedTests: IAppFunctionData[],
    AllDevices: boolean,
    SelectedDevices: number[],
}