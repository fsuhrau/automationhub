import { DeviceConnectionType } from './device.connection.type.enum';

export default interface IConnectionParameter {
    connectionType: DeviceConnectionType,
    ip: string,
    port: number,
}