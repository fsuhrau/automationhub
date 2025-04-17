import IRealDeviceData from './real.device';
import IRealDeviceConnectionData from './real.device.connection';
import IConnectionParameter from './device.connection.parameter';
import {DeviceType} from './device.type.enum';
import {INodeData} from "./node";
import {PlatformType} from "./platform.type.enum";
import IParameter from './device.parameter';
import {DeviceConnectionType} from "./device.connection.type.enum";
import {DeviceStateType} from "./deviceStateType";

export default interface IDeviceData {
    ID: number,
    NodeID: number,
    Node?: INodeData,
    CompanyID: number,
    DeviceIdentifier: string,
    DeviceType: DeviceType,
    PlatformType: PlatformType,
    Name: string,
    Alias: string,
    OS: string,
    OSVersion: string,
    Status: DeviceStateType,
    IsLocked: boolean,
    Manager: string,
    Dev?: IRealDeviceData | null,
    IsAcknowledged: boolean,
    Connection?: IRealDeviceConnectionData | null,
    ConnectionType: DeviceConnectionType,
    ConnectionParameter: IConnectionParameter,
    CreatedAt?: Date,
    UpdatedAt?: Date,
    DeletedAt?: Date,
    CustomParameter: IParameter[],
    DeviceParameter: IParameter[];
}
