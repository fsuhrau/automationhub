import ITestConfigUnityData from './test.config.unity';
import { TestExecutionType } from './test.execution.type.enum';
import { TestType } from './test.type.enum';
import ITestConfigDeviceData from './test.config.device';

export default interface ITestConfigData {
    id?: number,
    testId: number,
    executionType: TestExecutionType,
    type: TestType,
    allDevices: boolean,
    // records the screen of each device during a run
    recordSession?: boolean,
    devices: ITestConfigDeviceData[]
    unity?: ITestConfigUnityData | null,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
