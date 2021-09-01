import ITestConfigUnityData from "./test.config.unity";

export default interface ITestConfigData {
    ID?: number | null,
    TestID: number,
    Type: number,
    ExecutionType: number,
    Unity?: ITestConfigUnityData | null,
}