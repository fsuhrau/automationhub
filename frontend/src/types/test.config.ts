import ITestConfigUnityData from './test.config.unity';
import { TestExecutionType } from './test.execution.type.enum';
import { TestType } from './test.type.enum';
import ITestConfigDeviceData from './test.config.device';
import { PlatformType } from './platform.type.enum';
import { UnityTestCategory } from "./unity.test.category.type.enum";

export default interface ITestConfigData {
    ID?: number,
    TestID: number,
    ExecutionType: TestExecutionType,
    Type: TestType,
    AllDevices: boolean,
    Devices: ITestConfigDeviceData[]
    Unity?: ITestConfigUnityData | null,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
