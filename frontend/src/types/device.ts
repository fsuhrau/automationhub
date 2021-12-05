import IRealDeviceData from './real.device';
import IRealDeviceConnectionData from './real.device.connection';

export default interface IDeviceData {
    ID: number,
    CompanyID: number,
    DeviceIdentifier: string,
    DeviceType: number,
    Name: string,
    HardwareModel: string,
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
    Connection?: IRealDeviceConnectionData | null,
    CreatedAt?: Date,
    UpdatedAt?: Date,
    DeletedAt?: Date,
}