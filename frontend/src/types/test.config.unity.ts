import IUnityTestFunctionData from './unity.test.function';

export default interface ITestConfigUnityData {
    ID?: number | null,
    TestConfigID: number,
    RunAllTests: boolean,
    UnityTestFunctions: Array<IUnityTestFunctionData>,
}
