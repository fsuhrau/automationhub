import { TestType } from './test.type.enum';
import IAppFunctionData from './app.function';

export default interface ICreateTestData {
    Name: string,
    TestType: TestType,
    UnityAllTests: boolean,
    UnitySelectedTests: IAppFunctionData[],
    AllDevices: boolean,
    SelectedDevices: number[],
}