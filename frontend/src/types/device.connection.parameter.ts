import { DeviceConnectionType } from './device.connection.type.enum';

export default interface IConnectionParameter {
    ConnectionType: DeviceConnectionType,
    IP: string,
    Port: number,
}