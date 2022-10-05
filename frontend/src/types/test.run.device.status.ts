import IDeviceData from "./device";

export default interface ITesRunDeviceStatusData {
    ID?: number,
    TestRunID: number,
    DeviceID: number,
    Device: IDeviceData,
    StartupTime: number,
    HistAvgStartupTime: number,
    CreatedAt: Date,
    UpdatedAt: Date,
    DeletedAt: Date,
}
