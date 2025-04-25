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
    id: number,
    nodeId: number,
    node?: INodeData,
    companyId: number,
    deviceIdentifier: string,
    deviceType: DeviceType,
    platformType: PlatformType,
    name: string,
    alias: string,
    os: string,
    osVersion: string,
    status: DeviceStateType,
    isLocked: boolean,
    manager: string,
    dev?: IRealDeviceData | null,
    isAcknowledged: boolean,
    connection?: IRealDeviceConnectionData | null,
    connectionType: DeviceConnectionType,
    connectionParameter: IConnectionParameter,
    createdAt?: Date,
    updatedAt?: Date,
    deletedAt?: Date,
    customParameter: IParameter[],
    deviceParameter: IParameter[];
}
