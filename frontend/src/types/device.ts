import IRealDeviceData from './real.device';
import IRealDeviceConnectionData from './real.device.connection';
import IConnectionParameter from './device.connection.parameter';
import {DeviceType} from './device.type.enum';
import {INodeData} from "./node";
import {PlatformType} from "./platform.type.enum";
import IParameter from './device.parameter';

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
    HardwareModel: string,
    OS: string,
    OSVersion: string,
    Status: number,
    Manager: string,
    Dev?: IRealDeviceData | null,
    IsAcknowledged: boolean,
    Connection?: IRealDeviceConnectionData | null,
    ConnectionParameter: IConnectionParameter,
    CreatedAt?: Date,
    UpdatedAt?: Date,
    DeletedAt?: Date,
    CustomParameter: IParameter[],
    DeviceParameter: IParameter[];
}


export const deviceState = (state: number): string => {
    switch (state) {
        case 0:
            return 'null';
        case 1:
            return 'unknown';
        case 2:
            return 'shutdown';
        case 3:
            return 'disconnected';
        case 4:
            return 'booted';
        case 5:
            return 'locked';
    }
    return '';
};