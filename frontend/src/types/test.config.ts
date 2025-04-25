import ITestConfigUnityData from './test.config.unity';
import { TestExecutionType } from './test.execution.type.enum';
import { TestType } from './test.type.enum';
import ITestConfigDeviceData from './test.config.device';
import { PlatformType } from './platform.type.enum';
import { UnityTestCategory } from "./unity.test.category.type.enum";

export default interface ITestConfigData {
    id?: number,
    testId: number,
    executionType: TestExecutionType,
    type: TestType,
    allDevices: boolean,
    devices: ITestConfigDeviceData[]
    unity?: ITestConfigUnityData | null,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
