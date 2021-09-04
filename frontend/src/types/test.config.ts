import ITestConfigUnityData from './test.config.unity';

export default interface ITestConfigData {
    ID?: number,
    TestID: number,
    Type: number,
    ExecutionType: number,
    Unity?: ITestConfigUnityData | null,
}
