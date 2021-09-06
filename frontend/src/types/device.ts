import IRealDeviceData from './real.device';

export default interface IDeviceData {
    ID?: number | null,
    CompanyID: number,
    DeviceIdentifier: string,
    DeviceType: number,
    Name: string,
    RAM: number,
    SOC: string,
    DisplaySize: string,
    DPI: number,
    OS: string,
    OSVersion: string,
    GPU: string,
    ABI: string,
    OpenGLESVersion: number,
    Status: number,
    Manager: string,
    Dev?: IRealDeviceData | null,
    CreatedAt?: Date,
    UpdatedAt?: Date,
    DeletedAt?: Date,
}