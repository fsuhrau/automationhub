import IDeviceData from "./device";

export default interface ITesRunDeviceStatusData {
    id?: number,
    testRunId: number,
    deviceId: number,
    device: IDeviceData,
    startupTime: number,
    histAvgStartupTime: number,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date,
}
