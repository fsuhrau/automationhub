import IProtocolEntryData from './protocol.entry';
import { TestResultState } from './test.result.state.enum';
import IDeviceData from './device';

export default interface ITestProtocolData {
    ID?: number | null,
    TestRunID: number,
    DeviceID?: number | null,
    Device?: IDeviceData | null,
    TestName: string,
    StartedAt: Date,
    EndedAt?: Date,
    Entries: IProtocolEntryData[]
    TestResult: TestResultState
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
