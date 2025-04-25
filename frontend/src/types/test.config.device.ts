import IDeviceData from './device';

export default interface ITestConfigDeviceData {
    id?: number,
    testConfigId: number,
    deviceId: number,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
    device?: IDeviceData | null,
}
