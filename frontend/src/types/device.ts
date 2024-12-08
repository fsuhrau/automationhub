import IRealDeviceData from './real.device';
import IRealDeviceConnectionData from './real.device.connection';
import IDeviceParameter from './device.parameter';
import IConnectionParameter from './device.connection.parameter';
import { DeviceType } from './device.type.enum';
import { INodeData } from "./node";

export default interface IDeviceData {
    ID: number,
    NodeID: number,
    Node?: INodeData,
    CompanyID: number,
    DeviceIdentifier: string,
    DeviceType: DeviceType,
    Name: string,
    Alias: string,
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
    IsAcknowledged: boolean,
    Connection?: IRealDeviceConnectionData | null,
    ConnectionParameter: IConnectionParameter,
    Parameter: IDeviceParameter[],
    CreatedAt?: Date,
    UpdatedAt?: Date,
    DeletedAt?: Date,
}