import IDeviceData from './device';

export default interface ITestConfigDeviceData {
    ID?: number,
    TestConfigID: number,
    DeviceID: number,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
    Device?: IDeviceData | null,
}
